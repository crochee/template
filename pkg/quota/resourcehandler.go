package quota

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	redis8 "github.com/go-redis/redis/v8"

	"template/pkg/idx"
	"template/pkg/logger"
	wsserver "template/pkg/msg/server"
	"template/pkg/redis"
)

var handlerMap = map[string]resourceHandler{}

type resourceHandler interface {
	UsedQuotaHandler
	refresh(ctx context.Context, account string) error
	refreshQuota(ctx context.Context, account string) error
}

type resourceHandlerImpl struct {
	UsedQuotaHandler
	ExpireTime   time.Duration
	ResourceName string
}

func (e *resourceHandlerImpl) refresh(ctx context.Context, account string) error {
	used, err := e.QueryUsed(ctx, account)
	if err != nil {
		return err
	}
	quota, err := e.QueryQuota(ctx, account)
	if err != nil {
		return err
	}
	client := redis.NewRedisClient()
	pl := client.Pipeline()
	key := fmt.Sprintf(ResourceKey, account, e.ResourceName)
	pl.HSet(ctx, key, ResourceUsedField, used)
	pl.HSet(ctx, key, ResourceQuotaField, quota)
	pl.Expire(ctx, key, e.ExpireTime)
	uKey := fmt.Sprintf(ResourceUpdateTimeKey, account)
	pl.HSet(ctx, uKey, e.ResourceName, time.Now().UnixMilli())
	pl.Expire(ctx, uKey, e.ExpireTime)
	pl.ZAdd(ctx, QuotaAccountZSetKey, &redis8.Z{
		Score:  float64(time.Now().UnixMilli()),
		Member: account,
	})
	pl.Expire(ctx, QuotaAccountZSetKey, QuotaAccountExprieTime)
	_, err = pl.Exec(ctx)
	if err != nil {
		logger.FromContext(ctx).Err(err).Str("account", account).Msg("account refresh used quota push redis fail")
	}
	_ = pl.Close()
	return err
}

func (e *resourceHandlerImpl) refreshQuota(ctx context.Context, account string) error {
	quota, err := e.QueryQuota(ctx, account)
	if err != nil {
		return err
	}
	client := redis.NewRedisClient()
	pl := client.Pipeline()
	key := fmt.Sprintf(ResourceKey, account, e.ResourceName)
	pl.HSet(ctx, key, ResourceQuotaField, quota)
	pl.Expire(ctx, key, e.ExpireTime)
	_, err = pl.Exec(ctx)
	if err != nil {
		logger.FromContext(ctx).Err(err).Str("account", account).Msg("account refresh quota push redis fail")
	}
	_ = pl.Close()
	return err
}

type defaultResourceHandler struct {
}

func (d defaultResourceHandler) QueryUsed(ctx context.Context, account string) (used int, err error) {
	return 0, nil
}

func (d defaultResourceHandler) QueryQuota(ctx context.Context, account string) (quota int, err error) {
	return 0, nil
}

func (d defaultResourceHandler) refresh(ctx context.Context, account string) error {
	return nil
}

func (d defaultResourceHandler) refreshQuota(ctx context.Context, account string) error {
	return nil
}

type UsedQuotaHandler interface {
	// QueryUsed 查询用户的配额数据和使用量是数据
	QueryUsed(ctx context.Context, account string) (used int, err error)
	// QueryQuota 查询用户的配额数据和使用量是数据
	QueryQuota(ctx context.Context, account string) (quota int, err error)
}

func GetRefreshHandler(ctx context.Context, resourceType string) resourceHandler {
	handler, ok := handlerMap[resourceType]
	if !ok {
		logger.FromContext(ctx).Info().Str("resource", resourceType).Msg("Invalid resource type")
		return defaultResourceHandler{}
	}
	return handler
}

func RangeRefreshHandler(fn func(resource string, handler resourceHandler)) {
	for resource, handler := range handlerMap {
		fn(resource, handler)
	}
}

