// Package msg
package server_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"template/pkg/async"
	"template/pkg/logger"
	"template/pkg/logger/gormx"
	"template/pkg/msg"
	"template/pkg/msg/server"
)

func TestError(t *testing.T) {

	ctx, cancel := context.WithTimeout(logger.With(context.Background(), logger.New()), 20*time.Second)
	defer cancel()

	ctl := gomock.NewController(t)
	defer ctl.Finish()
	p := async.NewMockProducer(ctl)
	p.EXPECT().Publish(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ async.Channel, exchange, routingKey string, param interface{}) error {
			logger.From(ctx).Sugar().Infof("ex:%s key:%s ,%s\n", exchange, routingKey, param.([]amqp.Publishing)[0].Body)
			return nil
		}).
		AnyTimes()

	form := func(context.Context) gormx.Logger {
		return gormx.NewZapGormWriterFrom(ctx)
	}
	exp := msg.NewWriter(func(o *msg.WriterOption) {
		o.Publisher = p
		o.From = form
	})

	tpOpts := []sdktrace.TracerProviderOption{
		sdktrace.WithSpanProcessor(
			msg.NewBatchSpanProcessor(exp,
				msg.WithMaxQueueSize(2),
				msg.WithMaxExportBatchSize(1),
				msg.WithLoggerFrom(form),
			)),
		sdktrace.WithIDGenerator(msg.DefaultIDGenerator(func(context.Context) string {
			return uuid.NewV4().String()
		})),
	}
	tp := sdktrace.NewTracerProvider(
		tpOpts...,
	)
	otel.SetTracerProvider(tp)

	for i := 0; i < 6; i++ {
		ctx, span := otel.Tracer("server").Start(ctx, fmt.Sprintf("TestError%d", i))
		server.Merge(ctx, "debug")
		server.Resource(ctx, "12", "name", "34", "sub_name")
		for i := 0; i < 2; i++ {
			index := strconv.Itoa(i)
			server.Errorf(ctx, errors.New(index), "io%d", i)
		}
		span.End()
	}
	time.Sleep(2 * time.Second)
}
