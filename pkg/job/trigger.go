package job

import (
	"errors"
	"fmt"
	"time"
)

// Trigger represents the mechanism by which Jobs are scheduled.
type Trigger interface {
	// NextFireTime returns the next time at which the Trigger is scheduled to fire.
	NextFireTime(prev int64) (int64, error)

	// Description returns the description of the Trigger.
	Description() string
}

// constantDelayTrigger implements the quartz.Trigger interface; uses a fixed interval.
type constantDelayTrigger struct {
	Interval time.Duration
}

// Verify constantDelayTrigger satisfies the Trigger interface.
var _ Trigger = (*constantDelayTrigger)(nil)

// Every returns a new constantDelayTrigger using the given interval.
func Every(interval time.Duration) *constantDelayTrigger {
	return &constantDelayTrigger{
		Interval: interval,
	}
}

// NextFireTime returns the next time at which the constantDelayTrigger is scheduled to fire.
func (t *constantDelayTrigger) NextFireTime(prev int64) (int64, error) {
	next := prev + t.Interval.Nanoseconds()
	return next, nil
}

// Description returns the description of the trigger.
func (t *constantDelayTrigger) Description() string {
	return fmt.Sprintf("constantDelayTrigger with interval: %d", t.Interval)
}

// runOnceTrigger implements the quartz.Trigger interface.
// This type of Trigger can only be fired once and will delay immediately.
type runOnceTrigger struct {
	Delay   time.Duration
	expired bool
}

// Verify runOnceTrigger satisfies the Trigger interface.
var _ Trigger = (*runOnceTrigger)(nil)

// RunOnce returns a new runOnceTrigger with the given delay time.
func RunOnce(delay time.Duration) *runOnceTrigger {
	return &runOnceTrigger{
		Delay:   delay,
		expired: false,
	}
}

var ErrSkipScheduleJob = errors.New("skip scheduleJob")

// NextFireTime returns the next time at which the runOnceTrigger is scheduled to fire.
// Sets exprired to true afterwards.
func (ot *runOnceTrigger) NextFireTime(prev int64) (int64, error) {
	if !ot.expired {
		next := prev + ot.Delay.Nanoseconds()
		ot.expired = true
		return next, nil
	}
	return 0, ErrSkipScheduleJob
}

// Description returns the description of the trigger.
func (ot *runOnceTrigger) Description() string {
	status := "valid"
	if ot.expired {
		status = "expired"
	}

	return fmt.Sprintf("runOnceTrigger (%s).", status)
}

// Verify runAtTrigger satisfies the Trigger interface.
var _ Trigger = (*runAtTrigger)(nil)

// RunAt returns a new runOnceTrigger with the given delay time.
func RunAt(at int64) *runAtTrigger {
	return &runAtTrigger{
		at:      at,
		expired: false,
	}
}

// runAtTrigger implements the quartz.Trigger interface.
// This type of Trigger can only be fired once and will delay immediately.
type runAtTrigger struct {
	at      int64
	expired bool
}

// NextFireTime returns the next time at which the runAtTrigger is scheduled to fire.
// Sets exprired to true afterwards.
func (ot *runAtTrigger) NextFireTime(prev int64) (int64, error) {
	if !ot.expired {
		ot.expired = true
		if prev > ot.at {
			return prev, nil
		}
		return ot.at, nil
	}
	return 0, ErrSkipScheduleJob
}

// Description returns the description of the trigger.
func (ot *runAtTrigger) Description() string {
	status := "valid"
	if ot.expired {
		status = "expired"
	}

	return fmt.Sprintf("runAtTrigger (%s).", status)
}
