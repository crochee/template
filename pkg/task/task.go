package task

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strconv"

	"go.uber.org/multierr"

	"github.com/crochee/devt/pkg/idx"
)

type taskOption struct {
	storeInfo         StoreInfo
	commitCallbacks   []Callback
	rollbackCallbacks []Callback
	tasks             []Task
}

type Option func(*taskOption)

func WithStoreInfo(storeInfo StoreInfo) Option {
	return func(option *taskOption) {
		option.storeInfo = storeInfo
	}
}

func WithCommitCallbacks(callbacks ...Callback) Option {
	return func(option *taskOption) {
		option.commitCallbacks = callbacks
	}
}

func WithRollbackCallbacks(callbacks ...Callback) Option {
	return func(option *taskOption) {
		option.rollbackCallbacks = callbacks
	}
}

func WithTasks(tasks ...Task) Option {
	return func(option *taskOption) {
		option.tasks = tasks
	}
}

// Task is library's minimum unit
type Task interface {
	Commit(ctx context.Context, input interface{}, callbacks ...Callback) error
	Rollback(ctx context.Context, input interface{}, callbacks ...Callback) error
}

func NewFunc(do func(context.Context, interface{}) error, opts ...Option) Task {
	return NewFuncTask(do, nil, opts...)
}

func NewFuncTask(c, r func(context.Context, interface{}) error, opts ...Option) Task {
	id, err := idx.NextID()
	o := &taskOption{
		storeInfo:         DefaultTaskInfo(strconv.FormatUint(id, 10)),
		commitCallbacks:   make([]Callback, 0),
		rollbackCallbacks: make([]Callback, 0),
	}
	for _, opt := range opts {
		opt(o)
	}
	f := &funcTask{
		StoreInfo:         o.storeInfo,
		commitCallbacks:   o.commitCallbacks,
		rollbackCallbacks: o.rollbackCallbacks,
		c:                 c,
		r:                 r,
	}
	funcName := runtime.FuncForPC(reflect.ValueOf(c).Pointer()).Name()
	f.SetName(funcName)

	if err != nil {
		f.AddError(err, true)
		return f
	}
	f.SetDescription("func task")
	return f
}

type funcTask struct {
	StoreInfo
	commitCallbacks   []Callback
	rollbackCallbacks []Callback
	c                 func(context.Context, interface{}) error
	r                 func(context.Context, interface{}) error
}

func (f *funcTask) Commit(ctx context.Context, input interface{}, callbacks ...Callback) error {
	err := f.Error()
	if err == nil {
		f.SetState(Running)
		fErr := f.c(ctx, input)
		f.AddError(fErr, true)
		err = fErr
	}

	callbacks = append(f.commitCallbacks, callbacks...)
	for _, callback := range callbacks {
		callback.Trigger(ctx, f, input, err)
	}
	return err
}

func (f *funcTask) Rollback(ctx context.Context, input interface{}, callbacks ...Callback) error {
	if f.r == nil {
		return nil
	}

	err := f.r(ctx, input)

	f.AddError(err, true)

	callbacks = append(f.rollbackCallbacks, callbacks...)
	for _, callback := range callbacks {
		callback.Trigger(ctx, f, input, err)
	}
	return err
}

type recoverTask struct {
	Task
}

func SafeTask(t Task) Task {
	return &recoverTask{
		Task: t,
	}
}

func (rt *recoverTask) Commit(ctx context.Context, input interface{}, callbacks ...Callback) (err error) {
	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			err = multierr.Append(err, fmt.Errorf("[Recover] found:%v,trace:\n%s", r, buf))
		}
	}()
	err = rt.Task.Commit(ctx, input, callbacks...)
	return
}

func (rt *recoverTask) Rollback(ctx context.Context, input interface{}, callbacks ...Callback) (err error) {
	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			err = multierr.Append(err, fmt.Errorf("[Recover] found:%v,trace:\n%s", r, buf))
		}
	}()
	err = rt.Task.Rollback(ctx, input, callbacks...)
	return
}
