package routine

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"

	"go.uber.org/multierr"
)

type ErrGroup struct {
	waitGroup sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	sem       chan struct{}
	errOnce   sync.Once
	err       error
}

// NewGroup starts a recoverable goroutine ErrGroup with a context.
func NewGroup(ctx context.Context, opts ...Option) *ErrGroup {
	newCtx, cancel := context.WithCancel(ctx)

	opt := option{}
	for _, o := range opts {
		o(&opt)
	}
	g := &ErrGroup{
		ctx:    newCtx,
		cancel: cancel,
	}

	if opt.limit > 0 {
		g.sem = make(chan struct{}, opt.limit)
	}
	return g
}

// Go starts a recoverable goroutine with a context.
func (e *ErrGroup) Go(goroutine func(context.Context) error) {
	if e.sem != nil {
		e.sem <- struct{}{}
	}
	e.waitGroup.Add(1)
	go func() {
		var err error
		defer func() {
			if r := recover(); r != nil {
				err = multierr.Append(err, fmt.Errorf("%v.Stack:%s", r, debug.Stack()))
			}
			if err != nil {
				e.errOnce.Do(func() {
					e.err = err
					e.cancel()
				})
			}
			if e.sem != nil {
				<-e.sem
			}
			e.waitGroup.Done()
		}()
		err = goroutine(e.ctx)
	}()
}

func (e *ErrGroup) Wait() error {
	e.waitGroup.Wait()
	e.errOnce.Do(e.cancel)
	return e.err
}
