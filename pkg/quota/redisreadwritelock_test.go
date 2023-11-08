package quota

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

func initRedis() {
	viper.SetDefault("redis.addrs", []string{":7001", ":7002", ":7003", ":7004", ":7005", ":7000"})
	viper.SetDefault("redis.password", "")
}

func TestRLock(t *testing.T) {
	initRedis()
	type args struct {
		key       string
		userId    uint64
		leaseTime time.Duration
		num       int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test read lock and unlock",
			args: args{
				key:       "lock-1",
				userId:    1,
				leaseTime: time.Minute,
				num:       1,
			},
			want: "",
		}, {
			name: "test read more lock",
			args: args{
				key:       "lock-2",
				userId:    2,
				leaseTime: time.Minute,
				num:       500,
			},
			want: "",
		},
	}
	for _, test := range tests {
		i := 0
		wait := sync.WaitGroup{}
		for i < test.args.num {
			wait.Add(1)
			go func(index int) {
				defer wait.Done()
				lock := CreateReadLock(context.Background(), test.args.key, test.args.userId+uint64(index), test.args.leaseTime)
				err := lock.Lock()
				if err != nil {
					t.Log("add read lock fail", err)
					return
				}
				time.Sleep(time.Second * 2)
				err = lock.Unlock()
				if err != nil {
					t.Log("un read lock fail", err)
				}
			}(i)
			i++
		}
		wait.Wait()
	}
}

func TestWLock(t *testing.T) {
	initRedis()
	type args struct {
		key       string
		userId    uint64
		leaseTime time.Duration
		num       int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "test write lock and unlock",
			args: args{
				key:       "lock-1",
				userId:    1,
				leaseTime: time.Minute,
				num:       1,
			},
			want: 1,
		}, {
			name: "test write more lock",
			args: args{
				key:       "lock-2",
				userId:    2,
				leaseTime: time.Minute,
				num:       500,
			},
			want: 500,
		},
	}
	for _, test := range tests {
		i := 0
		var result int
		var errNum int32
		wait := sync.WaitGroup{}
		for i < test.args.num {
			wait.Add(1)
			go func(index int) {
				defer wait.Done()
				lock := CreateWriteLock(context.Background(), test.args.key, test.args.userId+uint64(index), test.args.leaseTime)
				err := lock.Lock()
				if err != nil {
					atomic.AddInt32(&errNum, 1)
					t.Log("add write lock fail", err)
					return
				}
				t.Log(index, " 加锁成功")
				time.Sleep(time.Millisecond * 200)
				result++
				t.Log(index, " 解锁")
				err = lock.Unlock()
				if err != nil {
					t.Log("un write lock fail", err)
				}
			}(i)
			i++
		}
		wait.Wait()
		if int(errNum)+result != test.want {
			t.Error("数据之间存在并发操作")
		}
	}
}

func TestRWLock(t *testing.T) {
	initRedis()
	type args struct {
		key       string
		userId    uint64
		leaseTime time.Duration
		order     []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test first read and write lock",
			args: args{
				key:       "lock-1",
				userId:    100,
				leaseTime: time.Minute,
				order: []string{"r", "r", "w", "r", "r", "r", "w", "r", "r",
					"r", "r", "w", "r", "r", "r", "w", "r", "r", "r", "r", "w", "r", "r",
					"r", "r", "w", "r", "r", "r", "w", "r", "r", "r", "r", "w", "r", "r",
					"r", "r", "r", "r", "r", "r"},
			},
			want: "",
		},
	}

	for _, test := range tests {
		checkMap := make(map[int]interface{})
		wait := sync.WaitGroup{}
		for i, ty := range test.args.order {
			wait.Add(1)
			go func(index int, ty string) {
				defer wait.Done()
				var lock Locker
				if ty == "r" {
					lock = CreateReadLock(context.Background(), test.args.key, test.args.userId+uint64(index), test.args.leaseTime)
				} else {
					lock = CreateWriteLock(context.Background(), test.args.key, test.args.userId+uint64(index), test.args.leaseTime)
				}
				err := lock.Lock()
				if err != nil {
					t.Log("add "+ty+" lock fail", err)
					return
				}
				if ty == "r" {
					_ = checkMap[index]
				} else {
					checkMap[index] = struct{}{}
				}
				time.Sleep(time.Millisecond * 5)
				err = lock.Unlock()
				if err != nil {
					t.Log("un "+ty+" lock fail", err)
				}
			}(i, ty)
		}
		wait.Wait()
	}
}
