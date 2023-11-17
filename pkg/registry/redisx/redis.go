package redisx

import (
	"context"
	"time"

	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"

	"template/pkg/logger/gormx"
	"template/pkg/registry"
)

const (
	Register   Action = "register"
	Deregister Action = "deregister"
)

type Action string

type option struct {
	expireTime time.Duration
	tickerTime time.Duration

	eventFlow    EventFlow
	encoder      Encoder
	keyGenerator KeyGenerator

	from func(context.Context) gormx.Logger
}

type Option func(*option)

func WithExpireTime(expireTime time.Duration) Option {
	return func(o *option) {
		o.expireTime = expireTime
	}
}

func WithTickerTime(tickerTime time.Duration) Option {
	return func(o *option) {
		o.tickerTime = tickerTime
	}
}

func WithEventFlow(eventFlow EventFlow) Option {
	return func(o *option) {
		o.eventFlow = eventFlow
	}
}

func WithEncoder(encoder Encoder) Option {
	return func(o *option) {
		o.encoder = encoder
	}
}

func WithKeyGenerator(keyGenerator KeyGenerator) Option {
	return func(o *option) {
		o.keyGenerator = keyGenerator
	}
}

func WithLogFrom(from func(context.Context) gormx.Logger) Option {
	return func(o *option) {
		o.from = from
	}
}

func NewRedisRegistry(ctx context.Context, client *redis.ClusterClient, opts ...Option) registry.Registry {
	r := &redisRegistry{
		client:  client,
		subStop: make(chan struct{}),
		ctx:     ctx,
		option: option{
			expireTime:   time.Minute,
			tickerTime:   30 * time.Second,
			encoder:      DefaultEncoder{},
			keyGenerator: DefaultKeyGenerator{},
			from:         gormx.Nop,
		},
	}
	for _, opt := range opts {
		opt(&r.option)
	}
	return r
}

type redisRegistry struct {
	client  *redis.ClusterClient
	subStop chan struct{}

	ctx context.Context
	option
}

func (r *redisRegistry) Register(info *registry.Info) error {
	if info.ServiceName == "" {
		return errors.New("ServiceName can not be empty")
	}
	if info.Addr == "" {
		return errors.New("Addr can not be empty")
	}
	key := r.keyGenerator.Create(info)
	value, err := r.encoder.Encode(Register, info, time.Now().Add(r.expireTime).Unix())
	if err != nil {
		return err
	}
	r.client.HSet(r.ctx, key, info.UUID, value)
	r.client.Publish(r.ctx, key, value)
	go r.subscribe(r.ctx, key)
	go r.keepAlive(r.ctx, key, info.UUID)
	return nil
}

func (r *redisRegistry) Deregister(info *registry.Info) error {
	if info.ServiceName == "" {
		return errors.New("ServiceName can not be empty")
	}
	if info.Addr == "" {
		return errors.New("Addr can not be empty")
	}
	key := r.keyGenerator.Create(info)
	value, err := r.encoder.Encode(Deregister, info, time.Now().Unix())
	if err != nil {
		return err
	}
	close(r.subStop)
	r.client.HDel(r.ctx, key, info.UUID)
	r.client.Publish(r.ctx, key, value)
	return nil
}

func (r *redisRegistry) subscribe(ctx context.Context, key string) {
	if r.eventFlow == nil {
		return
	}
	sub := r.client.Subscribe(ctx, key)
	defer sub.Close()
	select {
	case <-ctx.Done():
		return
	default:
		ch := sub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case <-r.subStop:
				return
			case msg := <-ch:
				r.eventFlow.Change(ctx, msg)
			}
		}
	}
}

func (r *redisRegistry) keepAlive(ctx context.Context, key, uuid string) {
	ticker := time.NewTicker(r.tickerTime)
	for {
		select {
		case <-ticker.C:
			m, err := r.client.HGetAll(ctx, key).Result()
			if err != nil {
				r.from(ctx).Warnf("HGetAll key,key:%s,err:%+v", key, err)
				return
			}
			for uuidStr, value := range m {
				r.compensateAndFix(ctx, key, uuid, uuidStr, value)
			}
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (r *redisRegistry) compensateAndFix(ctx context.Context, key, thisUuid, uuidStr, value string) {
	if thisUuid == uuidStr {
		action, info, expireAt, err := r.encoder.Decode(value)
		if err != nil {
			r.from(ctx).Warnf("Decode value failed,value:%s,err:%+v", value, err)
			return
		}
		var refreshData string
		if refreshData, err = r.encoder.Encode(
			action,
			info,
			time.Now().Add(r.expireTime).Unix(),
		); err != nil {
			r.from(ctx).Warnf("Encode to refresh failed,info:%+v,action:%s,expireAt:%d,err:%+v",
				info, action, expireAt, err)
			return
		}
		if _, err = r.client.HSet(r.ctx, key, uuidStr, refreshData).Result(); err != nil {
			r.from(ctx).Warnf("HSet value failed,value:%s,err:%+v", refreshData, err)
		}
		return
	}
	action, info, expireAt, err := r.encoder.Decode(value)
	if err != nil {
		r.from(ctx).Warnf("Decode value failed,value:%s,err:%+v", value, err)
		return
	}
	// 非正常断链，其他正常节点进行数据清理和通知
	if time.Now().Unix() > expireAt {
		// 唯一执行
		var lock *redislock.Lock
		if lock, err = redislock.Obtain(ctx, r.client, uuidStr, time.Minute, nil); err != nil {
			r.from(ctx).Warnf("redis get lock failed,id:%s,err:%+v", uuidStr, err)
			return
		}
		defer func() {
			if err = lock.Release(ctx); err != nil {
				r.from(ctx).Warnf("redis Release lock failed,id:%s,err:%+v", uuidStr, err)
			}
		}()
		if _, err = r.client.HDel(r.ctx, key, uuidStr).Result(); err != nil {
			r.from(ctx).Warnf("HSet value failed,key:%s,id:%s,err:%+v", key, uuidStr, err)
			return
		}
		var data string
		if data, err = r.encoder.Encode(
			Deregister,
			info,
			expireAt,
		); err != nil {
			r.from(ctx).Warnf("Encode to refresh failed,info:%+v,action:%s,expireAt:%d,err:%+v",
				info, action, expireAt, err)
			return
		}
		if _, err = r.client.Publish(r.ctx, key, data).Result(); err != nil {
			r.from(ctx).Warnf("Publish value failed,key:%s,data:%s,err:%+v", key, data, err)
		}
	}
}
