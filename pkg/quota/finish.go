package quota

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"template/pkg/code"
	"template/pkg/syncx"
	"template/pkg/utils"
)

var (
	// 配额相关错误
	ErrResourceQuotaInsufficient     = code.Froze("500-11004400", "资源配额不足")
	ErrResourceQuotaInvalid          = code.Froze("500-11004401", "资源配额数据无效，请重试")
	ErrResourceQuotaAbnormal         = code.Froze("500-11004402", "资源配额数据异常，需要重新刷新用户配额数据")
	ErrResourceQuotaScriptResInvalid = code.Froze("500-11004403", "资源配额脚本执行结果无效")
	ErrQuotaServerDisable            = code.Froze("500-11004404", "配额服务暂时不可用，请稍后再试")
	ErrWaitLockTimeout               = code.Froze("500-11004405", "等待锁超时")
)

const (
	stateEvauate = 1 << iota
	stateFinally
)

type FinishQuota interface {
	// 同步数据
	sync(ctx context.Context) error
	// 评估配额的过程
	evauate(ctx context.Context) error
	Finally(ctx context.Context) error
	Rollback(ctx context.Context) error
}

func NewFinisherFn(handler UsedQuotaHandler, param *Param,
	lock syncx.Locker) (FinishQuota, error) {
	return &noneCacheFinishQuotaFinisher{
		lock:    lock,
		state:   &utils.Status{},
		param:   param,
		handler: handler,
	}, nil
}

type noneCacheFinishQuotaFinisher struct {
	lock    syncx.Locker
	state   *utils.Status
	param   *Param
	handler UsedQuotaHandler
}

func (no *noneCacheFinishQuotaFinisher) sync(ctx context.Context) (err error) {
	return
}

func (no *noneCacheFinishQuotaFinisher) evauate(ctx context.Context) (err error) {
	panicked := true
	if err = no.lock.Lock(); err != nil {
		return
	}
	no.state.AddStatus(stateEvauate)
	defer func() {
		if panicked || err != nil {
			no.lock.Unlock()
		}
	}()

	var quota int
	if quota, err = no.handler.QueryQuota(ctx, no.param.AssociatedID); err != nil {
		panicked = false
		err = errors.WithStack(ErrQuotaServerDisable.WithResult(err.Error()))
		return
	}
	var used int
	if used, err = no.handler.QueryUsed(ctx, no.param.AssociatedID); err != nil {
		panicked = false
		return
	}
	if math.MaxUint64-no.param.Num < uint64(used) {
		panicked = false
		err = errors.WithStack(
			ErrResourceQuotaInvalid.WithResult(
				fmt.Sprintf("used:%d quota:%d num:%d,data is overflow", used, quota, no.param.Num),
			),
		)
		return
	}
	if uint64(used)+no.param.Num > uint64(quota) {
		panicked = false
		err = errors.WithStack(
			ErrResourceQuotaInsufficient.WithResult(
				fmt.Sprintf("used:%d quota:%d num:%d", used, quota, no.param.Num),
			),
		)
		return
	}
	panicked = false
	return
}

func (no *noneCacheFinishQuotaFinisher) Finally(ctx context.Context) error {
	// 已经执行过终态的逻辑则跳过
	if no.state.HasStatus(stateFinally) {
		return nil
	}
	no.state.AddStatus(stateFinally)
	// 没有执行过预占逻辑则跳过
	if no.state.NotHasStatus(stateEvauate) {
		return nil
	}
	return no.lock.Unlock()
}

func (no *noneCacheFinishQuotaFinisher) Rollback(ctx context.Context) error {
	// 已经执行过终态的逻辑则跳过
	if no.state.HasStatus(stateFinally) {
		return nil
	}
	no.state.AddStatus(stateFinally)
	// 没有执行过预占逻辑则跳过
	if no.state.NotHasStatus(stateEvauate) {
		return nil
	}
	return no.lock.Unlock()
}

type Finishers []FinishQuota

