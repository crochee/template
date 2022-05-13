package message

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/json-iterator/go"
	"github.com/streadway/amqp"

	"go-template/pkg/validator"
)

type Producer interface {
	Publish(ctx context.Context, channel Channel, exchange, routingKey string, param interface{}) error
	io.Closer
}

type TaskProducer struct {
	marshal   MarshalAPI // mq  assemble request or response
	handler   jsoniter.API
	validator validator.Validator
	wg        sync.WaitGroup
}

// NewTaskProducer gets Producer
func NewTaskProducer(opts ...Option) Producer {
	o := &option{
		marshal:   DefaultMarshal{},
		handler:   jsoniter.ConfigCompatibleWithStandardLibrary,
		validator: validator.NewValidator(),
	}

	for _, opt := range opts {
		opt(o)
	}
	return &TaskProducer{
		marshal:   o.marshal,
		handler:   o.handler,
		validator: o.validator,
	}
}

func (t *TaskProducer) Publish(ctx context.Context, channel Channel, exchange, routingKey string, param interface{}) error {
	t.wg.Add(1)
	defer t.wg.Done()
	if err := t.validator.ValidateStruct(param); err != nil {
		return err
	}
	data, err := t.handler.Marshal(param)
	if err != nil {
		return err
	}
	uuid := watermill.NewUUID()
	var amqpMsg amqp.Publishing
	if amqpMsg, err = t.marshal.Marshal(message.NewMessage(uuid, data)); err != nil {
		return fmt.Errorf("cann't marshal message,%w", err)
	}
	// 发送消息到队列中
	return channel.Publish(
		exchange,
		routingKey,
		// 如果为true，根据exchange类型和routekey类型，如果无法找到符合条件的队列，name会把发送的信息返回给发送者
		false,
		// 如果为true，当exchange发送到消息队列后发现队列上没有绑定的消费者,则会将消息返还给发送者
		false,
		// 发送信息
		amqpMsg,
	)
}

func (t *TaskProducer) Close() error {
	t.wg.Wait()
	return nil
}
