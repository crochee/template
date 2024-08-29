package clock

import "time"

type PassiveClock interface {
	Now() time.Time
	Since(time.Time) time.Duration
}

type RealClock struct{}

func (RealClock) Now() time.Time {
    return time.Now()
}

func (RealClock) Since(ts time.Time) time.Duration {
    return time.Since(ts)
}

func (RealClock) Sleep(d time.Duration) {
    time.Sleep(d)
}
