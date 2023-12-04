// Package msg
package server_test

import (
	"context"
	"errors"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"template/pkg/async"
	"template/pkg/msg"
	"template/pkg/msg/server"
)

func TestError(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	p := async.NewMockProducer(ctl)
	p.EXPECT().Publish(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ async.Channel, exchange, routingKey string, param interface{}) error {
			log.Printf("ex:%s key:%s ,%s\n", exchange, routingKey, param.([]amqp.Publishing)[0].Body)
			return nil
		}).
		AnyTimes()

	exp := msg.NewWriter(func(o *msg.WriterOption) {
		o.Publisher = p
	})

	tpOpts := []sdktrace.TracerProviderOption{
		sdktrace.WithSyncer(exp),
		sdktrace.WithIDGenerator(msg.DefaultIDGenerator(func(context.Context) string {
			return uuid.NewV4().String()
		})),
	}
	tp := sdktrace.NewTracerProvider(
		tpOpts...,
	)
	otel.SetTracerProvider(tp)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	ctx, span := otel.Tracer("server").Start(ctx, "TestError")
	server.Merge(ctx, "debug")
	server.Resource(ctx, "12", "name", "34", "sub_name")
	for i := 0; i < 2; i++ {
		index := strconv.Itoa(i)
		server.Errorf(ctx, errors.New(index), "io%d", i)
	}
	span.End()
}
