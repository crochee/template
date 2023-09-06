// Package msg
package msg

import (
	"context"
	"sync"

	"github.com/ThreeDotsLabs/watermill/message"
	uuid "github.com/gofrs/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/streadway/amqp"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"

	"template/pkg/async"
	"template/pkg/json"
	"template/pkg/logger"
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
}

type Writer struct {
	WriterOption
	mux sync.RWMutex
}

func (w *Writer) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	if len(spans) == 0 {
		return nil
	}
	result := make([]amqp.Publishing, 0, len(spans))
	for i := range spans {
		metadata := w.MetadataPool.Get()
		var hasErr bool
		events := spans[i].Events()
		tempEvents := make([]sdktrace.Event, 0, len(events))

		for _, event := range events {
			switch event.Name {
			case semconv.ExceptionEventName:
				metadata.ErrorTime = event.Time
				for _, attr := range event.Attributes {
					if attr.Key == semconv.ExceptionMessageKey {
						metadata.Summary = attr.Value.AsString()
					}
				}
				hasErr = true
			case CurlEvent:
				if len(tempEvents) > 0 {
					// 对curl事件进行裁剪，防止出现循环请求，导致内容过长
					oreq := w.getRequest(tempEvents[len(tempEvents)-1].Attributes)
					req := w.getRequest(event.Attributes)
					if oreq != "" && req == oreq {
						continue
					}
				}
			default:
			}
			tempEvents = append(tempEvents, sdktrace.Event{
				Name:                  event.Name,
				Attributes:            event.Attributes,
				DroppedAttributeCount: event.DroppedAttributeCount,
				Time:                  event.Time,
			})
		}
		if !hasErr {
			w.MetadataPool.Put(metadata)
			continue
		}
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
		data, err := w.JSONHandler.Marshal(tempEvents)
		if err != nil {
			logger.From(ctx).Error("marshal events failed", zap.Error(err))
			continue
		}
		metadata.Desc = string(data)
		metadata.TraceID = "req-" + uuid.UUID(spans[i].SpanContext().TraceID()).String()
		metadata.ServiceName = w.ServiceNameFunc()
		metadata.SpanID = spans[i].SpanContext().SpanID().String()

		if spans[i].Parent().HasSpanID() {
			metadata.ParentSpanID = spans[i].Parent().SpanID().String()
		}
		data, err = w.JSONHandler.Marshal(metadata)
		w.MetadataPool.Put(metadata)
		if err != nil {
			logger.From(ctx).Error("Publish failed", zap.Error(err))
			continue
		}
		msg, err := w.marshal.Marshal(message.NewMessage(metadata.TraceID, data))
		if err != nil {
			logger.From(ctx).Error("marshal failed", zap.Error(err))
			continue
		}
		result = append(result, msg)
	}
	if err := w.GetPublisher().Publish(ctx, w.GetChannel(), w.Cfg.Exchange(), w.Cfg.RoutingKey(), result); err != nil {
		logger.From(ctx).Error("Publish failed", zap.Error(err))
	}
	return nil
}

func (w *Writer) getRequest(attrs []attribute.KeyValue) string {
	for _, tempAttr := range attrs {
		if tempAttr.Key == MsgKey {
			var content struct {
				Request  string
				Response string
				Status   string
			}
			if err := json.Unmarshal([]byte(tempAttr.Value.AsString()), &content); err == nil {
				return content.Request
			}
		}
	}
	return ""
}

func (w *Writer) Shutdown(ctx context.Context) error {
	defer w.GetPublisher().Close()
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
