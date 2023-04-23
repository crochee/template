package async

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/streadway/amqp"
	"go.uber.org/multierr"

	"template/pkg/utils/v"
)

type clientOption struct {
	attempts int
	interval time.Duration
	config   *amqp.Config
	uri      string
	tx       bool
	qos      *QosOption
}

type QosOption struct {
	PrefetchCount int
	PrefetchSize  int
	Global        bool
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

func WithTx(tx bool) ChannelOption {
	return func(c *clientOption) {
		c.tx = tx
	}
}

func WithQos(qos QosOption) ChannelOption {
	return func(c *clientOption) {
		c.qos = &qos
	}
}

//go:generate mockgen -source=./channel.go -destination=./channel_mock.go -package=async

// Channel is a channel interface to make testing possible.
// It is highly recommended to use *amqp.Channel as the interface implementation.
type Channel interface {
	Publish(exchange, key string, mandatory, immediate bool, msg ...amqp.Publishing) error
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWail bool, args amqp.Table) (<-chan amqp.Delivery, error)
	DeclareAndBind(exchange, kind, queue, key string, args ...map[string]interface{}) error
}

func NewRabbitmqChannel(opts ...ChannelOption) (Channel, error) {
	o := &clientOption{
		uri:      "amqp://guest:guest@localhost:5672/",
		interval: time.Millisecond,
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
	return r, nil
}

type rabbitmqChannel struct {
	clientOption
	conn    *amqp.Connection
	channel *amqp.Channel
	mu      sync.Mutex
}

func (r *rabbitmqChannel) Close() error {
	return multierr.Append(r.channel.Close(), r.conn.Close())
}

func (r *rabbitmqChannel) connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	//检测链接是否正常
	if r.conn != nil && !r.conn.IsClosed() {
		return nil
	}
	var err error
	if r.config == nil {
		r.conn, err = amqp.Dial(r.uri)
	} else {
		r.conn, err = amqp.DialConfig(r.uri, *r.config)
	}
	if err != nil {
		return err
	}
	if r.channel, err = r.conn.Channel(); err != nil {
		return fmt.Errorf("can't open channel,%w", err)
	}
	return nil
}

func (r *rabbitmqChannel) retry() error {
	if r.attempts > 0 && r.conn.IsClosed() {
		var (
			tempAttempts int
			err          error
		)
		backOff := r.newBackOff() // 退避算法 保证时间间隔为指数级增长
		currentInterval := 0 * time.Millisecond
		timer := time.NewTimer(currentInterval)
		for range timer.C {
			shouldRetry := tempAttempts < r.attempts
			if !shouldRetry {
				timer.Stop()
				return err
			}
			retryErr := r.connect()
			if retryErr == nil {
				timer.Stop()
				return nil
			}
			var permanent *backoff.PermanentError
			if errors.As(retryErr, &permanent) {
				err = multierr.Append(err, fmt.Errorf("%w %d try", permanent.Err, tempAttempts+1))
				shouldRetry = false
			} else {
				err = multierr.Append(err, fmt.Errorf("%w %d try", retryErr, tempAttempts+1))
			}
			if !shouldRetry {
				timer.Stop()
				return err
			}
			// 计算下一次
			currentInterval = backOff.NextBackOff()
			tempAttempts++
			// 定时器重置
			timer.Reset(currentInterval)
		}
	}
	return nil
}

func (r *rabbitmqChannel) newBackOff() backoff.BackOff {
	if r.attempts < 2 || r.interval <= 0 {
		return &backoff.ZeroBackOff{}
	}

	b := backoff.NewExponentialBackOff()
	b.InitialInterval = r.interval

	// calculate the multiplier for the given number of attempts
	// so that applying the multiplier for the given number of attempts will not exceed 2 times the initial interval
	// it allows to control the progression along the attempts
	b.Multiplier = math.Pow(v.Binary, 1/float64(r.attempts-1))

	// according to docs, b.Reset() must be called before using
	b.Reset()
	return b
}

func (r *rabbitmqChannel) Publish(exchange, key string, mandatory, immediate bool, msg ...amqp.Publishing) error {
	if err := r.retry(); err != nil {
		return err
	}
	if r.tx {
		return r.txPublish(exchange, key, mandatory, immediate, msg...)
	}
	for _, m := range msg {
		if err := r.channel.Publish(exchange, key, mandatory, immediate, m); err != nil {
			return err
		}
	}
	return nil
}

func (r *rabbitmqChannel) txPublish(exchange, key string, mandatory, immediate bool, msg ...amqp.Publishing) (err error) {
	if err = r.channel.Tx(); err != nil {
		err = fmt.Errorf("can't start transaction,%w", err)
		return
	}
	defer func() {
		if err != nil {
			err = multierr.Append(err, r.channel.TxRollback())
		}
	}()
	for _, m := range msg {
		if err = r.channel.Publish(exchange, key, mandatory, immediate, m); err != nil {
			return
		}
	}
	err = r.channel.TxCommit()
	return
}

func (r *rabbitmqChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWail bool,
	args amqp.Table) (<-chan amqp.Delivery, error) {
	if err := r.retry(); err != nil {
		return nil, err
	}
	if r.qos != nil {
		if err := r.channel.Qos(r.qos.PrefetchCount, r.qos.PrefetchSize, r.qos.Global); err != nil {
			return nil, fmt.Errorf("set qos failed,%w", err)
		}
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

type NoopChannel struct{}

func (NoopChannel) Publish(exchange, key string, mandatory, immediate bool, msg ...amqp.Publishing) error {
	return nil
}
func (NoopChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWail bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	return nil, nil
}
func (NoopChannel) DeclareAndBind(exchange, kind, queue, key string, args ...map[string]interface{}) error {
	return nil
}
