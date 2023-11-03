package task

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"template/pkg/utils/v"
)

const (
	defaultAttempt  = 30
	defaultInterval = 60 * time.Second
)

type retryOption struct {
	attempts int
	interval time.Duration
	timeout  time.Duration
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

func WithTimeOut(timeout time.Duration) RetryOption {
	return func(o *retryOption) {
		o.timeout = timeout
	}
}

func NotRetryError(err error) error {
	if err == nil {
		return nil
	}
	return &backoff.PermanentError{
		Err: err,
	}
}

type retryTask struct {
	Task
	attempts int
	interval time.Duration
	timeout  time.Duration
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
	}
	for _, opt := range opts {
		opt(o)
	}
	return &retryTask{
		Task:     t,
		attempts: o.attempts,
		interval: o.interval,
		timeout:  o.timeout,
	}
}

func (rt *retryTask) Commit(ctx context.Context, input interface{}, callbacks ...Callback) error {
	err := rt.Task.Commit(ctx, input, callbacks...)
	if err == nil {
		return nil
	}

	var tempAttempts int
	backOff := rt.newBackOff() // 退避算法 保证时间间隔为指数级增长
	currentInterval := backOff.NextBackOff()
	timer := time.NewTimer(currentInterval)
	for {
		select {
		case <-timer.C:
			shouldRetry := tempAttempts < rt.attempts
			if !shouldRetry {
				timer.Stop()
				return err
			}
			retryErr := rt.Task.Commit(ctx, input, callbacks...)
			if retryErr == nil {
				timer.Stop()
				return nil
			}
			var permanent *backoff.PermanentError
			if errors.As(retryErr, &permanent) {
				err = multierr.Append(err, fmt.Errorf("%w %d try", permanent.Err, tempAttempts+1))
				shouldRetry = false
			} else {
				err = multierr.Append(err, fmt.Errorf("%w %d try", retryErr, tempAttempts+1))
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

func (rt *retryTask) newBackOff() backoff.BackOff {
	if rt.timeout > 0 {
		// 常量间隔时间
		rt.attempts = int(rt.timeout / rt.interval)
		return backoff.NewConstantBackOff(rt.interval)
	}
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
