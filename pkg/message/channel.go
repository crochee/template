package message

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/streadway/amqp"
	"go.uber.org/multierr"
)

type clientOption struct {
	attempts int
	interval time.Duration
	config   *amqp.Config
	uri      string
}

type ChannelOption func(*clientOption)

func WithAttempt(attempts int) ChannelOption {
	return func(c *clientOption) {
		c.attempts = attempts
	}
}

func WithInterval(interval time.Duration) ChannelOption {
	return func(c *clientOption) {
		c.interval = interval
	}
}

func WithURI(uri string) ChannelOption {
	return func(c *clientOption) {
		c.uri = uri
	}
}

func WithConfig(config *amqp.Config) ChannelOption {
	return func(c *clientOption) {
		c.config = config
	}
}

// Channel is a channel interface to make testing possible.
// It is highly recommended to use *amqp.Channel as the interface implementation.
type Channel interface {
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWail bool, args amqp.Table) (<-chan amqp.Delivery, error)
	DeclareAndBind(exchange, kind, queue, key string, args ...map[string]interface{}) error
}

func NewRabbitmqChannel(opts ...ChannelOption) (Channel, error) {
	o := &clientOption{
		uri: "amqp://guest:guest@localhost:5672/",
	}
	for _, opt := range opts {
		opt(o)
	}

	r := &rabbitmqChannel{
		clientOption: clientOption{
			attempts: o.attempts,
			interval: o.interval,
			config:   o.config,
			uri:      o.uri,
		},
	}
	if err := r.connect(); err != nil {
		return nil, err
	}
	runtime.SetFinalizer(r, func(w *rabbitmqChannel) {
		if err := multierr.Append(w.channel.Close(), w.conn.Close()); err != nil {
			_, _ = os.Stderr.WriteString(err.Error())
		}
	})
	return r, nil
}

type rabbitmqChannel struct {
	clientOption
	connected uint32
	conn      *amqp.Connection
	channel   *amqp.Channel
}

func (r *rabbitmqChannel) connect() error {
	var err error
	if r.config == nil {
		r.conn, err = amqp.Dial(r.uri)
	} else {
		r.conn, err = amqp.DialConfig(r.uri, *r.config)
	}
	if err != nil {
		return err
	}
	if r.conn.IsClosed() {
		return errors.New("rabbitmq connection is closed")
	}
	atomic.AddUint32(&r.connected, 1)
	if r.channel, err = r.conn.Channel(); err != nil {
		return fmt.Errorf("cann't open channel,%w", err)
	}
	return nil
}

func (r *rabbitmqChannel) retry() error {
	if atomic.LoadUint32(&r.connected) == 1 && r.conn.IsClosed() && r.attempts > 0 {
		timer := time.NewTicker(r.interval)
		var (
			tempAttempts int
			err          error
		)
		for range timer.C {
			shouldRetry := tempAttempts < r.attempts
			if !shouldRetry {
				break
			}
			if retryErr := r.connect(); retryErr != nil {
				err = multierr.Append(err, fmt.Errorf("%d try,%w", tempAttempts+1, retryErr))
			} else {
				shouldRetry = false
			}
			if !shouldRetry {
				break
			}
			tempAttempts++
		}
		timer.Stop()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *rabbitmqChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	if err := r.retry(); err != nil {
		return err
	}
	return r.channel.Publish(exchange, key, mandatory, immediate, msg)
}

func (r *rabbitmqChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWail bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	if err := r.retry(); err != nil {
		return nil, err
	}
	return r.channel.Consume(
		queue,
		// 用来区分多个消费者
		consumer,
		// 是否自动应答(自动应答确认消息，这里设置为否，在下面手动应答确认)
		autoAck,
		// 是否具有排他性
		exclusive,
		// 如果设置为true，表示不能将同一个connection中发送的消息
		// 传递给同一个connection的消费者
		noLocal,
		// 是否为阻塞
		noWail,
		args,
	)
}

func (r *rabbitmqChannel) DeclareAndBind(exchange, kind, queue, key string, args ...map[string]interface{}) error {
	if kind != "direct" && kind != "fanout" && kind != "topic" && kind != "headers" {
		return fmt.Errorf("invalid kind %s", kind)
	}
	if exchange == "" {
		return fmt.Errorf("invalid input %s", exchange)
	}
	if queue == "" {
		return fmt.Errorf("invalid input %s", queue)
	}
	if key == "" {
		return fmt.Errorf("invalid input %s", key)
	}
	var exchangeArg, queueArg, bindArg map[string]interface{}
	for index := range args {
		switch index {
		case 0:
			exchangeArg = args[index]
		case 1:
			queueArg = args[index]
		case 2:
			bindArg = args[index]
		default:
		}
	}
	if err := r.channel.ExchangeDeclare(
		exchange,
		kind,
		true,  // duration (note: is durable)
		false, // auto-delete
		false, // internal
		false, // nowait
		exchangeArg); err != nil {
		return err
	}
	if _, err := r.channel.QueueDeclare(queue,
		// 控制队列是否为持久的，当mq重启的时候不会丢失队列
		true,
		// 是否为自动删除
		false,
		// 是否具有排他性
		false,
		// 是否阻塞
		false,
		// 额外属性
		queueArg,
	); err != nil {
		return err
	}
	return r.channel.QueueBind(queue, key, exchange, false, bindArg)
}
