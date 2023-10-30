package pool

import (
	"fmt"
	"sort"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func ExampleResultPool() {
	p := NewWithResults()
	for i := 0; i < 10; i++ {
		i := i
		p.Go(func() interface{} {
			return i * 2
		})
	}
	res := p.Wait()
	// Result order is nondeterministic, so sort them first
	result := make([]int, 0, len(res))
	for _, r := range res {
		result = append(result, r.(int))
	}
	sort.Ints(result)
	fmt.Println(result)

	// Output:
	// [0 2 4 6 8 10 12 14 16 18]
}

func TestResultGroup(t *testing.T) {
	t.Parallel()

	t.Run("panics on configuration after init", func(t *testing.T) {
		t.Run("before wait", func(t *testing.T) {
			t.Parallel()
			g := NewWithResults()
			g.Go(func() interface{} { return 0 })
			require.Panics(t, func() { g.WithMaxGoroutines(10) })
		})

		t.Run("after wait", func(t *testing.T) {
			t.Parallel()
			g := NewWithResults()
			g.Go(func() interface{} { return 0 })
			_ = g.Wait()
			require.Panics(t, func() { g.WithMaxGoroutines(10) })
		})
	})

	t.Run("basic", func(t *testing.T) {
		t.Parallel()
		g := NewWithResults()
		expected := []int{}
		for i := 0; i < 100; i++ {
			i := i
			expected = append(expected, i)
			g.Go(func() interface{} {
				return i
			})
		}
		res := g.Wait()
		result := make([]int, 0, len(res))
		for _, r := range res {
			result = append(result, r.(int))
		}
		sort.Ints(result)
		require.Equal(t, expected, result)
	})

	t.Run("limit", func(t *testing.T) {
		t.Parallel()
		for _, maxGoroutines := range []int{1, 10, 100} {
			t.Run(strconv.Itoa(maxGoroutines), func(t *testing.T) {
				g := NewWithResults().WithMaxGoroutines(maxGoroutines)

				var currentConcurrent int64
				var errCount int64
				taskCount := maxGoroutines * 10
				expected := make([]int, taskCount)
				for i := 0; i < taskCount; i++ {
					i := i
					expected[i] = i
					g.Go(func() interface{} {
						cur := atomic.AddInt64(&currentConcurrent, 1)
						if cur > int64(maxGoroutines) {
							atomic.AddInt64(&errCount, 1)
						}
						time.Sleep(time.Millisecond)
						atomic.AddInt64(&currentConcurrent, -1)
						return i
					})
				}
				res := g.Wait()
				result := make([]int, 0, len(res))
				for _, r := range res {
					result = append(result, r.(int))
				}
				sort.Ints(result)
				require.Equal(t, expected, result)
				require.Equal(t, int64(0), atomic.LoadInt64(&errCount))
				require.Equal(t, int64(0), atomic.LoadInt64(&currentConcurrent))
			})
		}
	})
}
