package limit

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"

	"template/pkg/clock"
)

func TestNewRedisRateLimiter_TryAccept(t *testing.T) {
	rc := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			":7001",
			":7002",
			":7003",
			":7004",
			":7005",
			":7000"},
	})
	defer rc.Close()

	key := "so"
	rl := NewRedisRateLimiter(rc, 5, 5, key, clock.RealClock{})
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			assert.True(t, rl.TryAccept())
			wg.Done()
		}()
	}
	wg.Wait()
	assert.NoError(t, rc.Del(context.Background(), fmt.Sprintf("rate_limiter:{%s}", key)).Err())
}

func TestNewRedisRateLimiter_Accept(t *testing.T) {
	rc := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			":7001",
			":7002",
			":7003",
			":7004",
			":7005",
			":7000"},
	})
	defer rc.Close()

	key := "so"
	rl := NewRedisRateLimiter(rc, 1, 1, key, clock.RealClock{})
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			rl.Accept()
			wg.Done()
		}()
	}
	wg.Wait()
	assert.NoError(t, rc.Del(context.Background(), fmt.Sprintf("rate_limiter:{%s}", key)).Err())
}

func TestNewRedisRateLimiter_Wait(t *testing.T) {
	rc := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			":7001",
			":7002",
			":7003",
			":7004",
			":7005",
			":7000"},
	})
	defer rc.Close()

	key := "so"
	rl := NewRedisRateLimiter(rc, 1, 1, key, clock.RealClock{})
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			verr := rl.Wait(ctx)
			val, err := rc.HGetAll(context.Background(), fmt.Sprintf("rate_limiter:{%s}", key)).
				Result()
			assert.NoError(t, err)
			t.Logf("err: %+v, val: %+v", verr, val)

			wg.Done()
		}()
	}
	wg.Wait()
	assert.NoError(t, rc.Del(context.Background(), fmt.Sprintf("rate_limiter:{%s}", key)).Err())
}