func RefreshAccountUsedQuota(ctx context.Context, account string, force bool, resources ...string) error {
	userId, _ := idx.NextID()
	// 获取用户 写锁
	wlock := CreateWriteLock(ctx, fmt.Sprintf(LockKey, account), userId, LockLeaseTime)
	err := wlock.Lock()
	if err != nil {
		return err
	}
	defer wlock.Unlock()

	needRefreshResource := make([]string, 0, len(resources))
	if !force {
		// 判断用户资源是否更新过，如果是就 不更新
		client := redis.NewRedisClient()

		ups, err := client.HMGet(fmt.Sprintf(ResourceUpdateTimeKey, account), resources...)
		if err != nil {
			return err
		}
		for i, up := range ups {
			upd, _ := up.(string)
			updateTime, _ := strconv.ParseInt(upd, 10, 64)
			if time.Now().UnixMilli()-updateTime > 1000 {
				needRefreshResource = append(needRefreshResource, resources[i])
			}
		}
	} else {
		needRefreshResource = resources
	}

	for _, resource := range needRefreshResource {
		rq := GetRefreshHandler(ctx, resource)
		err := rq.refresh(ctx, account)
		if err != nil {
			wsserver.Errorf(ctx, err, "init account used quota fail,account: %s ,resource %s", account, resource)
			logger.FromContext(ctx).Err(err).Str("account", account).Str("resource", resource).Msg("refresh account used quota resource fail")
		}
	}
	return nil
}

func RefreshAccountQuota(ctx context.Context, account string, resources ...string) error {
	for _, resource := range resources {
		rq := GetRefreshHandler(ctx, resource)
		err := rq.refreshQuota(ctx, account)
		if err != nil {
			wsserver.Errorf(ctx, err, "init account quota fail,account: %s ,resource %s", account, resource)
			logger.FromContext(ctx).Err(err).Str("account", account).Str("resource", resource).Msg("refresh account quota resource fail")
		}
	}
	return nil
}

func GetAccountCacheResourceUsedQuota(_ context.Context, account, resource string) (used int, quota int, updateTime int64, err error) {
	client := redis.NewRedisClient()
	key := fmt.Sprintf(ResourceKey, account, resource)
	uKey := fmt.Sprintf(ResourceUpdateTimeKey, account)
	result, err := client.HMGet(key, ResourceUsedField, ResourceQuotaField)
	if err != nil {
		return 0, 0, 0, err
	}
	if len(result) != 2 {
		return 0, 0, 0, fmt.Errorf("redis resource value invalid:%v", result)
	}
	u, ok := result[0].(string)
	if !ok {
		return 0, 0, 0, fmt.Errorf("redis resource used value invalid:%v", result[0])
	}
	q, ok := result[1].(string)
	if !ok {
		return 0, 0, 0, fmt.Errorf("redis resource quota value invalid:%v", result[1])
	}
	used, err = strconv.Atoi(u)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("redis resource used value invalid:%v", result[0])
	}
	quota, err = strconv.Atoi(q)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("redis resource quota value invalid:%v", result[1])
	}
	updateTimeStr, err := client.HGet(uKey, resource)
	if err != nil {
		return used, quota, 0, err
	}
	updateTime, err = strconv.ParseInt(updateTimeStr, 10, 64)
	if err != nil {
		return used, quota, 0, fmt.Errorf("resource update time valid,value:%s", updateTimeStr)
	}
	return
}

type CheckUsedQuota struct {
	accountsFn func(ctx context.Context) ([]string, error)
}

func (r CheckUsedQuota) Execute(ctx context.Context) {
	if r.accountsFn == nil {
		return
	}
	accounts, err := r.accountsFn(ctx)
	if err != nil {
		logger.FromContext(ctx).Err(err).Msg("account list fail")
		return
	}
	for _, account := range accounts {
		RangeRefreshHandler(func(resource string, handler resourceHandler) {
			used, err := handler.QueryUsed(ctx, account)
			if err != nil {
				logger.FromContext(ctx).Err(err).Str("resource", resource).Str("account", account).
					Msg("query used fail")
				return
			}
			quota, err := handler.QueryQuota(ctx, account)
			if err != nil {
				logger.FromContext(ctx).Err(err).Str("resource", resource).Str("account", account).
					Msg("query quota fail")
				return
			}
			if used > quota {
				logger.FromContext(ctx).Info().Str("account", account).Str("resource", resource).
					Int("used", used).Int("quota", quota).
					Msg("account used exceed quota")
				wsserver.Errorf(ctx, errors.New("account used exceed quota"), "account: %s, resource:%s,used:%d,quota:%d", account, resource, used, quota)
			}
		})
	}
}
