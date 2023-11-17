package async

import (
	"context"
	"os"

	jsoniter "github.com/json-iterator/go"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"

	"template/pkg/conc/pool"
	"template/pkg/logger/gormx"
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
func NewTaskConsumer(opts ...Option) *taskConsumer {
	o := &option{
		manager:   NewManager(),
		marshal:   DefaultMarshal{},
		handler:   jsoniter.ConfigCompatibleWithStandardLibrary,
		validator: validator.NewValidator(),
		autoAck:   true,
		form:      gormx.Nop,
	}

	for _, opt := range opts {
		opt(o)
	}
	return &taskConsumer{
		pool:      pool.New(),
		manager:   o.manager,
		marshal:   o.marshal,
		handler:   o.handler,
		validator: o.validator,
		autoAck:   o.autoAck,
		from:      o.form,
	}
}

type taskConsumer struct {
	pool      *pool.Pool         // goroutine safe run pool
	manager   ManagerTaskHandler // manager executor how to run
	marshal   MarshalAPI         // mq  assemble request or response
	handler   jsoniter.API
	validator validator.Validator
	autoAck   bool
	from      func(context.Context) gormx.Logger
}

// Register registers a TaskHandler with name
func (t *taskConsumer) Register(handlers ...TaskHandler) {
	t.manager.Register(handlers...)
}

// Unregister unregisters a TaskHandler with name
func (t *taskConsumer) Unregister(names ...string) {
	t.manager.Unregister(names...)
}

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
	tagInfix += os.Args[0]
	tagSuffix := "." + uuid.NewV4().String()
	if len(tagPrefix)+len(tagInfix)+len(tagSuffix) > consumerTagLengthMax {
		tagInfix = tagInfix[:consumerTagLengthMax-len(tagPrefix)-len(tagSuffix)]
	}
	return tagPrefix + tagInfix + tagSuffix
}

// Subscribe consume message form Channel with queueName
func (t *taskConsumer) Subscribe(ctx context.Context, channel Channel, queueName string) error {
	var err error
	p := t.pool.WithContext(ctx).WithCancelOnError()
	p.Go(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
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
				return err
			}
			t.handleMessage(ctx, deliveries)
		}
	})
	return p.Wait()
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
			p := t.pool.WithContext(ctx)
			p.Go(func(ctx context.Context) error {
				if t.autoAck {
					if err := t.autoHandle(ctx, v); err != nil {
						t.from(ctx).Errorf("err:%+v", err)
					}
					return nil
				}
				if err := t.manualHandle(ctx, v); err != nil {
					t.from(ctx).Errorf("err:%+v", err)
				}
				return nil
			})
		}
	}
}

func (t *taskConsumer) manualHandle(ctx context.Context, d amqp.Delivery) error {
	msgStruct, err := t.marshal.Unmarshal(&d)
	if err != nil {
		t.from(ctx).Errorf("err:%+v", err)
		// 当requeue为true时，将该消息排队，以在另一个通道上传递给使用者。
		// 当requeue为false或服务器无法将该消息排队时，它将被丢弃。
		if err = d.Reject(false); err != nil {
			return err
		}
		return nil
	}
	logContext := t.from(ctx)
	logContext.Debugf("get body %s,uuid:%s", msgStruct.Payload, msgStruct.UUID)
	param := Get()
	if err = t.handler.Unmarshal(msgStruct.Payload, param); err != nil {
		logContext.Errorf("err:%+v", err)
		// 当requeue为true时，将该消息排队，以在另一个通道上传递给使用者。
		// 当requeue为false或服务器无法将该消息排队时，它将被丢弃。
		if err = d.Reject(false); err != nil {
			return err
		}
		return nil
	}
	if err = t.validator.ValidateStruct(param); err != nil {
		logContext.Errorf("err:%+v", err)
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
		logContext.Errorf("err:%+v", err)
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

	t.from(ctx).Debugf("get body consume:%s,uuid:%s", msgStruct.Payload, msgStruct.UUID)
	param := Get()
	if err = t.handler.Unmarshal(msgStruct.Payload, param); err != nil {
		return err
	}
	if err = t.validator.ValidateStruct(param); err != nil {
		return err
	}
	return t.manager.Run(ctx, param)
}
