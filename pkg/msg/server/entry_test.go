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
	jsoniter "github.com/json-iterator/go"
	"github.com/streadway/amqp"
	"go.opentelemetry.io/otel"

	"template/internal/ctxw"
	"template/pkg/async"
	"template/pkg/msg"
	"template/pkg/msg/server"
)

func TestError(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	cc := async.NewMockChannel(ctl)
	cc.EXPECT().Publish(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(exchange, key string, mandatory, immediate bool, msg ...amqp.Publishing) error {
			for _, v := range msg {
				var data interface{}
				if err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(v.Body, &data); err != nil {
					return err
				}
				log.Printf("ex:%s key:%s %v %v,%#v\n", exchange, key, mandatory, immediate, data)
			}
			return nil
		}).AnyTimes()

	p := async.NewMockProducer(ctl)
	p.EXPECT().Publish(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ async.Channel, exchange, routingKey string, param interface{}) error {
			log.Printf("ex:%s key:%s ,%s\n", exchange, routingKey, param.([]amqp.Publishing)[0].Body)
			return nil
		}).
		AnyTimes()
	server.New(ctxw.GetTraceID, func(o *msg.WriterOption) {
		o.Publisher = p
		o.Channel = cc
	})
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	ctx, span := otel.Tracer("server").Start(ctx, "TestError")
	server.Merge(ctx, "debug")
	server.Resource(ctx, "12", "name", "34", "sub_name")
	for i := 0; i < 1; i++ {
		index := strconv.Itoa(i)
		server.Errorf(ctx, errors.New(index), "io%d", i)
	}
	span.End()
	time.Sleep(6 * time.Second)
}
