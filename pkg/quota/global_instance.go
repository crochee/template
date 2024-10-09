package quota

import (
	"context"
	"time"

	"github.com/spf13/viper"

	"template/pkg/cache"
	"template/pkg/env"
	"template/pkg/logger"
	"template/pkg/utils"
)

var resourceQuota *resourceQuotaManager

// InitResourceQuotaManager 初始化准备工作， 注册 资源更新对象;
// interval 查询Redis 状态的间隔时间，
// expireTime 资源使用量配额数据在Redis的过期时间，
// option 注册的资源使用量配额查询器
func InitResourceQuotaManager(o ...option) {
	resourceQuota = NewResourceQuotaManager(o...)
}

// PrepareOccupying 配额预占账户为关联id的情况入口
func PrepareOccupying(
	ctx context.Context,
	account string,
	requirement map[string]int,
) (FinishQuota, error) {
	// 处理入参
	params := make([]*Param, 0, len(requirement))
	for resource, num := range requirement {
		if num == 0 {
			continue
		}
		params = append(params, &Param{
			AssociatedID: account,
			Name:         resource,
			Num:          int64(num),
		})
	}
	finish, err := Mgr().Begin(ctx, params)
	if err != nil {
		logger.FromContext(ctx).
			Err(err).
			Interface("requirement", requirement).
			Str("account", account).
			Send()
		return nil, err
	}
	return finish, nil
}

func Mgr() *resourceQuotaManager {
	return resourceQuota
}

// Rollback 单独的 错误结束回调操作的功能，适用于  异步操作后资源创建失败，回滚配额
func Rollback(ctx context.Context, account string, requirement map[string]int,
	_ time.Time) error {
	// 处理入参
	params := make([]*paramWithStatus, 0, len(requirement))
	for resource, num := range requirement {
		if num == 0 {
			continue
		}
		// 标识已执行预占逻辑
		status := &utils.Status{}
		status.AddStatus(stateEvauate)
		params = append(params, &paramWithStatus{
			Param: &Param{
				AssociatedID: account,
				Name:         resource,
				Num:          int64(num),
			},
			Status: status,
		})
	}
	finisher, err := Mgr().getFinisherWithStatus(ctx, params)
	if err != nil {
		return err
	}
	return finisher.Rollback(ctx)
}

// Finally 成功解锁
func Finally(ctx context.Context, account string, requirement map[string]struct{}) error {
	// 处理入参
	params := make([]*paramWithStatus, 0, len(requirement))
	for resource := range requirement {
		// 标识已执行预占逻辑
		status := &utils.Status{}
		status.AddStatus(stateEvauate)
		params = append(params, &paramWithStatus{
			Param: &Param{
				AssociatedID: account,
				Name:         resource,
			},
			Status: status,
		})
	}
	finisher, err := Mgr().getFinisherWithStatus(ctx, params)
	if err != nil {
		return err
	}
	err = finisher.Finally(ctx)
	if err != nil {
		return err
	}
	return nil
}

// CleanUsed 删除配额使用量，适用于资源删除时，对资源使用量进行扣减
func CleanUsed(ctx context.Context, account string, requirement map[string]int) error {
	// 处理入参
	params := make([]*paramWithStatus, 0, len(requirement))
	for resource := range requirement {
		status := &utils.Status{}
		params = append(params, &paramWithStatus{
			Param: &Param{
				AssociatedID: account,
				Name:         resource,
			},
			Status: status,
		})
	}
	finisher, err := Mgr().getFinisherWithStatus(ctx, params)
	if err != nil {
		return err
	}
	return finisher.Rollback(ctx)
}

// IsQuotaEnable 用户配额功能是否启用
func IsQuotaEnable(accounts ...string) (bool, error) {
	// 私有云不启用用户配额
	if env.IsPrivate() {
		return false, nil
	}

	// 读取配置文件
	if enable := viper.GetBool("user_quota.enable"); !enable {
		return false, nil
	}

	if len(accounts) == 0 {
		return true, nil
	}
	firstAccount := accounts[0]
	if firstAccount == "" {
		return true, nil
	}
	// Paas、 Internal 用户不限制配额
	ok, err := cache.IsPaas(firstAccount)
	if err != nil || ok {
		return false, err
	}
	ok, err = cache.IsInternal(firstAccount)
	if err != nil || ok {
		return false, err
	}

	return true, nil
}
