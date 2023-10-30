// Package msg
package msg

import (
	"context"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/streadway/amqp"
	"go.uber.org/zap"

	"template/pkg/async"
	"template/pkg/conc/pool"
	"template/pkg/logger"
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
		pool: pool.New(),
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
	pool      *pool.Pool // goroutine safe run pool
	ReaderOption
}

func (r *reader) Subscribe(ctx context.Context, channel async.Channel, queueName string) error {
	// 初始化
	p := r.pool.WithContext(ctx).WithCancelOnError()
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
	p.Go(func(ctx context.Context) error {
		// 失败次数
		var failedCount = 5
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			deliveries, err := channel.Consume(
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
			)
			if err != nil {
				logger.From(ctx).Error("", zap.Error(err))
				if failedCount < 1 {
					return err
				}
				failedCount--
				continue
			}
			r.handleMessage(ctx, deliveries)
			return nil
		}
	})
	p.Go(r.listen)
	return p.Wait()
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
			r.pool.Go(func() {
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
	r.pool.Go(func() {
		r.queuePool.Get(meta.TraceID).Write(meta)
	})
	return nil
}

func (r *reader) listen(ctx context.Context) error {
	for {
		select {
		case data := <-r.Queue.Read():
			r.handleProcess(data)
		case <-ctx.Done():
			_ = r.Queue.Close()
			results := r.Queue.ListAndClear()
			r.handleProcess(results...)
			return ctx.Err()
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
