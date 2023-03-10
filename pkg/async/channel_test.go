package async

import (
	"context"
	"errors"
	"fmt"
	"log"
	"template/pkg/json"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/golang/mock/gomock"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
)

// mockChannel is a mock of Channel.
type mockChannel struct {
	deliveries chan amqp.Delivery
}

// Publish runs a test function f and sends resultant message to a channel.
func (ch *mockChannel) Publish(exchange, key string, mandatory, immediate bool, msg ...amqp.Publishing) error {
	for _, v := range msg {
		if err := ch.handle(v); err != nil {
			return err
		}
	}
	return nil
}

func (ch *mockChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWail bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	dev := make(chan amqp.Delivery)
	go func() {
		v := <-ch.deliveries
		dev <- v
	}()
	return dev, nil
}

func (ch *mockChannel) DeclareAndBind(exchange, kind, queue, key string, args ...map[string]interface{}) error {
	return nil
}

func (ch *mockChannel) handle(msg amqp.Publishing) error {
	ch.deliveries <- amqp.Delivery{
		Acknowledger:    mockAck{},
		Headers:         msg.Headers,
		ContentType:     msg.ContentType,
		ContentEncoding: msg.ContentEncoding,
		DeliveryMode:    msg.DeliveryMode,
		Priority:        msg.Priority,
		CorrelationId:   msg.CorrelationId,
		ReplyTo:         msg.ReplyTo,
		Expiration:      msg.Expiration,
		MessageId:       msg.MessageId,
		Timestamp:       msg.Timestamp,
		Type:            msg.Type,
		UserId:          msg.UserId,
		AppId:           msg.AppId,
		ConsumerTag:     "",
		MessageCount:    0,
		DeliveryTag:     0,
		Redelivered:     false,
		Exchange:        "",
		RoutingKey:      "",
		Body:            msg.Body,
	}
	return nil
}

type mockAck struct {
}

func (m mockAck) Ack(tag uint64, multiple bool) error {
	return nil
}

func (m mockAck) Nack(tag uint64, multiple bool, requeue bool) error {
	return nil
}

func (m mockAck) Reject(tag uint64, requeue bool) error {
	return nil
}

func TestInteract(t *testing.T) {
	c := &mockChannel{deliveries: make(chan amqp.Delivery, 10)}
	tp := NewTaskProducer()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	tc := NewTaskConsumer(ctx)
	tc.Register(test{})
	tc.Register(&multiTest{list: []TaskHandler{test{}, &test1{}}})
	if err := tp.Publish(ctx, c, "", "", &Param{
		TaskType: "test",
		Metadata: nil,
		Data:     nil,
	}); err != nil {
		t.Fatal(err)
	}
	if err := tc.Subscribe(ctx, c, ""); err != nil {
		t.Fatal(err)
	}
}

func TestProduce(t *testing.T) {
	cp, err := NewRabbitmqChannel(WithURI("amqp://admin:1234567@localhost:5672/"))
	if err != nil {
		t.Fatal(err)
	}
	tp := NewTaskProducer()
	for i := 0; i < 20000; i++ {
		if err = tp.Publish(context.Background(), cp, "dcs.api.async", "msg.dcs.woden", &Param{
			TaskType: "rudder",
			Data:     []byte(fmt.Sprintf("task1:%d", i)),
		}); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 20000; i++ {
		if err = tp.Publish(context.Background(), cp, "dcs.api.async", "msg.dcs.woden", &Param{
			TaskType: "rudder",
			Data:     []byte(fmt.Sprintf("task2:%d", i)),
		}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestConsume(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockChannel := make(chan amqp.Delivery, 10)
	go func() {
		data, err := json.Marshal(&Param{
			TaskType: "test1",
			Data:     []byte("sd"),
		})
		if err != nil {
			t.Fatal(err)
		}
		msg, err := (DefaultMarshal{}).Marshal(message.NewMessage(uuid.NewV4().String(), data))
		if err != nil {
			t.Fatal(err)
		}

		mockChannel <- amqp.Delivery{
			Acknowledger:    mockAck{},
			Headers:         msg.Headers,
			ContentType:     msg.ContentType,
			ContentEncoding: msg.ContentEncoding,
			DeliveryMode:    msg.DeliveryMode,
			Priority:        msg.Priority,
			CorrelationId:   msg.CorrelationId,
			ReplyTo:         msg.ReplyTo,
			Expiration:      msg.Expiration,
			MessageId:       msg.MessageId,
			Timestamp:       msg.Timestamp,
			Type:            msg.Type,
			UserId:          msg.UserId,
			AppId:           msg.AppId,
			ConsumerTag:     "",
			MessageCount:    0,
			DeliveryTag:     0,
			Redelivered:     false,
			Exchange:        "",
			RoutingKey:      "",
			Body:            msg.Body,
		}
		close(mockChannel)
	}()
	cc := NewMockChannel(ctl)
	cc.EXPECT().Consume(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any()).Return(mockChannel, nil).AnyTimes()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tc := NewTaskConsumer(context.Background())
	tc.Register(testError{})
	tc.Register(&test1{})
	tc.Register(&multiTest{list: []TaskHandler{test{}, &test1{}}})
	tc.Register(&rudder{})
	if err := tc.Subscribe(ctx, cc, "msg.dcs.woden"); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}

type rudder struct {
}

func (t rudder) Name() string {
	return "rudder"
}

func (r rudder) Run(ctx context.Context, param *Param) error {
	log.Printf("%s %v %s\n", param.TaskType, param.Metadata, param.Data)
	return nil
}
