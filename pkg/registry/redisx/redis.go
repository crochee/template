package redisx

import (
	"context"
	"time"

	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"template/pkg/logger"
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

func NewRedisRegistry(ctx context.Context, client *redis.ClusterClient, opts ...Option) registry.Registry {
	newCtx, cancel := context.WithCancel(
		logger.With(context.Background(), logger.From(ctx)))
	r := &redisRegistry{
		client:  client,
		subStop: make(chan struct{}),
		ctx:     newCtx,
		cancel:  cancel,
		option: option{
			expireTime:   time.Minute,
			tickerTime:   30 * time.Second,
			encoder:      DefaultEncoder{},
			keyGenerator: DefaultKeyGenerator{},
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

	ctx    context.Context
	cancel context.CancelFunc
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
	r.cancel()
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
				logger.From(ctx).Warn("HGetAll key",
					zap.String("key", key),
					zap.Error(err))
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
			logger.From(ctx).Warn("Decode value failed",
				zap.String("value", value),
				zap.Error(err))
			return
		}
		var refreshData string
		if refreshData, err = r.encoder.Encode(
			action,
			info,
			time.Now().Add(r.expireTime).Unix(),
		); err != nil {
			logger.From(ctx).Warn("Encode to refresh failed",
				zap.Any("info", info),
				zap.String("action", string(action)),
				zap.Int64("expireAt", expireAt),
				zap.Error(err))
			return
		}
		if _, err = r.client.HSet(r.ctx, key, uuidStr, refreshData).Result(); err != nil {
			logger.From(ctx).Warn("HSet value failed",
				zap.String("value", refreshData),
				zap.Error(err))
		}
		return
	}
	action, info, expireAt, err := r.encoder.Decode(value)
	if err != nil {
		logger.From(ctx).Warn("Decode value failed",
			zap.String("value", value),
			zap.Error(err))
		return
	}
	// 非正常断链，其他正常节点进行数据清理和通知
	if time.Now().Unix() > expireAt {
		// 唯一执行
		var lock *redislock.Lock
		if lock, err = redislock.Obtain(ctx, r.client, uuidStr, time.Minute, nil); err != nil {
			logger.From(ctx).Warn("redis get lock failed",
				zap.Error(err), zap.String("id", uuidStr))
			return
		}
		defer func() {
			if err = lock.Release(ctx); err != nil {
				logger.From(ctx).Warn("redis Release lock failed",
					zap.Error(err), zap.String("id", uuidStr))
			}
		}()
		if _, err = r.client.HDel(r.ctx, key, uuidStr).Result(); err != nil {
			logger.From(ctx).Warn("HSet value failed",
				zap.String("key", key),
				zap.String("uuid", uuidStr),
				zap.Error(err))
			return
		}
		var data string
		if data, err = r.encoder.Encode(
			Deregister,
			info,
			expireAt,
		); err != nil {
			logger.From(ctx).Warn("Encode to refresh failed",
				zap.Any("info", info),
				zap.String("action", string(action)),
				zap.Int64("expireAt", expireAt),
				zap.Error(err))
			return
		}
		if _, err = r.client.Publish(r.ctx, key, data).Result(); err != nil {
			logger.From(ctx).Warn("Publish value failed",
				zap.String("key", key),
				zap.String("value", data),
				zap.Error(err))
		}

	}
}
