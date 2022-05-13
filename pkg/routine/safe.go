package routine

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
)

type Pool struct {
	waitGroup sync.WaitGroup
	ctx       context.Context
	option
}

// NewPool creates a Pool.
func NewPool(ctx context.Context, opts ...func(*option)) *Pool {
	p := &Pool{
		ctx:    ctx,
		option: option{recoverFunc: defaultRecoverGoroutine},
	}
	for _, opt := range opts {
		opt(&p.option)
	}
	return p
}

// Go starts a recoverable goroutine with a context.
func (p *Pool) Go(goroutine func(context.Context)) {
	p.waitGroup.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if p.recoverFunc != nil {
					p.recoverFunc(p.ctx, r)
				}
			}
			p.waitGroup.Done()
		}()
		goroutine(p.ctx)
	}()
}

// Wait Waits all started routines, waiting for their termination.
func (p *Pool) Wait() {
	p.waitGroup.Wait()
}

func defaultRecoverGoroutine(_ context.Context, err interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, "Error:%v\nStack: %s", err, debug.Stack())
}
