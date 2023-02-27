package server

import (
	"context"
	"fmt"
	"time"

	_ "go.uber.org/automaxprocs"

	"template/pkg/async"
	"template/pkg/msg"
)

var gw *msg.Writer

// Update 更新全局参数
func Update(ctx context.Context, opts ...func(*msg.WriterOption)) {
	if gw == nil {
		gw = msg.NewWriter(ctx, opts...)
		return
	}
	for _, o := range opts {
		o(&gw.WriterOption)
	}
}

// Error 写入错误
func Error(ctx context.Context, err error) {
	if gw == nil {
		return
	}
	gw.Error(ctx, 1, err, "")
}

// Errorf 写入错误和格式化消息
func Errorf(ctx context.Context, err error, format string, a ...interface{}) {
	if gw == nil {
		return
	}
	gw.Error(ctx, 1, err, fmt.Sprintf(format, a...))
}

// ErrorWith 写入错误和消息
func ErrorWith(ctx context.Context, err error, detail string) {
	if gw == nil {
		return
	}
	gw.Error(ctx, 1, err, detail)
}

// Merge 根据trace_id为错误的消息增加一条数据,例如curl信息
func Merge(ctx context.Context, msg string) {
	if gw == nil {
		return
	}
	gw.MergeInfo(ctx, 1, msg)
}

type ConfigParam struct {
	URI        string        `json:"uri"`
	RoutingKey string        `json:"routing_key"`
	Exchange   string        `json:"exchange"`
	Limit      uint64        `json:"limit"`
	Enable     uint32        `json:"enable"` // 1 enable,2 unable
	Topic      string        `json:"topic"`
	Timeout    time.Duration `json:"timeout"`
}

// UpdateConfig 更新配置文件
func UpdateConfig(param *ConfigParam) error {
	if gw == nil {
		return nil
	}
	if err := updateMQ(param); err != nil {
		return err
	}
	if param.Enable != 0 {
		if param.Enable == 1 {
			gw.Cfg.SetEnable(true)
		} else if param.Enable == 2 {
			gw.Cfg.SetEnable(false)
		}
	}
	if param.Limit != 0 {
		gw.Cfg.SetLimit(param.Limit)
	}
	if param.Timeout != 0 {
		gw.Cfg.SetTimeout(param.Timeout)
	}
	return nil
}

// GetConfig 获取配置
func GetConfig() *ConfigParam {
	if gw == nil {
		return nil
	}
	var enable uint32 = 2
	if gw.Cfg.Enable() {
		enable = 1
	}
	return &ConfigParam{
		URI:        gw.Cfg.URI(),
		RoutingKey: gw.Cfg.RoutingKey(),
		Exchange:   gw.Cfg.Exchange(),
		Limit:      gw.Cfg.Limit(),
		Enable:     enable,
		Topic:      gw.Cfg.Queue(),
		Timeout:    gw.Cfg.Timeout(),
	}
}

// updatePublisher 更新生产者
func updatePublisher(publisher async.Producer) error {
	if gw == nil {
		return nil
	}
	temp := gw.GetPublisher()
	if err := gw.SetPublisher(publisher); err != nil {
		return err
	}
	_ = temp.Close() // ignore err
	return nil
}

func updateChannel(channel async.Channel) error {
	if gw == nil {
		return nil
	}
	return gw.SetChannel(channel)
}

func updateMQ(param *ConfigParam) error {
	if gw == nil {
		return nil
	}
	if value, flag := configChange(param); flag.Flag != 0 {
		if err := setPublisher(value); err != nil {
			return err
		}
		if flag.HasStatus(1) {
			gw.Cfg.SetURI(value.URI)
		}
		if flag.HasStatus(1 << 1) {
			gw.Cfg.SetExchange(value.exchange)
		}
		if flag.HasStatus(1 << 2) {
			gw.Cfg.SetRoutingKey(value.routingKey)
		}
		if flag.HasStatus(1 << 3) {
			gw.Cfg.SetQueue(value.queueName)
		}
	}
	return nil
}

type mqConfigStatus struct {
	queueName  string
	URI        string
	exchange   string
	routingKey string
}

func configChange(param *ConfigParam) (*mqConfigStatus, *msg.Status) {
	flag := &msg.Status{}
	temp := &mqConfigStatus{}

	if uri := gw.Cfg.URI(); param.URI != "" {
		if uri != param.URI {
			flag.AddStatus(1)
		}
		temp.URI = param.URI
	} else {
		temp.URI = uri
	}
	if exchange := gw.Cfg.Exchange(); param.Exchange != "" {
		if exchange != param.Exchange {
			flag.AddStatus(1 << 1)
		}
		temp.exchange = param.Exchange
	} else {
		temp.exchange = exchange
	}
	if routingKey := gw.Cfg.RoutingKey(); param.RoutingKey != "" {
		if routingKey != param.RoutingKey {
			flag.AddStatus(1 << 2)
		}
		temp.routingKey = param.RoutingKey
	} else {
		temp.routingKey = routingKey
	}
	if queueName := gw.Cfg.Queue(); param.Topic != "" {
		if queueName != param.Topic {
			flag.AddStatus(1 << 3)
		}
		temp.queueName = param.Topic
	} else {
		temp.queueName = queueName
	}
	return temp, flag
}

func setPublisher(value *mqConfigStatus) error {
	channel, err := async.NewRabbitmqChannel(
		async.WithAttempt(300),
		async.WithInterval(2*time.Second),
		async.WithURI(value.URI),
	)
	if err != nil {
		return err
	}
	// 生产者
	pub := async.NewTaskProducer()
	if err = updateChannel(channel); err != nil {
		return err
	}
	return updatePublisher(pub)
}
