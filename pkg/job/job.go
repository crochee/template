package job

import (
	"context"
	"reflect"
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

type funcJob struct {
	key string
	f   func(context.Context)
}

func (f *funcJob) Description() string {
	return f.key
}

func (f *funcJob) Key() string {
	return f.key
}

func (f *funcJob) Execute(ctx context.Context) {
	f.f(ctx)
}

func NewFuncJob(f func(context.Context)) Job {
	return &funcJob{
		key: reflect.TypeOf(f).String(),
		f:   f,
	}
}
