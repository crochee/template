package limit

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"golang.org/x/time/rate"

	"template/pkg/code"
)

type redisRateLimiter struct {
	burst         int
	qps           float32
	store         redis.Scripter
	clock         Clock
	key           string
	reserveScript string
	cancelScript  string
}

func NewRedisRateLimiter(
	store redis.Scripter,
	qps float32,
	burst int,
	key string,
	clock Clock,
) RateLimiter {
	return &redisRateLimiter{
		burst: burst,
		qps:   qps,
		store: store,
		clock: clock,
		key:   fmt.Sprintf("rate_limiter:{%s}", key),
		reserveScript: `
-- KEYS[1] 锁名
-- ARGV[1] 请求速率qps
-- ARGV[2] 突发最大请求数
-- ARGV[3] 当前时间戳
-- ARGV[4] 请求占用的令牌数
-- ARGV[5] 最大等待时间,单位s
-- 返回值:
--  -1表示请求被拒绝（无论是因为请求的令牌数超过了突发最大令牌数，还是等待时间超过了最大等待时间）
--  计算出允许的时间戳。

-- 将参数转换为本地变量
local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])
local max_wait = tonumber(ARGV[5])

-- 单次请求的最大令牌数不能大于突发最大令牌数
if requested > burst then
    return -1
end
-- 使用HMGET一次性获取tokens和last
local values = redis.call("HMGET", KEYS[1], "tokens", "last")
local tokens = tonumber(values[1]) or 0
local last = tonumber(values[2]) or 0

-- 计算时间差和新的令牌数
local delta = math.max(0, now-last) * rate
local tokens = math.min(burst, tokens + delta)

-- 计算等待时间
local wait_sec = tokens >= requested and 0 or math.floor((requested - tokens) / rate)

-- 检查是否允许请求
if wait_sec > max_wait   then
    return - wait_sec
end

-- 更新令牌数
tokens = tokens - requested
local wait_time = now + wait_sec
redis.call("HMSET", KEYS[1], "tokens", tokens, "last", now, "last_event", wait_time)
-- 刷新过期时间为突增到最大值的2倍
local ttl = math.floor((burst - tokens) / rate) * 2
if ttl > 0 then
    redis.call("EXPIRE", KEYS[1], ttl)
end

-- 计算返回值
return wait_time
`,
		cancelScript: `
-- KEYS[1] 锁名
-- ARGV[1] 请求速率qps
-- ARGV[2] 突发最大请求数
-- ARGV[3] 当前时间戳
-- ARGV[4] 请求占用的令牌数
-- 返回值 0

-- 将参数转换为本地变量
local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])

if redis.call("EXISTS", KEYS[1]) == 0 then
    return 0
end

-- 使用HMGET一次性获取tokens和last
local values = redis.call("HMGET", KEYS[1], "tokens", "last", "last_event")
local tokens = tonumber(values[1]) or 0
local last = tonumber(values[2]) or 0
local last_event = tonumber(values[3]) or 0

if tokens == 0 || last_event < now then
    return 0
end

-- 计算时间差和新的令牌数
local delta = math.max(0, now-last) * rate
local tokens = math.min(burst, tokens + delta + requested)

-- 计算等待时间
if last == last_event then
    local prev_event = math.floor(requested  / rate) + last
    if prev_event >= now then
        last_event = prev_event
    end
end

redis.call("HMSET", KEYS[1], "tokens", tokens, "last", now, "last_event", last_event)
-- 刷新过期时间为突增到最大值的2倍
local ttl = math.floor((burst - tokens) / rate) * 2
if ttl > 0 then
    redis.call("EXPIRE", KEYS[1], ttl)
end
return 0
`,
	}
}

// TryAccept returns true if a token is taken immediately. Otherwise,
// it returns false.
func (r *redisRateLimiter) TryAccept() bool {
	_, err := r.reserveN(context.Background(), r.clock.Now(), 1, 0)
	return err == nil
}

// Accept returns once a token becomes available.
func (r *redisRateLimiter) Accept() {
	now := r.clock.Now()
	exp, err := r.reserveN(context.Background(), now, 1, rate.InfDuration)
	if err != nil {
		exp = int64(rate.InfDuration.Seconds())
	}
	r.clock.Sleep(time.Duration(exp-now.Unix()) * time.Second)
}

// Wait returns nil if a token is taken before the Context is done.
func (r *redisRateLimiter) Wait(ctx context.Context) error {
	// The test code calls lim.wait with a fake timer generator.
	// This is the real timer generator.
	newTimer := func(d time.Duration) (<-chan time.Time, func() bool, func()) {
		timer := time.NewTimer(d)
		return timer.C, timer.Stop, func() {}
	}
	err := r.wait(ctx, 1, r.clock.Now(), newTimer)
	if err != nil {
		return code.ErrTooManyRequests.WithResult(err.Error())
	}
	return nil
}

func (r *redisRateLimiter) reserveN(
	ctx context.Context,
	now time.Time,
	n int,
	maxFutureReserve time.Duration,
) (int64, error) {
	max_wait := int(maxFutureReserve.Seconds())
	resp, err := r.store.Eval(ctx,
		r.reserveScript,
		[]string{
			r.key,
		},
		int(r.qps),
		r.burst,
		now.Unix(),
		n,
		max_wait,
	).Int64()
	if err != nil {
		return 0, errors.WithStack(err)
	}
	if resp == -1 {
		return 0, errors.Errorf("rate: Wait(n=%d) exceeds limiter's burst %d", n, r.burst)
	}
	if resp < -1 {
		return 0, errors.Errorf(
			"rate: Wait(n=%d,t=%ds) over than limiter's max_future_reserve %ds",
			n,
			-resp,
			max_wait,
		)
	}
	return resp, nil
}

func (r *redisRateLimiter) wait(
	ctx context.Context,
	n int,
	t time.Time,
	newTimer func(d time.Duration) (<-chan time.Time, func() bool, func()),
) error {
	if n > r.burst {
		return errors.Errorf("rate: Wait(n=%d) exceeds limiter's burst %d", n, r.burst)
	}
	// Check if ctx is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	// Determine wait limit
	waitLimit := rate.InfDuration
	if deadline, ok := ctx.Deadline(); ok {
		waitLimit = deadline.Sub(t)
	}
	// Reserve
	exp, err := r.reserveN(ctx, t, n, waitLimit)
	if err != nil {
		return err
	}
	// allow
	if exp == 0 {
		return nil
	}
	// Wait if necessary
	delay := time.Duration(exp-t.Unix()) * time.Second
	if delay == 0 {
		return nil
	}
	ch, stop, advance := newTimer(delay)
	defer stop()
	advance() // only has an effect when testing
	select {
	case <-ch:
		// We can proceed.
		return nil
	case <-ctx.Done():
		// Context was canceled before we could proceed.  Cancel the
		// reservation, which may permit other events to proceed sooner.
		err = r.cancelAt(context.Background(), r.clock.Now(), n)
		return multierr.Append(ctx.Err(), err)
	}
}

func (r *redisRateLimiter) cancelAt(ctx context.Context, now time.Time, n int) error {
	err := r.store.Eval(ctx,
		r.cancelScript,
		[]string{
			r.key,
		},
		int(r.qps),
		r.burst,
		now.Unix(),
		n,
	).Err()
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
