package redis

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/spf13/viper"
)

const (
	redisStatActive = 1
	redisStatError  = 2
)

var redisState uint32

func checkRedis() {
	client := NewRedisClient()

	// 连续检测指定次数，只要超过指定次数都失败，才认为Redis不可用
	allFailed := false
	maxFailureCount := viper.GetInt("redis.allow_continuous_failure")
	if maxFailureCount == 0 {
		// 默认连续检测3次
		maxFailureCount = 3
	}
	for i := 0; i < maxFailureCount; i++ {
		if _, err := client.Ping(); err != nil {
			allFailed = true
			time.Sleep(time.Millisecond * 500)
			continue
		}
		allFailed = false
		atomic.StoreUint32(&redisState, redisStatActive)
		break
	}
	if allFailed {
		atomic.StoreUint32(&redisState, redisStatError)
	}
}

// CheckRedisInterval 通过定时任务周期性检测Redis服务状态
func CheckRedisInterval(ctx context.Context) {
	interval := viper.GetInt("redis.check_interval")
	if interval == 0 {
		// 默认5秒钟
		interval = 5
	}
	ticker := time.NewTicker(time.Second * time.Duration(interval))
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			go checkRedis()
		}
	}
}

// IsRedisActive Redis服务是否可用
func IsRedisActive() bool {
	return atomic.LoadUint32(&redisState) == redisStatActive
}
