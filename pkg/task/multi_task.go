package task

import (
	"context"
	"strconv"

	cmap "github.com/orcaman/concurrent-map"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/multierr"

	"template/pkg/routine"
)

type pipelineTask struct {
	StoreInfo
	commitCallbacks   []Callback
	rollbackCallbacks []Callback
	tasks             []Task
	cur               int
}

func PipelineTask(opts ...Option) Task {
	uid := uuid.NewV1()
	o := &taskOption{
		storeInfo:         DefaultTaskInfo(uid.String()),
		commitCallbacks:   make([]Callback, 0),
		rollbackCallbacks: make([]Callback, 0),
		tasks:             make([]Task, 0),
	}
	for _, opt := range opts {
		opt(o)
	}
	pt := &pipelineTask{
		StoreInfo:         o.storeInfo,
		commitCallbacks:   o.commitCallbacks,
		rollbackCallbacks: o.rollbackCallbacks,
		tasks:             o.tasks,
		cur:               0,
	}
	if o.storeInfo.Name() == "" {
		pt.SetName("pipeline-task-" + o.storeInfo.ID())
	}
	return pt
}

func (s *pipelineTask) Commit(ctx context.Context, input interface{}, callbacks ...Callback) error {
	s.SetState(Running)

	var err error
	for index, task := range s.tasks {
		if err = task.Commit(ctx, input, callbacks...); err != nil {
			s.cur = index
			break
		}
	}
	s.AddError(err, true)
	for _, callback := range s.commitCallbacks {
		callback.Trigger(ctx, s, input, err)
	}
	return err
}

func (s *pipelineTask) Rollback(ctx context.Context, input interface{}, callbacks ...Callback) error {
	var err error
	for i := s.cur; i >= 0; i-- {
		err = multierr.Append(err, s.tasks[i].Rollback(ctx, input, callbacks...))
	}
	s.AddError(err, true)
	for _, callback := range s.rollbackCallbacks {
		callback.Trigger(ctx, s, input, err)
	}
	return err
}

func ParallelTask(opts ...Option) Task {
	uid := uuid.NewV1()
	o := &taskOption{
		storeInfo:         DefaultTaskInfo(uid.String()),
		commitCallbacks:   make([]Callback, 0),
		rollbackCallbacks: make([]Callback, 0),
		tasks:             make([]Task, 0),
	}
	for _, opt := range opts {
		opt(o)
	}
	pt := &parallelTask{
		StoreInfo:         o.storeInfo,
		commitCallbacks:   o.commitCallbacks,
		rollbackCallbacks: o.rollbackCallbacks,
		tasks:             o.tasks,
		safeMap:           cmap.New(),
	}
	if o.storeInfo.Name() == "" {
		pt.SetName("parallel-task-" + o.storeInfo.ID())
	}
	return pt
}

type parallelTask struct {
	StoreInfo
	commitCallbacks   []Callback
	rollbackCallbacks []Callback
	tasks             []Task
	safeMap           cmap.ConcurrentMap
}

func (p *parallelTask) Commit(ctx context.Context, input interface{}, callbacks ...Callback) error {
	p.SetState(Running)
	g := routine.NewGroup(ctx)
	for index, task := range p.tasks {
		tempTask := task
		tempIndex := strconv.Itoa(index)
		g.Go(func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				err := tempTask.Commit(ctx, input, callbacks...)
				p.safeMap.Set(tempIndex, err == nil)
				return err
			}
		})
	}
	err := g.Wait()
	p.AddError(err, true)
	for _, callback := range p.commitCallbacks {
		callback.Trigger(ctx, p, input, err)
	}
	return err
}

func (p *parallelTask) Rollback(ctx context.Context, input interface{}, callbacks ...Callback) error {
	g := routine.NewGroup(ctx)
	for index, task := range p.tasks {
		v, ok := p.safeMap.Get(strconv.Itoa(index))
		if !ok {
			continue
		}
		if !v.(bool) {
			tempTask := task
			g.Go(func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					return tempTask.Rollback(ctx, input, callbacks...)
				}
			})
		}
	}
	err := g.Wait()
	p.AddError(err, true)
	for _, callback := range p.rollbackCallbacks {
		callback.Trigger(ctx, p, input, err)
	}
	return err
}
