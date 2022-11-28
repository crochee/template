package task

import (
	"context"
	"math"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/crochee/devt/pkg/utils/v"
)

type Policy uint8

const (
	PolicyRetry Policy = 1 + iota
	PolicyRevert

	defaultAttempt  = 30
	defaultInterval = 60 * time.Second
)

type retryOption struct {
	attempts int
	interval time.Duration
	policy   Policy
}

type RetryOption func(*retryOption)

func WithAttempt(attempt int) RetryOption {
	return func(o *retryOption) {
		o.attempts = attempt
	}
}

func WithInterval(interval time.Duration) RetryOption {
	return func(o *retryOption) {
		o.interval = interval
	}
}

func WithPolicy(policy Policy) RetryOption {
	return func(o *retryOption) {
		o.policy = policy
	}
}

type retryTask struct {
	Task
	attempts int
	interval time.Duration
	policy   Policy
}

// NewRetryFunc 生成一个错误重试的任务 错误重试，共30次，每两次间隔60s
func NewRetryFunc(do func(context.Context, interface{}) error, opts ...Option) Task {
	return RetryTask(NewFunc(do, opts...))
}

// NewRetryFuncTask 生成一个带回滚的错误重试的任务 错误重试，共30次，每两次间隔60s
func NewRetryFuncTask(c, r func(context.Context, interface{}) error, opts ...Option) Task {
	return RetryTask(NewFuncTask(c, r, opts...))
}

func RetryTask(t Task, opts ...RetryOption) Task {
	o := &retryOption{
		attempts: defaultAttempt,
		interval: defaultInterval,
		policy:   PolicyRetry,
	}
	for _, opt := range opts {
		opt(o)
	}
	return &retryTask{
		Task:     t,
		attempts: o.attempts,
		interval: o.interval,
		policy:   o.policy,
	}
}

func (rt *retryTask) Commit(ctx context.Context, input interface{}, callbacks ...Callback) error {
	err := rt.Task.Commit(ctx, input, callbacks...)
	if err == nil {
		return nil
	}
	if rt.policy == PolicyRetry {
		var tempAttempts int
		backOff := rt.newBackOff() // 退避算法 保证时间间隔为指数级增长
		currentInterval := 0 * time.Millisecond
		timer := time.NewTimer(currentInterval)
		for {
			select {
			case <-timer.C:
				shouldRetry := tempAttempts < rt.attempts
				if !shouldRetry {
					timer.Stop()
					return err
				}
				if retryErr := rt.Task.Commit(ctx, input, callbacks...); retryErr == nil {
					shouldRetry = false
				} else {
					var permanent *backoff.PermanentError
					if errors.As(retryErr, &permanent) {
						err = multierr.Append(err, errors.WithMessagef(permanent.Err, "%d try", tempAttempts+1))
						shouldRetry = false
					}
					err = multierr.Append(err, errors.WithMessagef(retryErr, "%d try", tempAttempts+1))
				}
				if !shouldRetry {
					timer.Stop()
					return err
				}
				// 计算下一次
				currentInterval = backOff.NextBackOff()
				tempAttempts++
				// 定时器重置
				timer.Reset(currentInterval)
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			}
		}
	}
	return err
}

func (rt *retryTask) newBackOff() backoff.BackOff {
	if rt.attempts < 2 || rt.interval <= 0 {
		return &backoff.ZeroBackOff{}
	}

	b := backoff.NewExponentialBackOff()
	b.InitialInterval = rt.interval

	// calculate the multiplier for the given number of attempts
	// so that applying the multiplier for the given number of attempts will not exceed 2 times the initial interval
	// it allows to control the progression along the attempts
	b.Multiplier = math.Pow(v.Binary, 1/float64(rt.attempts-1))

	// according to docs, b.Reset() must be called before using
	b.Reset()
	return b
}
