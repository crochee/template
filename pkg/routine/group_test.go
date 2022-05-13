package routine

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGroup(t *testing.T) {
	testList := []struct {
		name     string
		input    []func(ctx context.Context) error
		expected bool
	}{
		{
			name: "error",
			input: []func(ctx context.Context) error{
				func(ctx context.Context) error {
					return nil
				},
				func(ctx context.Context) error {
					return errors.New("error")
				},
			},
			expected: true,
		},
		{
			name: "panic",
			input: []func(ctx context.Context) error{
				func(ctx context.Context) error {
					return nil
				},
				func(ctx context.Context) error {
					panic("panic")
					return nil
				},
			},
			expected: true,
		},
		{
			name: "nil",
			input: []func(ctx context.Context) error{
				func(ctx context.Context) error {
					return nil
				},
				func(ctx context.Context) error {
					return nil
				},
			},
			expected: false,
		},
	}
	for _, data := range testList {
		t.Run(data.name, func(t *testing.T) {
			g := NewGroup(context.Background())
			for _, f := range data.input {
				g.Go(f)
			}
			if data.expected {
				assert.Error(t, g.Wait())
			} else {
				assert.NoError(t, g.Wait())
			}
		})
	}
}

func BenchmarkNewGroup(b *testing.B) {
	b.ReportAllocs()
	g := NewGroup(context.Background())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Go(func(ctx context.Context) error {
			return errors.New("error")
		})
		g.Go(func(ctx context.Context) error {
			panic("panic")
			return nil
		})
		g.Go(func(ctx context.Context) error {
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		b.Log(err)
	}
}
