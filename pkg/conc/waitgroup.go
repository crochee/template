package conc

import (
	"sync"

	"template/pkg/conc/panics"
)

// NewWaitGroup creates a new WaitGroup.
func NewWaitGroup() *WaitGroup {
	return &WaitGroup{}
}

// WaitGroup is the primary building block for scoped concurrency.
// Goroutings can be spawned in the WaitGroup with the Go(), and calling
// Wait() will ensure all spawned goroutines exists before continuing.
// Any panics in a child goroutine will be caught and propagated to the
// caller of Wait().
type WaitGroup struct {
	wg sync.WaitGroup
	pc panics.Catcher
}

// Go spawns a new gorouting in the WaitGroup.
func (h *WaitGroup) Go(f func()) {
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		h.pc.Try(f)
	}()
}

// Wait will block until all spawned goroutings with Go exit and will
// propagate any panics spawned in child goroutings.
func (h *WaitGroup) Wait() {
	h.wg.Wait()

	h.pc.Repanic()
}

// WaitAndRecover will block until all spawned goroutings with Go exit and
// will return a *panics.Recovered if any panics were raised in child gorouting.
func (h *WaitGroup) WaitAndRecover() *panics.Recovered {
	h.wg.Wait()

	return h.pc.Recovered()
}
