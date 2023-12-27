// Package msg
package msg

import (
	"context"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	uuid "github.com/gofrs/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/streadway/amqp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"template/pkg/async"
	"template/pkg/logger/gormx"
)

func NewWriter(opts ...func(*WriterOption)) *Writer {
	w := &Writer{
		WriterOption: WriterOption{
			ServiceNameFunc: ServiceName,
			marshal:         async.DefaultMarshal{},
			MetadataPool:    NewMetadataPool(),
			Cfg:             NewCfgHandler(),
			JSONHandler:     jsoniter.ConfigCompatibleWithStandardLibrary,
			Publisher:       async.NoopProducer{},
			Channel:         async.NoopChannel{},
			From:            gormx.Nop,
		},
	}

	for _, o := range opts {
		o(&w.WriterOption)
	}

	return w
}

type WriterOption struct {
	ServiceNameFunc ServiceNameHandler
	marshal         async.MarshalAPI // mq  assemble request or response
	MetadataPool    MetadataPool
	Cfg             CfgHandler
	JSONHandler     jsoniter.API
	Publisher       async.Producer
	Channel         async.Channel
	From            func(context.Context) gormx.Logger
}

type Writer struct {
	WriterOption
	mux sync.RWMutex
}

type DescContent struct {
	List []Event `json:"list"`
}

type Event struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
	Time  time.Time   `json:"time"`
}

type HTTPInfo struct {
	Request  string `json:"request"`
	Response string `json:"response"`
	Status   string `json:"status"`
}

func (w *Writer) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	if len(spans) == 0 {
		return nil
	}
	result := make([]amqp.Publishing, 0, len(spans))
	for i := range spans {
		metadata := w.MetadataPool.Get()
		events := spans[i].Events()
		// tempEvents := make([]sdktrace.Event, 0, len(events))
		tempEvents := &DescContent{}
		// 去重map初始化
		oreqMap := make(map[string]Event)

		for _, event := range events {
			switch event.Name {
			case semconv.ExceptionEventName:
				metadata.ErrorTime = event.Time
				for _, attr := range event.Attributes {
					switch attr.Key {
					case semconv.ExceptionMessageKey:
						metadata.Summary = attr.Value.AsString()
						tempEvents.List = append(tempEvents.List, Event{
							Name:  "exception",
							Value: metadata.Summary,
							Time:  event.Time,
						})
					case MsgKey:
						tempEvents.List = append(tempEvents.List, Event{
							Name:  "exception description",
							Value: attr.Value.AsString(),
							Time:  event.Time,
						})
					default:
					}
				}
			case CurlEvent:
				for _, tempAttr := range event.Attributes {
					if tempAttr.Key == MsgKey {
						var content HTTPInfo
						if err := w.JSONHandler.Unmarshal([]byte(tempAttr.Value.AsString()), &content); err != nil {
							oreqMap[tempAttr.Value.AsString()] = Event{
								Name:  "http info",
								Value: tempAttr.Value.AsString(),
								Time:  event.Time,
							}
							continue
						}
						oreqMap[content.Request] = Event{
							Name:  "http info",
							Value: content,
							Time:  event.Time,
						}
					}
				}
			default:
			}
		}
		if len(tempEvents.List) == 0 {
			w.MetadataPool.Put(metadata)
			continue
		}
		for _, httpInfo := range oreqMap {
			tempEvents.List = append(tempEvents.List, httpInfo)
		}
		data, err := w.JSONHandler.Marshal(tempEvents)
		if err != nil {
			w.From(ctx).Errorf("marshal events failed,%+v", err)
			continue
		}
		metadata.Desc = string(data)

		for _, attr := range spans[i].Attributes() {
			switch attr.Key {
			case LocateKey:
				metadata.Locate = attr.Value.AsString()
			case AccountIDKey:
				metadata.AccountID = attr.Value.AsString()
			case UserIDKey:
				metadata.UserID = attr.Value.AsString()
			case ResIDKey:
				metadata.ResID = attr.Value.AsString()
			case ResTypeKey:
				metadata.ResType = attr.Value.AsString()
			case SubResIDKey:
				metadata.SubResID = attr.Value.AsString()
			case SubResTypeKey:
				metadata.SubResType = attr.Value.AsString()
			default:
			}
		}
		metadata.TraceID = "req-" + uuid.UUID(spans[i].SpanContext().TraceID()).String()
		metadata.ServiceName = w.ServiceNameFunc()
		metadata.SpanID = spans[i].SpanContext().SpanID().String()

		if spans[i].Parent().HasSpanID() {
			metadata.ParentSpanID = spans[i].Parent().SpanID().String()
		}
		data, err = w.JSONHandler.Marshal(metadata)
		w.MetadataPool.Put(metadata)
		if err != nil {
			w.From(ctx).Errorf("Publish failed,%+v,trace_id:%s", err, metadata.TraceID)
			continue
		}
		msg, err := w.marshal.Marshal(message.NewMessage(metadata.TraceID, data))
		if err != nil {
			w.From(ctx).Errorf("marshal failed,%+v,trace_id:%s", err, metadata.TraceID)
			continue
		}
		result = append(result, msg)
	}
	if err := w.GetPublisher().Publish(ctx, w.GetChannel(), w.Cfg.Exchange(), w.Cfg.RoutingKey(), result); err != nil {
		w.From(ctx).Errorf("Publish failed,%+v", err)
	}
	return nil
}

func (w *Writer) Shutdown(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}

func (w *Writer) SetPublisher(publisher async.Producer) {
	w.mux.Lock()
	w.Publisher = publisher
	w.mux.Unlock()
}

func (w *Writer) GetPublisher() async.Producer {
	w.mux.RLock()
	publisher := w.Publisher
	w.mux.RUnlock()
	return publisher
}

func (w *Writer) SetChannel(channel async.Channel) {
	w.mux.Lock()
	w.Channel = channel
	w.mux.Unlock()
}

func (w *Writer) GetChannel() async.Channel {
	w.mux.RLock()
	channel := w.Channel
	w.mux.RUnlock()
	return channel
}
