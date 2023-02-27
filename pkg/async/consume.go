package async

import (
	"context"
	"runtime/debug"
	"strconv"
	"sync/atomic"

	"github.com/json-iterator/go"
	"github.com/streadway/amqp"
	"go.uber.org/zap"

	"template/pkg/logger"
	"template/pkg/routine"
	"template/pkg/validator"
)

// Consumer async impl
type Consumer interface {
	Subscribe(ctx context.Context, channel Channel, queueName string) error
}

type ConsumerRegistrar interface {
	Register(handlers ...TaskHandler)
	Unregister(names ...string)
	Consumer
}

// NewTaskConsumer gets Consumer
func NewTaskConsumer(ctx context.Context, opts ...Option) *taskConsumer {
	o := &option{
		manager:   NewManager(),
		marshal:   DefaultMarshal{},
		handler:   jsoniter.ConfigCompatibleWithStandardLibrary,
		validator: validator.NewValidator(),
		autoAck:   true,
	}

	for _, opt := range opts {
		opt(o)
	}
	return &taskConsumer{
		pool: routine.NewPool(ctx, routine.Recover(func(ctx context.Context, i interface{}) {
			logger.From(ctx).Sugar().Errorf("err:%v\n%s", i, debug.Stack())
		})),
		manager:   o.manager,
		marshal:   o.marshal,
		handler:   o.handler,
		validator: o.validator,
		autoAck:   o.autoAck,
	}
}

type taskConsumer struct {
	pool      *routine.Pool      // goroutine safe run pool
	manager   ManagerTaskHandler // manager executor how to run
	marshal   MarshalAPI         // mq  assemble request or response
	handler   jsoniter.API
	validator validator.Validator
	autoAck   bool
}

// Register registers a TaskHandler with name
func (t *taskConsumer) Register(handlers ...TaskHandler) {
	t.manager.Register(handlers...)
}

// Unregister unregisters a TaskHandler with name
func (t *taskConsumer) Unregister(names ...string) {
	t.manager.Unregister(names...)
}

var consumerSeq uint64

const consumerTagLengthMax = 0xFF // see writeShortstr

func uniqueConsumerTag(name string) string {
	return commandNameBasedUniqueConsumerTag(name)
}

func commandNameBasedUniqueConsumerTag(name string) string {
	tagPrefix := "ctag."
	tagInfix := name
	if tagInfix == "" {
		tagInfix = "amqp"
	}
	tagSuffix := "." + strconv.FormatUint(atomic.AddUint64(&consumerSeq, 1), 10)

	if len(tagPrefix)+len(tagInfix)+len(tagSuffix) > consumerTagLengthMax {
		tagInfix = tagInfix[:consumerTagLengthMax-len(tagPrefix)-len(tagSuffix)]
	}
	return tagPrefix + tagInfix + tagSuffix
}

// Subscribe consume message form Channel with queueName
func (t *taskConsumer) Subscribe(ctx context.Context, channel Channel, queueName string) error {
	var err error
	t.pool.Go(ctx, func(ctx context.Context) {
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
				uniqueConsumerTag(queueName),
				// 是否自动应答(自动应答确认消息，这里设置为否，在下面手动应答确认)
				t.autoAck,
				// 是否具有排他性
				false,
				// 如果设置为true，表示不能将同一个connection中发送的消息
				// 传递给同一个connection的消费者
				false,
				// 是否为阻塞
				false,
				nil,
			); err != nil {
				logger.From(ctx).Error("Consume failed", zap.Error(err))
				return
			}
			t.handleMessage(ctx, deliveries)
		}
	})
	t.pool.Wait()
	return err
}

func (t *taskConsumer) handleMessage(ctx context.Context, deliveries <-chan amqp.Delivery) {
	for {
		select {
		case <-ctx.Done():
			return
		case v, ok := <-deliveries:
			if !ok {
				return
			}
			t.pool.Go(ctx, func(ctx context.Context) {
				if t.autoAck {
					if err := t.autoHandle(ctx, v); err != nil {
						logger.From(ctx).Error("", zap.Error(err))
					}
					return
				}
				if err := t.manualHandle(ctx, v); err != nil {
					logger.From(ctx).Error("", zap.Error(err))
				}
			})
		}
	}
}

func (t *taskConsumer) manualHandle(ctx context.Context, d amqp.Delivery) error {
	msgStruct, err := t.marshal.Unmarshal(&d)
	if err != nil {
		logger.From(ctx).Error("", zap.Error(err))
		// 当requeue为true时，将该消息排队，以在另一个通道上传递给使用者。
		// 当requeue为false或服务器无法将该消息排队时，它将被丢弃。
		if err = d.Reject(false); err != nil {
			return err
		}
		return nil
	}
	logContext := logger.From(ctx)
	logContext.Debug("get body",
		zap.ByteString("payload", msgStruct.Payload),
		zap.String("uuid", msgStruct.UUID),
	)
	param := Get()
	if err = t.handler.Unmarshal(msgStruct.Payload, param); err != nil {
		logContext.Error("", zap.Error(err))
		// 当requeue为true时，将该消息排队，以在另一个通道上传递给使用者。
		// 当requeue为false或服务器无法将该消息排队时，它将被丢弃。
		if err = d.Reject(false); err != nil {
			return err
		}
		return nil
	}
	if err = t.validator.ValidateStruct(param); err != nil {
		logContext.Error("", zap.Error(err))
		// 当requeue为true时，将该消息排队，以在另一个通道上传递给使用者。
		// 当requeue为false或服务器无法将该消息排队时，它将被丢弃。
		if err = d.Reject(false); err != nil {
			return err
		}
		return nil
	}
	err = t.manager.Run(ctx, param)
	Put(param)
	if err != nil {
		logContext.Error("", zap.Error(err))
		// 当requeue为true时，将该消息排队，以在另一个通道上传递给使用者。
		// 当requeue为false或服务器无法将该消息排队时，它将被丢弃。
		if err = d.Reject(true); err != nil {
			return err
		}
		return nil
	}
	// 手动确认收到本条消息, true表示回复当前信道所有未回复的ack，用于批量确认。
	// false表示回复当前条目
	return d.Ack(false)
}

func (t *taskConsumer) autoHandle(ctx context.Context, d amqp.Delivery) error {
	msgStruct, err := t.marshal.Unmarshal(&d)
	if err != nil {
		return err
	}

	logger.From(ctx).Debug("get body",
		zap.Any("consume body:%s", msgStruct.Payload),
		zap.String("uuid", msgStruct.UUID),
	)
	param := Get()
	if err = t.handler.Unmarshal(msgStruct.Payload, param); err != nil {
		return err
	}
	if err = t.validator.ValidateStruct(param); err != nil {
		return err
	}
	return t.manager.Run(ctx, param)
}
