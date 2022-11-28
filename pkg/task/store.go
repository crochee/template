package task

import (
	"sync"
	"sync/atomic"
	"time"

	uuid "github.com/satori/go.uuid"
	"go.uber.org/multierr"
)

type State string

const (
	Ready   State = "ready"
	Running State = "running"
	Success State = "success"
	Error   State = "error"
	Deleted State = "deleted"
)

type StoreInfo interface {
	ID() string
	Name() string
	Trigger() string
	State() State
	Description() string
	CreateTime() time.Time
	UpdateTime() time.Time
	Metadata() map[string]interface{}
	Error() error

	SetTrigger(string)
	SetName(string)
	SetState(State)
	SetDescription(string)
	SetMetadata(map[string]interface{})
	AddError(err error, states ...bool)
}

func DefaultTaskInfo(id string, f ...func() time.Time) StoreInfo {
	nowFunc := time.Now
	if len(f) > 0 {
		nowFunc = f[0]
	}
	if id == "" {
		id = uuid.NewV4().String()
	}
	d := &defaultTaskInfo{
		id:          id,
		nowFunc:     nowFunc,
		name:        atomic.Value{},
		trigger:     atomic.Value{},
		state:       atomic.Value{},
		description: atomic.Value{},
		meta:        atomic.Value{},
		createTime:  time.Time{},
		updateTime:  atomic.Value{},
	}
	now := d.nowFunc()
	d.createTime = now
	d.updateTime.Store(now)
	d.SetState(Ready)
	return d
}

type defaultTaskInfo struct {
	id          string
	nowFunc     func() time.Time
	name        atomic.Value
	trigger     atomic.Value
	state       atomic.Value
	description atomic.Value
	meta        atomic.Value
	createTime  time.Time
	updateTime  atomic.Value
	mutex       sync.RWMutex
	err         error
}

func (t *defaultTaskInfo) ID() string {
	return t.id
}

func (t *defaultTaskInfo) Name() string {
	v, _ := t.name.Load().(string)
	return v
}

func (t *defaultTaskInfo) Trigger() string {
	v, _ := t.trigger.Load().(string)
	return v
}

func (t *defaultTaskInfo) State() State {
	v, _ := t.state.Load().(State)
	return v
}

func (t *defaultTaskInfo) Description() string {
	v, _ := t.description.Load().(string)
	return v
}

func (t *defaultTaskInfo) CreateTime() time.Time {
	return t.createTime
}

func (t *defaultTaskInfo) UpdateTime() time.Time {
	v, _ := t.updateTime.Load().(time.Time)
	return v
}

func (t *defaultTaskInfo) Metadata() map[string]interface{} {
	v, _ := t.meta.Load().(map[string]interface{})
	return v
}

func (t *defaultTaskInfo) Error() error {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.err
}

func (t *defaultTaskInfo) SetTrigger(s string) {
	t.trigger.Store(s)
	t.updateTime.Store(t.nowFunc())
}

func (t *defaultTaskInfo) SetMetadata(m map[string]interface{}) {
	t.meta.Store(m)
	t.updateTime.Store(t.nowFunc())
}

func (t *defaultTaskInfo) SetName(name string) {
	t.name.Store(name)
	t.updateTime.Store(t.nowFunc())
}

func (t *defaultTaskInfo) SetState(state State) {
	t.state.Store(state)
	t.updateTime.Store(t.nowFunc())
}

func (t *defaultTaskInfo) SetDescription(desc string) {
	t.description.Store(desc)
	t.updateTime.Store(t.nowFunc())
}

func (t *defaultTaskInfo) setError(err error) {
	if err != nil {
		t.mutex.Lock()
		t.err = err
		t.mutex.Unlock()
		t.updateTime.Store(t.nowFunc())
	}
}

func (t *defaultTaskInfo) AddError(err error, states ...bool) {
	setState := true
	if len(states) > 0 {
		setState = states[0]
	}
	curErr := t.Error()
	if curErr != nil {
		if err != nil {
			if setState {
				t.SetState(Error)
			}
		}
		t.setError(multierr.Append(curErr, err))
		return
	}
	if setState {
		if err != nil {
			t.SetState(Error)
			t.setError(err)
			return
		}
		t.SetState(Success)
		return
	}
	if err != nil {
		t.setError(err)
	}
}
