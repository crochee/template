package quota

import (
	"context"
	"sync/atomic"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"

	"template/pkg/cache"
	"template/pkg/conc/pool"
	"template/pkg/env"
	"template/pkg/logger"
	"template/pkg/redis"
)

// Redis状态
const (
	redisStatusActive = 0 // 存活，可以对外提供服务
	redisStatusError  = 1 // 故障，无法对外提供服务
)

var ResourceQuota ResourceQuotaManager

type option func(map[string]resourceHandler, time.Duration)

// WithUsedQuotaHandler 注册需要配额管理的资源使用量和配额查询工具
func WithUsedQuotaHandler(resource string, handler UsedQuotaHandler) option {
	return func(m map[string]resourceHandler, expireTime time.Duration) {
		m[resource] = &resourceHandlerImpl{
			UsedQuotaHandler: handler,
			ResourceName:     resource,
			ExpireTime:       expireTime,
		}
	}
}

// InitResourceQuotaManager 初始化准备工作， 注册 资源更新对象;
// interval 查询Redis 状态的间隔时间，
// expireTime 资源使用量配额数据在Redis的过期时间，
// option 注册的资源使用量配额查询器
func InitResourceQuotaManager(interval time.Duration, expireTime time.Duration,
	newContext func(context.Context) context.Context, o ...option) {
	// 如果配额功能未开启，则不初始化
	if enable, _ := IsQuotaEnable(); !enable {
		return
	}

	if interval.Seconds() < 3 {
		interval = time.Second * 3
	}
	ResourceQuota = ResourceQuotaManager{
		newContext:    newContext,
		checkInterval: interval.Milliseconds(),
	}

	for _, f := range o {
		f(handlerMap, expireTime)
	}
}

// InitResourceQuotaData 刷新指定用户的 配额使用量数据
// account 需要初始化数据到 redis上的 account
func InitResourceQuotaData(ctx context.Context, accountsFn func(ctx context.Context) ([]string, error),
	routineFn func(inputCtx context.Context, receiver interface{ Execute(context.Context) }),
	setTraceID func(ctx context.Context, traceID string) context.Context) error {
	resources := make([]string, 0, len(handlerMap))
	for resource := range handlerMap {
		resources = append(resources, resource)
	}

	accounts, err := accountsFn(ctx)
	if err != nil {
		return err
	}

	quotaPool := pool.New().WithContext(ctx)
	quotaPool.Go(func(ctx context.Context) error {
		for _, act := range accounts {
			err := RefreshAccountUsedQuota(ctx, act, true, resources...)
			if err != nil {
				logger.FromContext(ctx).Err(err).Str("account", act).Strs("resources", resources).Msg("account init refresh account used quota fail")
			}
		}

		// 定期扫描 是否有配额溢出的数据,如果有，输出日志并推送到 echometer
		interval := viper.GetInt("quota.check_interval")
		if interval == 0 {
			interval = 24
		}
		t := time.NewTicker(time.Duration(interval) * time.Hour)
		task := CheckUsedQuota{accountsFn: accountsFn}
		for range t.C {
			ctx = setTraceID(ctx, "req-"+uuid.NewV4().String())
			routineFn(ctx, task)
		}
		t.Stop()
		return nil
	})
	return nil
}

type ResourceQuotaManager struct {
	newContext     func(context.Context) context.Context
	redisStatus    int32
	frontCheckTime int64
	checkInterval  int64
}

func PrepareOccupying(ctx context.Context, account string, requirement map[string]uint) (FinishQuota, error) {
	rq := ResourceQuota.GetPrepareOccupyingQuota(ctx)
	finish, err := rq.PrepareOccupying(ctx, account, requirement)
	if err != nil {
		logger.FromContext(ctx).Err(err).Interface("requirement", requirement).Str("account", account).Send()
		return nil, err
	}
	return finish, nil
}

// GetPrepareOccupyingQuota 根据 Redis 的状态，选择 是用 Redis 还是 Mysql 预占
func (r *ResourceQuotaManager) GetPrepareOccupyingQuota(ctx context.Context) PrepareOccupyingQuota {
	return &RedisResourceQuota{
		newContext: r.newContext,
	}
}

// Rollback 单独的 错误结束回调操作的功能，适用于  异步操作后资源创建失败，回滚配额
func Rollback(ctx context.Context, account string, requirement map[string]uint,
	resourceErrorTime time.Time) error {
	return ResourceQuota.rollback(ctx, account, requirement, &resourceErrorTime)
}

func (r *ResourceQuotaManager) rollback(ctx context.Context, account string, requirement map[string]uint,
	resourceErrorTime *time.Time) error {
	rollback := &RedisResourceQuota{
		account:           account,
		requirement:       requirement,
		isIgnore:          true,
		resourceErrorTime: resourceErrorTime,
	}
	return rollback.Rollback(ctx)
}

// CleanUsed 删除配额使用量，适用于资源删除时，对资源使用量进行扣减
func CleanUsed(ctx context.Context, account string, requirement map[string]uint) error {
	err := ResourceQuota.rollback(ctx, account, requirement, nil)
	if err != nil {
		logger.FromContext(ctx).Err(err).Str("account", account).Interface("requirement", requirement).
			Msg("fail to rollback resource quota")
		return err
	}
	return nil
}

// CheckRedisStatus 校验一下 Redis 是否可用
func (r *ResourceQuotaManager) CheckRedisStatus() {
	client := redis.NewRedisClient()
	if _, err := client.Ping(); err != nil {
		atomic.StoreInt32(&r.redisStatus, redisStatusError)
	} else {
		atomic.StoreInt32(&r.redisStatus, redisStatusActive)
	}
	atomic.StoreInt64(&r.frontCheckTime, time.Now().UnixMilli())
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

	// Paas、 Internal 用户不限制配额
	if len(accounts) > 0 {
		ok, err := cache.IsPaas(accounts[0])
		if err != nil || ok {
			return false, err
		}
		ok, err = cache.IsInternal(accounts[0])
		if err != nil || ok {
			return false, err
		}
	}

	return true, nil
}
