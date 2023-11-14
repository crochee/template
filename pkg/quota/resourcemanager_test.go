package quota

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/net/context"

	"template/pkg/quota/errcode"
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

func initRedisAndHander(ctx context.Context) {
	initRedis()
	viper.SetDefault("user_quota.enable", "true")
	InitResourceQuotaManager(time.Second, time.Minute*3,
		func(ctx context.Context) context.Context {
			return ctx
		},
		WithUsedQuotaHandler("CPUA", &mockHandler{Used: 10, Quota: 20}),
		WithUsedQuotaHandler("CPUB", &mockHandler{Used: 0, Quota: 20}),
		WithUsedQuotaHandler("CPUC", &mockHandler{Used: 18, Quota: 20}),
		WithUsedQuotaHandler("CPUD", &mockHandler{Used: 20, Quota: 20}),
	)

	//
	//InitResourceQuotaData(ctx,)
}

func TestResourceManagerPrepareOccupying(t *testing.T) {
	ctx := context.Background()
	initRedisAndHander(ctx)

	type arg struct {
		account     string
		requirement map[string]uint
	}
	tests := []struct {
		name    string
		args    arg
		wantErr error
	}{
		{
			name: "test PrepareOccupying ok",
			args: arg{
				account: "test a",
				requirement: map[string]uint{
					"CPUA": 3,
					"CPUB": 3,
					"CPUC": 1,
				},
			},
			wantErr: nil,
		},
		{
			name: "test PrepareOccupying fail",
			args: arg{
				account: "test a",
				requirement: map[string]uint{
					"CPUA": 3,
					"CPUB": 3,
					"CPUC": 5,
				},
			},
			wantErr: errcode.ErrCodeResourceQuotaInsufficient.WithResult("CPUC"),
		},
	}

	for _, test := range tests {
		finish, err := PrepareOccupying(ctx, test.args.account, test.args.requirement)
		if !errors.Is(err, test.wantErr) {
			t.Logf("%s PrepareOccupying fail,error:%v, want error:%v", test.name, err, test.wantErr)
		}
		if err == nil {
			_ = finish.Finally(ctx)
		}
	}
}

func TestResourceManagerPrepareOccupyingAndRollback(t *testing.T) {
	ctx := context.Background()
	initRedisAndHander(ctx)

	type arg struct {
		account     string
		requirement map[string]uint
	}
	tests := []struct {
		name    string
		args    arg
		wantErr error
	}{
		{
			name: "test PrepareOccupying ok",
			args: arg{
				account: "test a",
				requirement: map[string]uint{
					"CPUA": 3,
					"CPUB": 3,
				},
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		finish, err := PrepareOccupying(ctx, test.args.account, test.args.requirement)
		if err != nil {
			t.Logf("%s PrepareOccupying fail,error:%v", test.name, err)
			continue
		}
		err = finish.Rollback(ctx)
		if !errors.Is(err, test.wantErr) {
			t.Logf("%s ,rollback fail,error:%v, want error:%v", test.name, err, test.wantErr)
		}
		_ = finish.Finally(ctx)
	}
}

func TestResourceManagerPrepareOccupyingAndRefresh(t *testing.T) {
	ctx := context.Background()
	initRedisAndHander(ctx)

	type arg struct {
		account     string
		requirement map[string]uint
	}
	tests := []struct {
		name    string
		args    arg
		wantErr error
	}{
		{
			name: "test PrepareOccupying ok",
			args: arg{
				account: "test a",
				requirement: map[string]uint{
					"CPUA": 3,
					"CPUB": 3,
				},
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		wait := sync.WaitGroup{}
		finish, err := PrepareOccupying(ctx, test.args.account, test.args.requirement)
		if err != nil {
			t.Logf("%s PrepareOccupying fail,error:%v", test.name, err)
			continue
		}
		wait.Add(1)
		go func(account string) {
			err = RefreshAccountUsedQuota(ctx, account, false, "CPUD")
			t.Logf("%s刷新任务结束,error:%v", test.name, err)
			wait.Done()
		}(test.args.account)
		_ = finish.Finally(ctx)
		wait.Wait()
	}
}

func TestResourceManagerRollback(t *testing.T) {
	ctx := context.Background()
	initRedisAndHander(ctx)

	type arg struct {
		account     string
		requirement map[string]uint
		errTime     time.Time
	}
	tests := []struct {
		name    string
		args    arg
		wantErr error
	}{
		{
			name: "test Rollback ok",
			args: arg{
				account: "test a",
				requirement: map[string]uint{
					"CPUA": 3,
					"CPUB": 3,
				},
				errTime: time.Now(),
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		err := Rollback(ctx, test.args.account, test.args.requirement, test.args.errTime)
		if err != nil {
			t.Logf("%s rollback fail,error:%v", test.name, err)
			continue
		}
	}
}
