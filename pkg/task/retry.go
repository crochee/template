package task

import (
	"context"
	"math"
	"time"

	"github.com/cenkalti/backoff/v4"
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

func WithTimeout(timeout time.Duration) RetryOption {
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
	backOff := backoff.WithContext(rt.newBackOff(), ctx)
	backOff.Reset()
	retryOperate := func() error {
		return rt.Task.Commit(ctx, input, callbacks...)
	}
	return backoff.Retry(retryOperate, backOff)
}

func (r *retryTask) newBackOff() backoff.BackOff {
	if r.timeout > 0 {
		// 常量间隔时间
		r.attempts = int(r.timeout / r.interval)
		return backoff.WithMaxRetries(backoff.NewConstantBackOff(r.interval), uint64(r.attempts))
	}
	if r.interval > 0 {
		b := backoff.NewExponentialBackOff()
		b.InitialInterval = r.interval
		b.MaxInterval = 120 * b.InitialInterval
		b.MaxElapsedTime = 15 * b.MaxInterval
		if r.attempts > 0 {
			b.Multiplier = math.Pow(2, 1/float64(r.attempts-1))
			return backoff.WithMaxRetries(b, uint64(r.attempts))
		}
		return b
	}
	b := &backoff.ZeroBackOff{}
	if r.attempts > 0 {
		return backoff.WithMaxRetries(b, uint64(r.attempts))
	}
	return b
}
