package job

import (
	"context"
)

// Job represents an interface to be implemented by structs which represent a 'job'
// to be performed.
type Job interface {
	// Description returns the description of the Job.
	Description() string

	// Key returns the unique key for the Job.
	Key() string

	// Execute is called by a SchedulerRuntime when the Trigger associated with this job fires.
	Execute(context.Context)
}
