package quota

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"

	"template/pkg/syncx"
)

func TestPrepareOccupying(t *testing.T) {
	ctx := context.Background()

	redisClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			":7001",
			":7002",
			":7003",
			":7004",
			":7005",
			":7000"},
	})
	defer redisClient.Close()
	InitResourceQuotaManager(
		0,
		0,
		nil,
		WithUsedQuotaHandler("CPUA", &mockHandler{Used: 1, Quota: 4}),
		WithUsedQuotaHandler("CPUB", &mockHandler{Used: 2, Quota: 4}),
		WithUsedQuotaHandler("CPUC", &mockHandler{Used: 3, Quota: 4}),
		WithIsQuotaEnable(func(accounts ...string) (bool, error) {
			return true, nil
		}),
		WithLockFn(func(key string) syncx.Locker {
			return syncx.NewMutex(key, redisClient)
		}),
		WithFinisherFn(
			func(handler UsedQuotaHandler, param *Param, lock syncx.Locker) (FinishQuota, error) {
				return NewRedisFinishQuota(
					handler,
					param,
					lock,
					redisClient,
					3*time.Second,
				), nil
			},
		),
	)

	type args struct {
		ctx         context.Context
		account     string
		requirement map[string]uint
	}
	tests := []struct {
		name      string
		args      args
		want      FinishQuota
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "ok",
			args: args{
				ctx:     ctx,
				account: "test",
				requirement: map[string]uint{
					"CPUA": 1,
					"CPUB": 1,
					"CPUC": 1,
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "fail",
			args: args{
				ctx:     ctx,
				account: "test",
				requirement: map[string]uint{
					"CPUA": 3,
					"CPUB": 3,
					"CPUC": 5,
				},
			},
			assertion: func(t assert.TestingT, err error, i ...interface{}) bool {
				flag := assert.Error(t, err, i...)
				if flag {
					assert.True(t, errors.Is(err, ErrResourceQuotaInsufficient))
				}
				return flag
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PrepareOccupying(tt.args.ctx, tt.args.account, tt.args.requirement)
			tt.assertion(t, err)
			if err != nil {
				return
			}
			defer func() {
				if err := got.Rollback(ctx); err != nil {
					t.Log(err)
				}
			}()
			if err = got.Finally(ctx); err != nil {
				t.Log(err)
			}
		})
	}
}
