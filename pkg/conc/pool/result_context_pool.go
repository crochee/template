package pool

import (
	"context"
)

// ResultContextPool is a pool that runs tasks that take a context and return a
// result. The context passed to the task will be canceled if any of the tasks
// return an error, which makes its functionality different than just capturing
// a context with the task closure.
//
// The configuration methods (With*) will panic if they are used after calling
// Go() for the first time.
type ResultContextPool struct {
	contextPool    ContextPool
	agg            resultAggregator
	collectErrored bool
}

// Go submits a task to the pool. If all goroutines in the pool
// are busy, a call to Go() will block until the task can be started.
func (p *ResultContextPool) Go(f func(context.Context) (interface{}, error)) {
	p.contextPool.Go(func(ctx context.Context) error {
		res, err := f(ctx)
		if err == nil || p.collectErrored {
			p.agg.add(res)
		}
		return err
	})
}

// Wait cleans up all spawned goroutines, propagates any panics, and
// returns an error if any of the tasks errored.
func (p *ResultContextPool) Wait() ([]interface{}, error) {
	err := p.contextPool.Wait()
	return p.agg.results, err
}

// WithCollectErrored configures the pool to still collect the result of a task
// even if the task returned an error. By default, the result of tasks that errored
// are ignored and only the error is collected.
func (p *ResultContextPool) WithCollectErrored() *ResultContextPool {
	p.panicIfInitialized()
	p.collectErrored = true
	return p
}

// WithFirstError configures the pool to only return the first error
// returned by a task. By default, Wait() will return a combined error.
func (p *ResultContextPool) WithFirstError() *ResultContextPool {
	p.panicIfInitialized()
	p.contextPool.WithFirstError()
	return p
}

// WithCancelOnError configures the pool to cancel its context as soon as
// any task returns an error. By default, the pool's context is not
// canceled until the parent context is canceled.
func (p *ResultContextPool) WithCancelOnError() *ResultContextPool {
	p.panicIfInitialized()
	p.contextPool.WithCancelOnError()
	return p
}

// WithFailFast is an alias for the combination of WithFirstError and
// WithCancelOnError. By default, the errors from all tasks are returned and
// the pool's context is not canceled until the parent context is canceled.
func (p *ResultContextPool) WithFailFast() *ResultContextPool {
	p.panicIfInitialized()
	p.contextPool.WithFailFast()
	return p
}

// WithMaxGoroutines limits the number of goroutines in a pool.
// Defaults to unlimited. Panics if n < 1.
func (p *ResultContextPool) WithMaxGoroutines(n int) *ResultContextPool {
	p.panicIfInitialized()
	p.contextPool.WithMaxGoroutines(n)
	return p
}

func (p *ResultContextPool) panicIfInitialized() {
	p.contextPool.panicIfInitialized()
}