func (fi Finishers) sync(ctx context.Context) error {
	var err error
	for i := len(fi) - 1; i >= 0; i-- {
		err = multierr.Append(err, fi[i].sync(ctx))
	}
	return err
}

func (fi Finishers) evauate(ctx context.Context) error {
	readyFinishers := make(Finishers, 0, len(fi))
	for _, finisher := range fi {
		if err := finisher.evauate(ctx); err != nil {
			return multierr.Append(err, readyFinishers.Rollback(ctx))
		}
		readyFinishers = append(readyFinishers, finisher)
	}
	return nil
}

func (fi Finishers) Finally(ctx context.Context) error {
	var err error
	for i := len(fi) - 1; i >= 0; i-- {
		err = multierr.Append(err, fi[i].Finally(ctx))
	}
	return err
}

func (fi Finishers) Rollback(ctx context.Context) error {
	var err error
	for i := len(fi) - 1; i >= 0; i-- {
		err = multierr.Append(err, fi[i].Rollback(ctx))
	}
	return err
}

func CreateDefaultFinishQuota() FinishQuota {
	return noopFinishQuota{}
}

type noopFinishQuota struct {
}

func (no noopFinishQuota) sync(ctx context.Context) error {
	return nil
}

// 评估配额的过程
func (no noopFinishQuota) evauate(ctx context.Context) error {
	return nil
}

func (no noopFinishQuota) Finally(ctx context.Context) error {
	return nil
}

func (no noopFinishQuota) Rollback(ctx context.Context) error {
	return nil
}

func NewRedisFinishQuota(
	handler UsedQuotaHandler,
	param *Param,
	lock syncx.Locker,
	cli *redis.ClusterClient,
	expire time.Duration,
) FinishQuota {
	return &redisFinishQuota{
		lock:    lock,
		state:   &utils.Status{},
		param:   param,
		handler: handler,
		cli:     cli,
		expire:  expire,
		rollbackScript: `
            	-- KEYS[1] 锁名
            	-- ARGV[1] 占用的资源数
		        local used = redis.call('HGET', KEYS[1], 'used')
				if tonumber(used) == nil then
					return 'Invalid'
				end
				if tonumber(used) < tonumber(ARGV[1]) then
					return 'Fail'
				end
		        redis.call('HINCRBY', KEYS[1], 'used', -tonumber(ARGV[1]))
				if redis.call('TTL', KEYS[1]) == -1 then
					redis.call('DEL', KEYS[1])
				end
		        return 'OK'
		        `,
		preAppropriationScript: `
            	-- KEYS[1] 锁名
            	-- ARGV[1] 占用的资源数
            	-- ARGV[2] 配额数据
		        local used = redis.call('HGET', KEYS[1], 'used')
				if tonumber(used) == nil then
					return 'Invalid'
				end
				if tonumber(used) + tonumber(ARGV[1]) > tonumber(ARGV[2]) then
					return 'Fail'
				end
		        redis.call('HINCRBY', KEYS[1], 'used', tonumber(ARGV[1]))
				if redis.call('TTL', KEYS[1]) == -1 then
					redis.call('DEL', KEYS[1])
				end
		        return 'OK'
		        `}
}

type redisFinishQuota struct {
	lock                   syncx.Locker
	state                  *utils.Status
	param                  *Param
	handler                UsedQuotaHandler
	cli                    *redis.ClusterClient
	expire                 time.Duration
	rollbackScript         string
	preAppropriationScript string
}

func (re *redisFinishQuota) resourceKey(param *Param) string {
	return fmt.Sprintf("dcs:resource:{%s}:%s", param.AssociatedID, param.Name)
}

func (re *redisFinishQuota) sync(ctx context.Context) error {
	// 修正错误数据
	used, err := re.handler.QueryUsed(ctx, re.param.AssociatedID)
	if err != nil {
		return errors.WithStack(err)
	}
	// 由于命令过多，开启pipeline执行
	pl := re.cli.Pipeline()
	// 操作配额和使用量数据
	resourceKey := re.resourceKey(re.param)
	pl.HSet(ctx, resourceKey, "used", used)
	pl.Expire(ctx, resourceKey, re.expire)
	_, err = pl.Exec(ctx)
	_ = pl.Close()
	return err
}

