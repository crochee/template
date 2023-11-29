package job

import (
	"context"
)

// ScheduledJob wraps a scheduled Job with its metadata.
type ScheduledJob struct {
	Job     Job
	Trigger Trigger
}

// SchedulerRuntime represents a Job orchestrator.
// Schedulers are responsible for executing Jobs when their associated
// Triggers fire (when their scheduled time arrives).
type SchedulerRuntime interface {
	// Start starts the scheduler.
	Start(ctx context.Context) error

	// ScheduleJob schedules a job using a specified trigger.
	ScheduleJob(ctx context.Context, job Job, trigger Trigger) error

	// GetJobKeys returns the keys of all of the scheduled jobs.
	GetJobKeys() []string

	// GetScheduledJob returns the scheduled job with the specified key.
	GetScheduledJob(key string) (*ScheduledJob, error)

	// DeleteJob removes the job with the specified key from the SchedulerRuntime execution queue.
	DeleteJob(ctx context.Context, key string) error

	// Has Looks up an item under specified key.
	Has(key string) bool
}
