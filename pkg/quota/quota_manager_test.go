package quota

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"

	"template/pkg/syncx"
	"template/pkg/utils"
)

type mockHandler struct {
	Used  int
	Quota int
}

func (m *mockHandler) QueryUsed(ctx context.Context, account string) (used int, err error) {
	return m.Used, nil
}

func (m *mockHandler) QueryQuota(ctx context.Context, account string) (quota int, err error) {
	return m.Quota, nil
}

func TestTransaction(t *testing.T) {
	ctx := context.Background()

	mgr := NewResourceQuotaManager(
		WithUsedQuotaHandler("CPUA", &mockHandler{Used: 1, Quota: 4}),
		WithUsedQuotaHandler("CPUB", &mockHandler{Used: 2, Quota: 4}),
		WithUsedQuotaHandler("CPUC", &mockHandler{Used: 3, Quota: 4}),
		WithIsQuotaEnable(func(accounts ...string) (bool, error) {
			return true, nil
		}),
	)

	type args struct {
		ctx    context.Context
		params []*Param
	}
	tests := []struct {
		name      string
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "ok",
			args: args{
				ctx: ctx,
				params: []*Param{
					{
						AssociatedID: "test a",
						Name:         "CPUA",
						Num:          1,
					},
					{
						AssociatedID: "test a",
						Name:         "CPUB",
						Num:          1,
					},
					{
						AssociatedID: "test a",
						Name:         "CPUC",
						Num:          1,
					},
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "fail",
			args: args{
				ctx: ctx,
				params: []*Param{
					{
						AssociatedID: "test a",
						Name:         "CPUA",
						Num:          3,
					},
					{
						AssociatedID: "test a",
						Name:         "CPUB",
						Num:          3,
					},
					{
						AssociatedID: "test a",
						Name:         "CPUC",
						Num:          5,
					},
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
			tt.assertion(t, mgr.Transaction(ctx, tt.args.params, func(ctx context.Context) error {

				return nil
			}))
		})
	}
}

func BenchmarkTransaction(b *testing.B) {
	redisClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			":7001",
			":7002",
			":7003",
			":7004",
			":7005",
			":7000"},
	})
	mgr := NewResourceQuotaManager(
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
			func(handler UsedQuotaHandler, param *Param, lock syncx.Locker, state *utils.Status) (FinishQuota, error) {
				return NewRedisFinishQuota(
					handler,
					param,
					lock,
					redisClient,
					time.Hour,
					state,
				), nil
			},
		),
	)
	params := []*Param{
		{
			AssociatedID: "test a",
			Name:         "CPUA",
			Num:          3,
		},
		{
			AssociatedID: "test a",
			Name:         "CPUB",
			Num:          3,
		},
		{
			AssociatedID: "test a",
			Name:         "CPUC",
			Num:          5,
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.Transaction(context.Background(), params, func(ctx context.Context) error {
			return nil
		})
	}
}
