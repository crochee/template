// Package msg
package msg

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/json-iterator/go"
	"github.com/streadway/amqp"
	"go.uber.org/zap"

	"template/pkg/async"
	"template/pkg/logger"
	"template/pkg/routine"
	"template/pkg/validator"
)

func NewReader(opts ...func(*ReaderOption)) *reader {
	r := &reader{
		ReaderOption: ReaderOption{
			LevelFunc: func() map[string]uint8 {
				return map[string]uint8{}
			},
			MetadataPool: NewMetadataPool(),
			Queue:        NewQueue(1024 * 10),
			JSONHandler:  jsoniter.ConfigCompatibleWithStandardLibrary,
			Validator:    validator.NewValidator(),
			Plugins:      make([]Plugin, 0, 2),
			Handles:      make([]func(*Metadata) error, 0, 2),
			Marshal:      async.DefaultMarshal{},
		},
	}
	for _, o := range opts {
		o(&r.ReaderOption)
	}
	return r
}

type LevelFunc func() map[string]uint8

type ReaderOption struct {
	LevelFunc    LevelFunc
	MetadataPool MetadataPool
	Queue        Queue
	JSONHandler  jsoniter.API
	Validator    validator.Validator
	Plugins      []Plugin
	Handles      []func(*Metadata) error
	Marshal      async.MarshalAPI
}

type reader struct {
	queuePool QueuePool
	pool      *routine.Pool // goroutine safe run pool
	ReaderOption
}

func (r *reader) Subscribe(ctx context.Context, channel async.Channel, queueName string) error {
	// 初始化
	r.pool = routine.NewPool(ctx, routine.Recover(func(ctx context.Context, i interface{}) {
		logger.From(ctx).Sugar().Errorf("err:%v\n%s", i, debug.Stack())
	}))
	if r.queuePool == nil {
		r.queuePool = NewQueuePool(ctx, 30*time.Second, 90*time.Second, func() Queue {
			return NewPriorityQueue(r.LevelFunc)
		})
	}
	r.queuePool.SendOn(func(metadata []*Metadata) {
		for _, data := range metadata {
			r.Queue.Write(data)
		}
	})
	// 业务处理
	var err error
	r.pool.Go(ctx, func(ctx context.Context) {
		// 失败次数
		var failedCount = 5
		for {
			select {
			case <-ctx.Done():
				err = ctx.Err()
				return
			default:
			}
			var deliveries <-chan amqp.Delivery
			if deliveries, err = channel.Consume(
				queueName,
				// 用来区分多个消费者
				"dcs.consumer."+queueName,
				// 是否自动应答(自动应答确认消息，这里设置为否，在下面手动应答确认)
				true,
				// 是否具有排他性
				false,
				// 如果设置为true，表示不能将同一个connection中发送的消息
				// 传递给同一个connection的消费者
				false,
				// 是否为阻塞
				false,
				nil,
			); err != nil {
				logger.From(ctx).Error("", zap.Error(err))
				if failedCount < 1 {
					return
				}
				failedCount--
				continue
			}
			r.handleMessage(ctx, deliveries)
		}
	})
	r.pool.Go(ctx, r.listen)
	r.pool.Wait()
	return err
}

func (r *reader) handleMessage(ctx context.Context, deliveries <-chan amqp.Delivery) {
	for {
		select {
		case <-ctx.Done():
			return
		case v, ok := <-deliveries:
			if !ok {
				return
			}
			r.pool.Go(ctx, func(ctx context.Context) {
				if err := r.autoHandle(ctx, v); err != nil {
					logger.From(ctx).Error("", zap.Error(err))
				}
			})
		}
	}
}

func (r *reader) autoHandle(ctx context.Context, d amqp.Delivery) error {
	msg, err := r.Marshal.Unmarshal(&d)
	if err != nil {
		return err
	}
	meta := r.MetadataPool.Get()
	if err = r.JSONHandler.Unmarshal(msg.Payload, meta); err != nil {
		r.MetadataPool.Put(meta)
		return err
	}
	if err = r.Validator.ValidateStruct(meta); err != nil {
		r.MetadataPool.Put(meta)
		return err
	}
	r.pool.Go(ctx, func(ctx context.Context) {
		r.queuePool.Get(meta.TraceID).Write(meta)
	})
	return nil
}

func (r *reader) listen(ctx context.Context) {
	for {
		select {
		case data := <-r.Queue.Read():
			r.handleProcess(data)
		case <-ctx.Done():
			_ = r.Queue.Close()
			results := r.Queue.ListAndClear()
			r.handleProcess(results...)
			return
		}
	}
}

func (r *reader) handleProcess(metadataList ...*Metadata) {
	for _, metadata := range metadataList {
		for _, plugin := range r.Plugins { // 插件使用
			plugin.Use(metadata)
		}
		for _, handle := range r.Handles { // finish使用
			if err := handle(metadata); err != nil {
				break
			}
		}
		r.MetadataPool.Put(metadata)
	}
}
