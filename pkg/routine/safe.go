package routine

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
)

type Pool struct {
	waitGroup   sync.WaitGroup
	ctx         context.Context
	recoverFunc func(ctx context.Context, r interface{})
	copyContext func(dst context.Context, src context.Context) context.Context
}

// NewPool creates a Pool.
func NewPool(ctx context.Context, opts ...Option) *Pool {
	opt := option{
		recoverFunc: defaultRecoverGoroutine,
		copyContext: defaultCopyContext,
	}
	for _, o := range opts {
		o(&opt)
	}
	return &Pool{
		ctx:         ctx,
		recoverFunc: opt.recoverFunc,
		copyContext: opt.copyContext,
	}
}

// Go starts a recoverable goroutine with a context.
func (p *Pool) Go(ctx context.Context, goroutine func(context.Context)) {
	p.waitGroup.Add(1)
	go func(ctx context.Context) {
		defer func() {
			if r := recover(); r != nil {
				p.recoverFunc(ctx, r)
			}
			p.waitGroup.Done()
		}()
		goroutine(ctx)
	}(p.copyContext(p.ctx, ctx))
}

// Wait Waits all started routines, waiting for their termination.
func (p *Pool) Wait() {
	p.waitGroup.Wait()
}

func defaultRecoverGoroutine(_ context.Context, err interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, "Error:%v\nStack: %s", err, debug.Stack())
}

func defaultCopyContext(_, src context.Context) context.Context {
	return src
}

var DefaultPool = NewPool(context.Background())

// Go
func Go(f func()) {
	GoWithContext(context.Background(), func(_ context.Context) {
		f()
	})
}

// GoWithContext
func GoWithContext(ctx context.Context, f func(context.Context)) {
	DefaultPool.Go(ctx, f)
}

func Wait() {
	DefaultPool.Wait()
}
