package limit

import (
	"context"
	"time"

	"golang.org/x/time/rate"

	"template/pkg/clock"
)


type RateLimiter interface {
    TryAccept() bool
	// Accept returns once a token becomes available.
	Accept()
	// Wait returns nil if a token is taken before the Context is done.
	Wait(ctx context.Context) error
}

// An injectable, mockable clock interface.
type Clock interface {
	clock.PassiveClock
	Sleep(time.Duration)
}

type stdRateLimiter struct {
	limiter *rate.Limiter
	qps     float32
	clock   Clock
}

func NewStdRateLimiter(qps float32, burst int, clock Clock) RateLimiter {
	return &stdRateLimiter{
		limiter: rate.NewLimiter(rate.Limit(qps), burst),
		qps:     qps,
		clock:   clock,
	}
}

// TryAccept returns true if a token is taken immediately. Otherwise,
// it returns false.
func (s *stdRateLimiter) TryAccept() bool {
	return s.limiter.AllowN(s.clock.Now(), 1)
}

// Accept returns once a token becomes available.
func (s *stdRateLimiter) Accept() {
	now := s.clock.Now()
	s.clock.Sleep(s.limiter.ReserveN(now, 1).DelayFrom(now))
}

// Wait returns nil if a token is taken before the Context is done.
func (s *stdRateLimiter) Wait(ctx context.Context) error {
	return s.limiter.Wait(ctx)
}