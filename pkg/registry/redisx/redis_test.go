package redisx

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"

	"template/pkg/logger"
	"template/pkg/registry"
)

func TestRedis(t *testing.T) {
	ctx := logger.With(context.Background(), logger.New())
	cli := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			":7000",
			":7001",
			":7002",
			":7003",
			":7004",
			":7005",
		},
	})
	defer cli.Close()
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	ef := NewMockEventFlow(ctl)
	ef.EXPECT().Change(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, msg *redis.Message) {
			logger.From(ctx).
				WithOptions(zap.WithCaller(false)).
				Sugar().Infof("%#v", msg)
		}).AnyTimes()
	encoder := NewMockEncoder(ctl)
	encoder.EXPECT().Encode(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(action Action, info *registry.Info, expireAt int64) (string, error) {
			// add task id
			value, found := info.Tags["task_id"]
			if found {
				if tasks, ok := value.([]string); ok {
					info.Tags["task_id"] = append(tasks, []string{"12", "34"}...)
				}
			} else {
				info.Tags["task_id"] = []string{"12", "34"}
			}
			return DefaultEncoder{}.Encode(action, info, expireAt)
		}).AnyTimes()
	encoder.EXPECT().Decode(gomock.Any()).DoAndReturn(
		func(data string) (Action, *registry.Info, int64, error) {
			return DefaultEncoder{}.Decode(data)
		}).AnyTimes()
	r := NewRedisRegistry(ctx,
		cli,
		WithEncoder(encoder),
		WithEventFlow(ef),
		WithExpireTime(30*time.Second),
		WithTickerTime(15*time.Second),
	)
	info := &registry.Info{
		UUID:        uuid.NewV4().String(),
		ServiceName: "template",
		Addr:        "localhost:10086",
		Weight:      0,
		Tags: map[string]interface{}{
			"task_id": []string{"1"},
		},
	}
	if err := r.Register(info); err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := r.Deregister(info); err != nil {
			t.Error(err)
		}
	}()
	time.Sleep(70 * time.Second)
}

func TestRedis1(t *testing.T) {
	ctx := logger.With(context.Background(), logger.New())
	cli := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			":7000",
			":7001",
			":7002",
			":7003",
			":7004",
			":7005",
		},
	})
	defer cli.Close()
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	ef := NewMockEventFlow(ctl)
	ef.EXPECT().Change(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, msg *redis.Message) {
			logger.From(ctx).
				WithOptions(zap.WithCaller(false)).
				Sugar().Infof("%#v", msg)
		}).AnyTimes()
	encoder := NewMockEncoder(ctl)
	encoder.EXPECT().Encode(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(action Action, info *registry.Info, expireAt int64) (string, error) {
			// add task id
			value, found := info.Tags["task_id"]
			if found {
				if tasks, ok := value.([]string); ok {
					info.Tags["task_id"] = append(tasks, []string{"56", "78"}...)
				}
			} else {
				info.Tags["task_id"] = []string{"56", "78"}
			}
			return DefaultEncoder{}.Encode(action, info, expireAt)
		}).AnyTimes()
	encoder.EXPECT().Decode(gomock.Any()).DoAndReturn(
		func(data string) (Action, *registry.Info, int64, error) {
			return DefaultEncoder{}.Decode(data)
		}).AnyTimes()
	r := NewRedisRegistry(ctx,
		cli,
		WithEncoder(encoder),
		WithEventFlow(ef),
		WithExpireTime(30*time.Second),
		WithTickerTime(15*time.Second),
	)
	info := &registry.Info{
		UUID:        uuid.NewV4().String(),
		ServiceName: "template",
		Addr:        "localhost:10086",
		Weight:      0,
		Tags: map[string]interface{}{
			"task_id": []string{"2"},
		},
	}
	if err := r.Register(info); err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := r.Deregister(info); err != nil {
			t.Error(err)
		}
	}()
	time.Sleep(70 * time.Second)
}