// 评估配额的过程
func (re *redisFinishQuota) evauate(ctx context.Context) (err error) {
	panicked := true
	if err = re.lock.Lock(); err != nil {
		return
	}
	re.state.SetStatus(stateEvauate)
	defer func() {
		if panicked || err != nil {
			re.lock.Unlock()
		}
	}()
	// 获取配额数据
	quota, err := re.handler.QueryQuota(ctx, re.param.AssociatedID)
	if err != nil {
		return errors.WithStack(err)
	}
	// 尝试预占配额
	var retry bool
	if retry, err = re.preHandle(ctx, quota, re.sync); err != nil {
		panicked = false
		return
	}
	if retry {
		// 发现有数据刷新则重新预占
		if _, err = re.preHandle(ctx, quota, func(ctx context.Context) error {
			return errors.WithStack(ErrResourceQuotaAbnormal.WithResult("refresh used,but failed"))
		}); err != nil {
			panicked = false
			return
		}
	}
	panicked = false
	return
}

func (re *redisFinishQuota) preHandle(
	ctx context.Context,
	quota int,
	handleInvalid func(ctx context.Context) error,
) (bool, error) {
	resp, err := re.cli.Eval(ctx, re.preAppropriationScript,
		[]string{re.resourceKey(re.param)}, re.param.Num, quota).Result()
	if err != nil {
		return false, errors.WithStack(ErrQuotaServerDisable.WithResult(err.Error()))
	}
	result, ok := resp.(string)
	if !ok {
		return false, errors.WithStack(
			ErrQuotaServerDisable.WithResult(fmt.Sprintf("invalid response:%v", resp)),
		)
	}
	switch result {
	case "Fail":
		return false, errors.WithStack(
			ErrResourceQuotaInsufficient.WithResult(
				fmt.Sprintf("num:%d", re.param.Num),
			),
		)
	case "Invalid":
		// 第一次时可能没有数据，进行数据修正
		return true, handleInvalid(ctx)
	case "OK":
	}
	return false, nil
}

func (re *redisFinishQuota) Finally(ctx context.Context) error {
	// 已经执行过终态的逻辑则跳过
	if re.state.HasStatus(stateFinally) {
		return nil
	}
	re.state.AddStatus(stateFinally)
	// 没有执行过预占逻辑则跳过
	if re.state.NotHasStatus(stateEvauate) {
		return nil
	}
	return re.lock.Unlock()
}

func (re *redisFinishQuota) Rollback(ctx context.Context) (err error) {
	// 已经执行过终态的逻辑则跳过
	if re.state.HasStatus(stateFinally) {
		return nil
	}
	re.state.AddStatus(stateFinally)

	// 没有执行过预占逻辑需要锁住回滚的过程
	if re.state.NotHasStatus(stateEvauate) {
		if err = re.lock.Lock(); err != nil {
			return
		}
	}
	defer func() {
		re.lock.Unlock()
	}()
	// 回滚使用量
	var resp interface{}
	if resp, err = re.cli.Eval(ctx, re.rollbackScript,
		[]string{re.resourceKey(re.param)}, re.param.Num).Result(); err != nil {
		err = errors.WithStack(ErrQuotaServerDisable.WithResult(err.Error()))
		return
	}
	result, ok := resp.(string)
	if !ok {
		err = errors.WithStack(
			ErrQuotaServerDisable.WithResult(fmt.Sprintf("invalid response:%v", resp)),
		)
		return
	}
	switch result {
	case "Fail":
		err = errors.WithStack(
			ErrResourceQuotaAbnormal.WithResult(
				fmt.Sprintf("num:%d", re.param.Num),
			),
		)
	case "Invalid":
		err = re.sync(ctx)
	case "OK":
	}
	return
}
