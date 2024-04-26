package quota

import (
	"context"
	"fmt"

	"go.uber.org/multierr"

	"template/pkg/syncx"
	"template/pkg/utils"
)

type Param struct {
	// 关联的资源id，指配额的数据是挂载在指定id上的，例如单账户下的网络数量，则关联id为账户id,
	// 单网络下的子网数量，则关联id为网络id
	AssociatedID string
	// 当前需要占用配额的资源名称
	Name string
	// 当前需要占用配额的资源数量
	Num uint64
}

// 配额预占接口
type PrepareOccupyingQuota interface {
	Begin(ctx context.Context, params []*Param) (FinishQuota, error)
	Transaction(ctx context.Context, params []*Param, fc func(context.Context) error) error
}

type resourceQuotaManager struct {
	getAccountFn  func(context.Context) string
	handlers      map[string]UsedQuotaHandler
	finisherFn    func(UsedQuotaHandler, *Param, syncx.Locker, *utils.Status) (FinishQuota, error)
	lockFn        func(string) syncx.Locker
	isQuotaEnable func(accounts ...string) (bool, error)
	keyFn         func(*Param) string
}

type paramWithStatus struct {
	*Param
	*utils.Status
}

func (no *resourceQuotaManager) getFinisherWithStatus(
	ctx context.Context,
	params []*paramWithStatus,
) (FinishQuota, error) {
	// 合并参数
	paramMaps := make(map[string]*paramWithStatus)
	for _, param := range params {
		key := param.Name + param.AssociatedID
		if v, ok := paramMaps[key]; ok {
			v.Num += param.Num
			continue
		}
		paramMaps[key] = param
	}
	// 参数合并后，如果依然为空，则不需要进行配额校验
	if len(paramMaps) == 0 {
		return noopFinishQuota{}, nil
	}
	enable, err := no.isQuotaEnable(no.getAccountFn(ctx))
	if err != nil {
		return nil, err
	}
	// 配额功能未开启，则不进行配额校验
	if !enable {
		return noopFinishQuota{}, nil
	}
	// 构造过程
	finishers := make(Finishers, 0, len(params))
	for _, param := range paramMaps {
		handler, ok := no.handlers[param.Name]
		if !ok {
			return nil, fmt.Errorf("handler not found: %s", param.Name)
		}
		finisher, err := no.finisherFn(
			handler,
			param.Param,
			no.lockFn(no.keyFn(param.Param)),
			param.Status,
		)
		if err != nil {
			return nil, err
		}
		finishers = append(finishers, finisher)
	}
	return finishers, nil
}

func (no *resourceQuotaManager) getFinisher(
	ctx context.Context,
	params []*Param,
) (FinishQuota, error) {
	list := make([]*paramWithStatus, 0, len(params))
	for _, param := range params {
		list = append(list, &paramWithStatus{
			Param:  param,
			Status: &utils.Status{},
		})
	}
	return no.getFinisherWithStatus(ctx, list)
}

func (no *resourceQuotaManager) Transaction(
	ctx context.Context,
	params []*Param,
	fc func(context.Context) error,
) (err error) {
	panicked := true
	var finish FinishQuota
	if finish, err = no.Begin(ctx, params); err != nil {
		return
	}
	defer func() {
		// Make sure to rollback when panic, Block error or Commit error
		if panicked || err != nil {
			err = multierr.Append(err, finish.Rollback(ctx))
		}
	}()

	if err = fc(ctx); err == nil {
		panicked = false
		return finish.Finally(ctx)
	}

	panicked = false
	return
}

func (no *resourceQuotaManager) Begin(
	ctx context.Context,
	params []*Param,
) (FinishQuota, error) {
	finishers, err := no.getFinisher(ctx, params)
	if err != nil {
		return nil, err
	}
	// 评估过程
	if err := finishers.evauate(ctx); err != nil {
		return nil, err
	}
	return finishers, nil
}

type option func(*resourceQuotaManager)

// WithUsedQuotaHandler 注册需要配额管理的资源使用量和配额查询工具
func WithUsedQuotaHandler(resource string, handler UsedQuotaHandler) option {
	return func(r *resourceQuotaManager) {
		r.handlers[resource] = NewSingleFlightUsedQuotaHandler(resource, handler)
	}
}

func WithLockFn(lockFn func(string) syncx.Locker) option {
	return func(r *resourceQuotaManager) {
		r.lockFn = lockFn
	}
}

func WithFinisherFn(
	finisherFn func(UsedQuotaHandler, *Param, syncx.Locker, *utils.Status) (FinishQuota, error),
) option {
	return func(r *resourceQuotaManager) {
		r.finisherFn = finisherFn
	}
}

func WithIsQuotaEnable(
	isQuotaEnable func(accounts ...string) (bool, error),
) option {
	return func(r *resourceQuotaManager) {
		r.isQuotaEnable = isQuotaEnable
	}
}

func WithKeyFn(keyFn func(*Param) string) option {
	return func(r *resourceQuotaManager) {
		r.keyFn = keyFn
	}
}

func WithGetAccountsFn(getAccountFn func(context.Context) string) option {
	return func(r *resourceQuotaManager) {
		r.getAccountFn = getAccountFn
	}
}

// NewResourceQuotaManager 创建配额管理器,默认锁为标准库的内存互斥锁
func NewResourceQuotaManager(options ...option) *resourceQuotaManager {
	r := &resourceQuotaManager{
		getAccountFn: func(context.Context) string {
			return ""
		},
		handlers:   make(map[string]UsedQuotaHandler),
		finisherFn: NewFinisherFn,
		lockFn: func(string) syncx.Locker {
			return syncx.NewStdMutex()
		},
		isQuotaEnable: IsQuotaEnable,
		keyFn: func(param *Param) string {
			return fmt.Sprintf("quota_%s_%s", param.Name, param.AssociatedID)
		},
	}
	for _, option := range options {
		option(r)
	}
	return r
}
