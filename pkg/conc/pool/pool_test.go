package pool

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func ExamplePool() {
	p := New().WithMaxGoroutines(3)
	for i := 0; i < 5; i++ {
		p.Go(func() {
			fmt.Println("conc")
		})
	}
	p.Wait()
	// Output:
	// conc
	// conc
	// conc
	// conc
	// conc
}

func TestPool(t *testing.T) {
	t.Parallel()

	t.Run("basic", func(t *testing.T) {
		t.Parallel()

		g := New()
		var completed int64
		for i := 0; i < 100; i++ {
			g.Go(func() {
				time.Sleep(1 * time.Millisecond)
				atomic.AddInt64(&completed, 1)
			})
		}
		g.Wait()
		require.Equal(t, atomic.LoadInt64(&completed), int64(100))
	})

	t.Run("panics on configuration after init", func(t *testing.T) {
		t.Run("before wait", func(t *testing.T) {
			t.Parallel()
			g := New()
			g.Go(func() {})
			require.Panics(t, func() { g.WithMaxGoroutines(10) })
		})

		t.Run("after wait", func(t *testing.T) {
			t.Parallel()
			g := New()
			g.Go(func() {})
			g.Wait()
			require.Panics(t, func() { g.WithMaxGoroutines(10) })
		})
	})

	t.Run("limit", func(t *testing.T) {
		t.Parallel()
		for _, maxConcurrent := range []int{1, 10, 100} {
			t.Run(strconv.Itoa(maxConcurrent), func(t *testing.T) {
				g := New().WithMaxGoroutines(maxConcurrent)

				var currentConcurrent int64
				var errCount int64
				taskCount := maxConcurrent * 10
				for i := 0; i < taskCount; i++ {
					g.Go(func() {
						cur := atomic.AddInt64(&currentConcurrent, 1)
						if cur > int64(maxConcurrent) {
							atomic.AddInt64(&errCount, 1)
						}
						time.Sleep(time.Millisecond)
						atomic.AddInt64(&currentConcurrent, -1)
					})
				}
				g.Wait()
				require.Equal(t, int64(0), atomic.LoadInt64(&errCount))
				require.Equal(t, int64(0), atomic.LoadInt64(&currentConcurrent))
			})
		}
	})

	t.Run("propagate panic", func(t *testing.T) {
		t.Parallel()
		g := New()
		for i := 0; i < 10; i++ {
			i := i
			g.Go(func() {
				if i == 5 {
					panic(i)
				}
			})
		}
		require.Panics(t, g.Wait)
	})

	t.Run("panics do not exhaust goroutines", func(t *testing.T) {
		t.Parallel()
		g := New().WithMaxGoroutines(2)
		for i := 0; i < 10; i++ {
			g.Go(func() {
				panic(42)
			})
		}
		require.Panics(t, g.Wait)
	})

	t.Run("panics on invalid WithMaxGoroutines", func(t *testing.T) {
		t.Parallel()
		require.Panics(t, func() { New().WithMaxGoroutines(0) })
	})

	t.Run("returns correct MaxGoroutines", func(t *testing.T) {
		t.Parallel()
		p := New().WithMaxGoroutines(42)
		require.Equal(t, 42, p.MaxGoroutines())
	})

	t.Run("is reusable", func(t *testing.T) {
		t.Parallel()
		var count int64
		p := New()
		for i := 0; i < 10; i++ {
			p.Go(func() {
				atomic.AddInt64(&count, 1)
			})
		}
		p.Wait()
		require.Equal(t, int64(10), atomic.LoadInt64(&count))
		for i := 0; i < 10; i++ {
			p.Go(func() {
				atomic.AddInt64(&count, 1)
			})
		}
		p.Wait()
		require.Equal(t, int64(20), atomic.LoadInt64(&count))
	})
}

func BenchmarkPool(b *testing.B) {
	b.Run("startup and teardown", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			p := New()
			p.Go(func() {})
			p.Wait()
		}
	})

	b.Run("per task", func(b *testing.B) {
		p := New()
		f := func() {}
		for i := 0; i < b.N; i++ {
			p.Go(f)
		}
		p.Wait()
	})
}
