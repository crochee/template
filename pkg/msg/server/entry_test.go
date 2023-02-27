// Package msg
package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"

	"template/config"
	"template/pkg/async"
	"template/pkg/logger"
	"template/pkg/msg"
)

func TestMain(m *testing.M) {
	// 初始化配置
	configFile := "../../../config/anchor.yaml"
	if err := config.LoadConfig(configFile); err != nil {
		log.Fatal(err.Error())
	}
	os.Exit(m.Run())
}

func TestError(t *testing.T) {
	ctx := logger.With(context.Background(),
		logger.New(
			logger.WithLevel(viper.GetString("log.level")),
			logger.WithWriter(logger.SetWriter(true))),
	)
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	channel, err := async.NewRabbitmqChannel(async.WithURI(viper.GetString("rabbitmq.producer.fault.URI")))
	if err != nil {
		t.Fatal(err)
	}
	ctx1, cancel2 := context.WithTimeout(ctx, 30*time.Second)
	defer cancel2()
	Update(ctx1, func(option *msg.WriterOption) {
		option.Publisher = async.NewTaskProducer()
		cfg := msg.NewCfgHandler()
		cfg.SetQueue(viper.GetString("rabbitmq.producer.fault.queue"))
		cfg.SetURI(viper.GetString("rabbitmq.producer.fault.URI"))
		cfg.SetExchange(viper.GetString("rabbitmq.producer.fault.exchange"))
		cfg.SetRoutingKey(viper.GetString("rabbitmq.producer.fault.routing-key"))
		cfg.SetEnable(true)
		cfg.SetTimeout(2 * time.Second)
		cfg.SetLimit(10)
		option.Cfg = cfg
		option.Channel = channel
		option.TraceID = func(ctx context.Context) string {
			return uuid.NewV4().String()
		}
	})

	Merge(ctx, "debug")
	for i := 0; i < 10007; i++ {
		index := strconv.Itoa(i)
		Errorf(context.Background(), errors.New(index), "io%d", i)
	}
	r := msg.NewReader(func(option *msg.ReaderOption) {
		option.Handles = append(option.Handles, handleLog)
	})
	if err = r.Subscribe(ctx, channel, viper.GetString("rabbitmq.producer.fault.queue")); err != nil {
		t.Fatal(err)
	}
}

func handleLog(metadata *msg.Metadata) error {
	fmt.Printf("%+v\n", metadata)
	return nil
}
