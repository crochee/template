package server

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/automaxprocs/maxprocs"

	"template/pkg/async"
	"template/pkg/logger/gormx"
	"template/pkg/msg"
)

// Automatically set GOMAXPROCS to match Linux container CPU quota.
func init() {
	maxprocs.Set(maxprocs.Logger(nil))
}

var (
	exp *msg.Writer

	processorRWMutex sync.RWMutex
	processor        *msg.BatchSpanProcessor
)

// New 更新全局参数
func New(ctx context.Context, getTraceID func(context.Context) string, opts ...func(*msg.WriterOption)) {
	exp = msg.NewWriter(opts...)

	processor = msg.NewBatchSpanProcessor(exp,
		msg.WithLoggerFrom(func(context.Context) gormx.Logger {
			return gormx.NewZapGormWriterFrom(ctx)
		}),
	)

	tpOpts := []sdktrace.TracerProviderOption{
		sdktrace.WithSpanProcessor(processor),
		sdktrace.WithIDGenerator(msg.DefaultIDGenerator(getTraceID)),
	}
	tp := sdktrace.NewTracerProvider(
		tpOpts...,
	)
	otel.SetTracerProvider(tp)
}

// Error 写入错误
func Error(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(msg.LocateKey.String(msg.CallerFunc(0)))
	span.RecordError(err)
}

// Errorf 写入错误和格式化消息
func Errorf(ctx context.Context, err error, format string, a ...interface{}) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(msg.LocateKey.String(msg.CallerFunc(0)))
	span.RecordError(err,
		trace.WithAttributes(msg.MsgKey.String(fmt.Sprintf(format, a...))))
}

// ErrorWith 写入错误和消息
func ErrorWith(ctx context.Context, err error, detail string) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(msg.LocateKey.String(msg.CallerFunc(0)))
	span.RecordError(err,
		trace.WithAttributes(msg.MsgKey.String(detail)))
}

// Merge 根据trace_id为错误的消息增加一条数据,例如curl信息
func Merge(ctx context.Context, info string) {
	trace.SpanFromContext(ctx).AddEvent(msg.CurlEvent, trace.WithAttributes(msg.MsgKey.String(info)))
}

func Resource(ctx context.Context, resource, resourceType string, subRes ...string) {
	if resource == "" || resourceType == "" {
		return
	}
	attrs := []attribute.KeyValue{
		msg.ResIDKey.String(resource),
		msg.ResTypeKey.String(resourceType),
	}
	for i, v := range subRes {
		switch i {
		case 0:
			attrs = append(attrs, msg.SubResIDKey.String(v))
		case 1:
			attrs = append(attrs, msg.SubResTypeKey.String(v))
		default:
		}
	}
	trace.SpanFromContext(ctx).SetAttributes(attrs...)
}

type ConfigParam struct {
	URI        string `json:"uri"`
	RoutingKey string `json:"routing_key"`
	Exchange   string `json:"exchange"`
	Enable     uint32 `json:"enable"` // 1 enable,2 unable
	Topic      string `json:"topic"`
}

// UpdateConfig 更新配置文件
func UpdateConfig(param *ConfigParam) error {
	if exp == nil {
		return nil
	}
	if param.Enable != 0 {
		s := exp.Cfg.Status()
		s.CheckChangeClose(func() {
			exp.Channel = async.NoopChannel{}
		})
		s.CheckChangeOpen(func() {
			if err := updateMQPublisher(param); err != nil {
				log.Println(err)
			}
		})
		switch param.Enable {
		case 1:
			exp.Cfg.SetEnable(true)
		case 2:
			exp.Cfg.SetEnable(false)
		default:
		}
	}
	return nil
}

// GetConfig 获取配置
func GetConfig() *ConfigParam {
	if exp == nil {
		return nil
	}
	var enable uint32 = 2
	if exp.Cfg.Enable() {
		enable = 1
	}
	return &ConfigParam{
		URI:        exp.Cfg.URI(),
		RoutingKey: exp.Cfg.RoutingKey(),
		Exchange:   exp.Cfg.Exchange(),
		Enable:     enable,
		Topic:      exp.Cfg.Topic(),
	}
}

func updateMQPublisher(param *ConfigParam) error {
	if exp == nil {
		return nil
	}
	if value, flag := configChange(param); flag.Flag != 0 {
		if err := setPublisher(value); err != nil {
			return err
		}
		if flag.HasStatus(1) {
			exp.Cfg.SetURI(value.URI)
		}
		if flag.HasStatus(1 << 1) {
			exp.Cfg.SetExchange(value.exchange)
		}
		if flag.HasStatus(1 << 2) {
			exp.Cfg.SetRoutingKey(value.routingKey)
		}
		if flag.HasStatus(1 << 3) {
			exp.Cfg.SetTopic(value.topic)
		}
	}
	return nil
}

type mqConfigStatus struct {
	URI        string
	exchange   string
	routingKey string
	topic      string
}

func configChange(param *ConfigParam) (*mqConfigStatus, *msg.Status) {
	flag := &msg.Status{}
	temp := &mqConfigStatus{}

	if uri := exp.Cfg.URI(); param.URI != "" {
		if uri != param.URI {
			flag.AddStatus(1)
		}
		temp.URI = param.URI
	} else {
		temp.URI = uri
	}
	if exchange := exp.Cfg.Exchange(); param.Exchange != "" {
		if exchange != param.Exchange {
			flag.AddStatus(1 << 1)
		}
		temp.exchange = param.Exchange
	} else {
		temp.exchange = exchange
	}
	if routingKey := exp.Cfg.RoutingKey(); param.RoutingKey != "" {
		if routingKey != param.RoutingKey {
			flag.AddStatus(1 << 2)
		}
		temp.routingKey = param.RoutingKey
	} else {
		temp.routingKey = routingKey
	}
	if topic := exp.Cfg.Topic(); param.Topic != "" {
		if topic != param.Topic {
			flag.AddStatus(1 << 3)
		}
		temp.topic = param.Topic
	} else {
		temp.topic = topic
	}

	return temp, flag
}

func setPublisher(value *mqConfigStatus) error {
	if exp == nil {
		return nil
	}
	channel, err := async.NewRabbitmqChannel(
		async.WithURI(value.URI),
		async.WithTx(true),
		async.WithAttempt(12*60*24*2),
		async.WithInterval(5*time.Second),
	)
	if err != nil {
		return err
	}
	if err = channel.DeclareAndBind(value.exchange, "direct", value.topic, value.routingKey); err != nil {
		return err
	}
	exp.SetChannel(channel)
	return nil
}
