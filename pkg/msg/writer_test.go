// Package msg
package msg

import (
	"context"
	"encoding/json"
	"log"
	"template/pkg/logger/gormx"
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"
	uuid "github.com/satori/go.uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func TestNull(t *testing.T) {
	data, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", data)
}

var _ sdktrace.SpanExporter = &Exporter{}

// Exporter is an implementation of trace.SpanSyncer that writes spans to stdout.
type Exporter struct {
}

// SpanStub is a stand-in for a Span.
type SpanStub struct {
	Name                   string
	SpanContext            trace.SpanContext
	Parent                 trace.SpanContext
	SpanKind               trace.SpanKind
	StartTime              time.Time
	EndTime                time.Time
	Attributes             []attribute.KeyValue
	Events                 []sdktrace.Event
	Links                  []sdktrace.Link
	Status                 sdktrace.Status
	DroppedAttributes      int
	DroppedEvents          int
	DroppedLinks           int
	ChildSpanCount         int
	Resource               *resource.Resource
	InstrumentationLibrary instrumentation.Library
}

// ExportSpans writes spans in json format to stdout.
func (Exporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	if len(spans) == 0 {
		return nil
	}
	for i := range spans {
		// Encode span stubs, one by one
		data, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(SpanStub{
			Name:                   spans[i].Name(),
			SpanContext:            spans[i].SpanContext(),
			Parent:                 spans[i].Parent(),
			SpanKind:               spans[i].SpanKind(),
			StartTime:              spans[i].StartTime(),
			EndTime:                spans[i].EndTime(),
			Attributes:             spans[i].Attributes(),
			Events:                 spans[i].Events(),
			Links:                  spans[i].Links(),
			Status:                 spans[i].Status(),
			DroppedAttributes:      spans[i].DroppedAttributes(),
			DroppedEvents:          spans[i].DroppedEvents(),
			DroppedLinks:           spans[i].DroppedLinks(),
			ChildSpanCount:         spans[i].ChildSpanCount(),
			Resource:               spans[i].Resource(),
			InstrumentationLibrary: spans[i].InstrumentationLibrary(),
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%s %s %s %s\n", spans[i].SpanContext().SpanID(), uuid.UUID(spans[i].SpanContext().TraceID()).String(),
			spans[i].Parent().SpanID(),
			data)
	}
	return nil
}

// Shutdown is called to stop the exporter, it preforms no action.
func (Exporter) Shutdown(ctx context.Context) error {

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}

var (
	fooKey     = attribute.Key("ex.com/foo")
	barKey     = attribute.Key("ex.com/bar")
	anotherKey = attribute.Key("ex.com/another")
)

var (
	lemonsKey = attribute.Key("ex.com/lemons")
)

// SubOperation is an example to demonstrate the use of named tracer.
// It creates a named tracer with its package path.
func SubOperation(ctx context.Context) error {
	// Using global provider. Alternative is to have application provide a getter
	// for its component to get the instance of the provider.
	tr := otel.Tracer("example/namedtracer/foo")

	var span trace.Span
	_, span = tr.Start(ctx, "Sub operation...")
	defer span.End()
	span.SetAttributes(lemonsKey.String("five"))
	span.AddEvent("Sub span event")

	return nil
}

func TestS(t *testing.T) {
	exp := Exporter{}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithIDGenerator(DefaultIDGenerator(func(ctx context.Context) string {
			return uuid.NewV4().String()
		}, gormx.Nop)),
		sdktrace.WithSyncer(exp),
	)
	otel.SetTracerProvider(tp)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// ctx = ctxw.SetTraceID(ctx, uuid.NewV4().String())
	// Cleanly shutdown and flush telemetry when the application exits.
	defer func(ctx context.Context) {
		// Do not make the application hang when it is shutdown.
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := exp.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}(ctx)
	// Create a named tracer with package path as its name.
	tracer := otel.Tracer("example/namedtracer/main")

	m0, _ := baggage.NewMember(string(fooKey), "foo1")
	m1, _ := baggage.NewMember(string(barKey), "bar1")
	b, _ := baggage.New(m0, m1)
	ctx = baggage.ContextWithBaggage(ctx, b)

	var span trace.Span
	ctx, span = tracer.Start(ctx, "operation")
	defer span.End()
	span.AddEvent("Nice operation!", trace.WithAttributes(attribute.Int("bogons", 100)))
	span.SetAttributes(anotherKey.String("yes"))
	if err := SubOperation(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestEvent(t *testing.T) {
	value := &DescContent{
		List: []Event{
			{
				Name:  "exception",
				Value: "retry all count",
				Time:  time.Now(),
			},
			{
				Name:  "exception description",
				Value: "timeout to wait for volume 485380786413930760 status",
				Time:  time.Now(),
			},
			{
				Name: "http info",
				Value: HTTPInfo{
					Request:  `curl --location '172.31.248.3:30086/v2/hosts'`,
					Response: `{"code":1100009,"message":"没有预先创建订单，请先创建订单","result":null}`,
					Status:   "200",
				},
				Time: time.Now(),
			},
		},
	}
	desc, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	data := &Metadata{
		TraceID:      "",
		ServiceName:  "",
		Locate:       "",
		SpanID:       "",
		ParentSpanID: "",
		AccountID:    "",
		UserID:       "",
		ResID:        "",
		ResType:      "",
		SubResID:     "",
		SubResType:   "",
		Summary:      "retry all count",
		Desc:         string(desc),
		ErrorTime:    time.Now(),
	}
	t.Logf("%#v", data)
}
