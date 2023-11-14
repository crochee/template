package quota

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"template/pkg/env"
	"template/pkg/idx"
	"template/pkg/logger"
	wsserver "template/pkg/msg/server"
	"template/pkg/quota/errcode"
	"template/pkg/redis"
)

type PrepareOccupyingQuota interface {
	PrepareOccupying(ctx context.Context, account string, requirement map[string]uint) (FinishQuota, error)
}

type FinishQuota interface {
	Finally(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type defaultFinishQuota struct {
}

func (d defaultFinishQuota) Finally(ctx context.Context) error {
	return nil
}

func (d defaultFinishQuota) Rollback(ctx context.Context) error {
	return nil
}

func CreateDefaultFinishQuota() FinishQuota {
	return defaultFinishQuota{}
}

type RedisResourceQuota struct {
	newContext        func(context.Context) context.Context
	account           string          // 预占账户
	requirement       map[string]uint // 需要预占的资源类型和数量
	isOccupy          bool            // 是否已占用
	isIgnore          bool            // 忽略预占
	resourceErrorTime *time.Time
	userId            uint64
}

// PrepareOccupying 预占用户的资源配额
func (r *RedisResourceQuota) PrepareOccupying(ctx context.Context, account string, requirement map[string]uint) (FinishQuota, error) {
	// 如果配额功能没有开启，则不执行配额预占
	if enable, err := IsQuotaEnable(account); err != nil || !enable {
		if !env.IsPrivate() {
			logger.FromContext(ctx).Warn().Str("account", account).Msg("this account has not ben enabled quota during PrepareOccupying")
		}
		return defaultFinishQuota{}, err
	}

	defer TimeCost(ctx, time.Now(), "resource prepareOccupying "+account)

	r.requirement = requirement
	r.account = account
	r.userId, _ = idx.NextID()
	resourceNames := make([]string, 0, len(requirement))
	keys := make([]string, 0, len(requirement))
	values := make([]interface{}, 0, len(requirement))
	for resource, num := range requirement {
		if num == 0 {
			continue
		}
		resourceNames = append(resourceNames, resource)
		keys = append(keys, fmt.Sprintf(ResourceKey, r.account, resource))
		values = append(values, num)
	}
	if len(keys) == 0 {
		return defaultFinishQuota{}, nil
	}

	// 获取 用户 读锁
	rlock := CreateReadLock(ctx, fmt.Sprintf(LockKey, r.account), r.userId, LockLeaseTime)
	if err := rlock.Lock(); err != nil {
		logger.FromContext(ctx).Err(err).Str("account", account).Msg("fail to lock read locker")
		return nil, err
	}

	// 预占之前 先 刷新用户配额
	err := RefreshAccountQuota(ctx, account, resourceNames...)
	if err != nil {
		logger.FromContext(ctx).Err(err).Str("account", account).Msg("fail to refresh account quota")
		return nil, err
	}
	err = r.operatePrepareOccupying(ctx, keys, resourceNames, values)
	if err != nil {
		_ = rlock.Unlock()
		if errcode.ErrCodeResourceQuotaInvalid.Is(err) {
			// 预占发现数据异常， 刷新配额数据
			err = RefreshAccountUsedQuota(ctx, r.account, false, resourceNames...)
			if err != nil {
				return nil, err
			}
			// 刷新配额后 重试一次
			err = rlock.Lock()
			if err != nil {
				return nil, err
			}
			err = r.operatePrepareOccupying(ctx, keys, resourceNames, values)
			if err != nil {
				_ = rlock.Unlock()
			} else {
				return r, nil
			}
		}

		usedQuota := getAccountUsedQuota(ctx, account, resourceNames...)
		logger.FromContext(ctx).Err(err).Str("account", account).Interface("usedQuota", usedQuota).Msg("prepare occupying fail")
		// 记录配额不足时的真实数据
		wsserver.Errorf(ctx, err, "prepare occupying fail,account: %s, used quota data:%+v", account, usedQuota)
		return nil, err
	}
	return r, nil
}

func getAccountUsedQuota(ctx context.Context, account string, resources ...string) []UsedQuotaRes {
	var res []UsedQuotaRes
	for _, resource := range resources {
		handler, ok := handlerMap[resource]
		if !ok {
			continue
		}

		var errMsg string
		used, err := handler.QueryUsed(ctx, account)
		if err != nil {
			errMsg = "err: " + err.Error() + ", "
		}
		quota, err := handler.QueryQuota(ctx, account)
		if err != nil {
			errMsg = "err: " + err.Error() + ", "
		}
		usedCache, quotaCache, updateTime, errCache := GetAccountCacheResourceUsedQuota(ctx, account, resource)
		if errCache != nil {
			errMsg = errMsg + "cache err: " + errCache.Error()
		}

		res = append(res, UsedQuotaRes{
			ResourceName:     resource,
			RealQuota:        quota,
			RealUsed:         used,
			CacheQuota:       quotaCache,
			CacheUsed:        usedCache,
			RefreshCacheTime: updateTime,
			IsError:          errMsg != "",
			ErrorMsg:         errMsg,
		})
	}
	return res
}

func (r *RedisResourceQuota) operatePrepareOccupying(ctx context.Context, keys, resourceNames []string, values []interface{}) error {
	client := redis.NewRedisClient()
	res, err := client.Eval(PreAppropriationScript, keys, values...)
	if err != nil {
		return errcode.ErrCodeQuotaServerDisable.WithResult(err)
	}
	result, _ := res.(string)
	if result == OperationSuccess {
		r.isOccupy = true
		return nil
	} else if strings.HasPrefix(result, OperationFailure) {
		// 配额不足错误
		var resourceName string
		indexArray := strings.Split(result, "-")
		if len(indexArray) > 1 {
			index, err := strconv.Atoi(indexArray[1])
			if err == nil {
				resourceName = resourceNames[index-1]
			}
		}
		return errcode.ErrCodeResourceQuotaInsufficient.WithResult(resourceName)
	} else if strings.HasPrefix(result, ResourceDataInvalid) {
		// 数据不正确错误
		return errcode.ErrCodeResourceQuotaInvalid
	}
	logger.FromContext(ctx).Err(err).Interface("result", res).Msg("invalid preApplication result!")
	return errcode.ErrCodeResourceQuotaScriptResInvalid.WithResult(res)
}

func (r *RedisResourceQuota) Finally(ctx context.Context) error {
	// 如果配额功能没有开启，则不执行配额预占
	if enable, err := IsQuotaEnable(r.account); err != nil || !enable {
		return err
	}

	ctx = r.newContext(ctx)

	// redis 需要释放 读锁
	rlock := CreateReadLock(ctx, fmt.Sprintf(LockKey, r.account), r.userId, LockLeaseTime)
	return rlock.Unlock()
}

// Rollback 回滚配额
func (r *RedisResourceQuota) Rollback(ctx context.Context) error {
	// 如果配额功能没有开启，则不执行配额预占
	if enable, err := IsQuotaEnable(r.account); err != nil || !enable {
		if !env.IsPrivate() {
			logger.FromContext(ctx).Warn().Str("account", r.account).Msg("this account has not ben enabled quota during Rollback")
		}
		return err
	}

	ctx = r.newContext(ctx)

	defer func() {
		if err := recover(); err != nil {
			logger.FromContext(ctx).Error().Interface("panic", err).Str("account", r.account).Msg("rollback panic")
		}
		TimeCost(ctx, time.Now(), "resource rollback "+r.account)
	}()

	if !r.isOccupy && !r.isIgnore {
		return nil
	}
	client := redis.NewRedisClient()

	resourceNames := make([]string, 0, len(r.requirement))
	keys := make([]string, 0, len(r.requirement))
	values := make([]interface{}, 0, len(r.requirement))

	if !r.isOccupy && r.resourceErrorTime != nil && !r.resourceErrorTime.IsZero() {
		for resource := range r.requirement {
			resourceNames = append(resourceNames, resource)
		}
		// 校验 回滚数据时间是否无效
		ups, err := client.HMGet(fmt.Sprintf(ResourceUpdateTimeKey, r.account), resourceNames...)
		if err != nil {
			logger.FromContext(ctx).Err(err).Str("account", r.account).Msg("fail to hmget ResourceUpdateTimeKey")
			return err
		}
		for i, up := range ups {
			upd, _ := up.(string)
			updateTime, _ := strconv.ParseInt(upd, 10, 64)
			// TODO 实用性，待验证
			if r.resourceErrorTime.UnixMilli() < updateTime {
				// 如果 资源出错后，redis 进行了更新，
				// 或者 redis 在 回滚之前重启。那redis 里的数据就是最新的，不需要再进行回滚了
				delete(r.requirement, resourceNames[i])
			}
		}
		resourceNames = make([]string, 0, len(r.requirement))
	}

	for resource, num := range r.requirement {
		if num == 0 {
			continue
		}
		keys = append(keys, fmt.Sprintf(ResourceKey, r.account, resource))
		values = append(values, num)
		resourceNames = append(resourceNames, resource)
	}
	if len(keys) == 0 {
		return nil
	}

	res, err := client.Eval(RollbackScript, keys, values...)
	if err != nil {
		logger.FromContext(ctx).Err(err).Interface("keys", keys).Interface("values", values).Msg("fail to eval RollbackScript")
		return err
	}
	result, _ := res.(string)
	if result == OperationSuccess {
		return nil
	} else if strings.HasPrefix(result, OperationFailure) {
		// 配额数据异常
		var resourceName string
		indexArray := strings.Split(result, "-")
		if len(indexArray) > 1 {
			index, err := strconv.Atoi(indexArray[1])
			if err == nil {
				resourceName = resourceNames[index-1]
			}
		}
		return errcode.ErrCodeResourceQuotaAbnormal.WithResult(resourceName)
	} else if strings.HasPrefix(result, ResourceDataInvalid) {
		// 数据不正确错误
		_ = RefreshAccountUsedQuota(ctx, r.account, false, resourceNames...)
		return nil
	}
	logger.FromContext(ctx).Err(err).Interface("result", res).Msg("invalid rollback result!")
	return errcode.ErrCodeResourceQuotaScriptResInvalid.WithResult(res)
}

type MysqlResourceQuota struct {
	account     string          // 预占账户
	requirement map[string]uint // 需要预占的资源类型和数量
	IsOccupy    bool            // 是否已占用
}

func (m *MysqlResourceQuota) PrepareOccupying(ctx context.Context, account string, requirement map[string]uint) (FinishQuota, error) {
	// 如果配额功能没有开启，则不执行配额预占
	if enable, err := IsQuotaEnable(account); err != nil || !enable {
		return defaultFinishQuota{}, err
	}

	return nil, nil
}

func (m *MysqlResourceQuota) Finally(ctx context.Context) error {
	// 如果配额功能没有开启，则不执行配额预占
	if enable, err := IsQuotaEnable(m.account); err != nil || !enable {
		return err
	}

	// mysql 预占结束后，资源创建后，需要释放 预占锁
	return nil
}

func (m MysqlResourceQuota) Rollback(ctx context.Context) error {
	// 如果配额功能没有开启，则不执行配额预占
	if enable, err := IsQuotaEnable(m.account); err != nil || !enable {
		return err
	}

	// mysql 预占结束后，资源创建出错 需要释放 预占锁
	return nil
}

func TimeCost(c context.Context, start time.Time, info string) {
	tc := time.Since(start)
	logger.FromContext(c).Info().Int64("cost", tc.Milliseconds()).Msg(info)
}
