package pool

import (
	"context"
)

// ResultErrorPool is a pool that executes tasks that return a generic result
// type and an error. Tasks are executed in the pool with Go(), then the
// results of the tasks are returned by Wait().
//
// The order of the results is not guaranteed to be the same as the order the
// tasks were submitted. If your use case requires consistent ordering,
// consider using the `stream` package or `Map` from the `iter` package.
//
// The configuration methods (With*) will panic if they are used after calling
// Go() for the first time.
type ResultErrorPool struct {
	errorPool      ErrorPool
	agg            resultAggregator
	collectErrored bool
}

// Go submits a task to the pool. If all goroutines in the pool
// are busy, a call to Go() will block until the task can be started.
func (p *ResultErrorPool) Go(f func() (interface{}, error)) {
	p.errorPool.Go(func() error {
		res, err := f()
		if err == nil || p.collectErrored {
			p.agg.add(res)
		}
		return err
	})
}

// Wait cleans up any spawned goroutines, propagating any panics and
// returning the results and any errors from tasks.
func (p *ResultErrorPool) Wait() ([]interface{}, error) {
	err := p.errorPool.Wait()
	return p.agg.results, err
}

// WithCollectErrored configures the pool to still collect the result of a task
// even if the task returned an error. By default, the result of tasks that errored
// are ignored and only the error is collected.
func (p *ResultErrorPool) WithCollectErrored() *ResultErrorPool {
	p.panicIfInitialized()
	p.collectErrored = true
	return p
}

// WithContext converts the pool to a ResultContextPool for tasks that should
// run under the same context, such that they each respect shared cancellation.
// For example, WithCancelOnError can be configured on the returned pool to
// signal that all goroutines should be cancelled upon the first error.
func (p *ResultErrorPool) WithContext(ctx context.Context) *ResultContextPool {
	p.panicIfInitialized()
	return &ResultContextPool{
		contextPool: *p.errorPool.WithContext(ctx),
	}
}

// WithFirstError configures the pool to only return the first error
// returned by a task. By default, Wait() will return a combined error.
func (p *ResultErrorPool) WithFirstError() *ResultErrorPool {
	p.panicIfInitialized()
	p.errorPool.WithFirstError()
	return p
}

// WithMaxGoroutines limits the number of goroutines in a pool.
// Defaults to unlimited. Panics if n < 1.
func (p *ResultErrorPool) WithMaxGoroutines(n int) *ResultErrorPool {
	p.panicIfInitialized()
	p.errorPool.WithMaxGoroutines(n)
	return p
}

func (p *ResultErrorPool) panicIfInitialized() {
	p.errorPool.panicIfInitialized()
}
