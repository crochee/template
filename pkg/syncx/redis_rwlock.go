package syncx

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type RWMutex struct {
	client *redis.ClusterClient

	lockScript    string
	rLockScript   string
	renewalScript string
	unlockScript  string

	key string

	option
}

func NewRWMutex(key string, client *redis.ClusterClient, opts ...Option) *RWMutex {
	o := option{
		expiration:     10 * time.Second,
		waitTimeout:    30 * time.Second,
		clientIDPrefix: uuid.NewV4().String(),
	}
	for _, opt := range opts {
		opt(&o)
	}

	m := &RWMutex{
		client: client,
		lockScript: `
	-- KEYS[1] 锁名
	-- ARGV[1] 协程唯一标识：客户端标识
	-- ARGV[2] 过期时间
	if redis.call('exists',KEYS[1]) == 0 then
		redis.call('set',KEYS[1],ARGV[1])
		redis.call('pexpire',KEYS[1],ARGV[2])
		return nil
	end
	return redis.call('pttl',KEYS[1])
`,
		rLockScript: `
	-- KEYS[1] 锁名
	-- ARGV[1] 协程唯一标识：客户端标识
	-- ARGV[2] 过期时间
	local t = redis.call('type',KEYS[1])["ok"]
	if t == "string" then
		return redis.call('pttl',KEYS[1])
	else
		redis.call('hincrby',KEYS[1],ARGV[1],1)
		redis.call('pexpire',KEYS[1],ARGV[2])
		return nil
	end
`,
		renewalScript: `
	-- KEYS[1] 锁名
	-- ARGV[1] 过期时间
	-- ARGV[2] 客户端协程唯一标识
	local t = redis.call('type',KEYS[1])["ok"]
	if t =="string" then
		if redis.call('get',KEYS[1])==ARGV[2] then
			return redis.call('pexpire',KEYS[1],ARGV[1])
		end
		return 0
	elseif t == "hash" then
		if redis.call('hexists',KEYS[1],ARGV[2])==0 then
			return 0
		end
		return redis.call('pexpire',KEYS[1],ARGV[1])
	else
		return 0
	end
`,
		unlockScript: `
	-- KEYS[1] 锁名
	-- KEYS[2] 发布订阅的channel
	-- ARGV[1] 协程唯一标识：客户端标识
	-- ARGV[2] 解锁时发布的消息
	local t = redis.call('type',KEYS[1])["ok"]
	if  t == "hash" then
		if redis.call('hexists',KEYS[1],ARGV[1]) == 0 then
			return 0
		end
		if redis.call('hincrby',KEYS[1],ARGV[1],-1) <= 0 then
			redis.call('hdel',KEYS[1],ARGV[1])
			if (redis.call('hlen',KEYS[1]) > 0 )then
				return 2
			end
			redis.call('del',KEYS[1])
			redis.call('publish',KEYS[2],ARGV[2])
            return 1
		else
			return 2
		end
	elseif t == "none" then
			redis.call('publish',KEYS[2],ARGV[2])
			return 1
    elseif t == "string" then
        if redis.call('get',KEYS[1]) == ARGV[1] then
			redis.call('del',KEYS[1])
			redis.call('publish',KEYS[2],ARGV[2])
            return 1
        end
        return 0
	else
		return 0
	end
`,
		key:    key,
		option: o,
	}
	return m
}

func (r RWMutex) Lock() error {
	ctx, cancel := context.WithTimeout(context.Background(), r.waitTimeout)
	defer cancel()
	var (
		// 单位：ms
		expiration    = int64(r.expiration / time.Millisecond)
		clientID      = r.clientIDPrefix
		receivePubSub = make(chan struct{})
		releasePubSub = make(chan struct{})
		once          sync.Once
		breakLoop     bool
		err           error
	)
	for !breakLoop {
		breakLoop, err = r.tryLockLoop(
			ctx,
			&once,
			receivePubSub,
			releasePubSub,
			clientID,
			expiration,
			r.lock,
		)
	}
	close(releasePubSub)
	if err != nil {
		return err
	}
	// 加锁成功，开个协程，定时续锁
	go func() {
		ticker := time.NewTicker(r.expiration / 3)
		defer ticker.Stop()
		for range ticker.C {
			res, err := r.client.Eval(context.Background(), r.renewalScript,
				[]string{r.key}, expiration, clientID).Int64()
			if err != nil || res == 0 {
				return
			}
		}
	}()
	return nil
}

func (r RWMutex) tryLockLoop(
	ctx context.Context,
	once *sync.Once,
	receivePubSub chan struct{},
	releasePubSub chan struct{},
	clientID string,
	expiration int64,
	lockFunc func(clientID string, expiration int64) (int64, error),
) (breakLoop bool, err error) {
	var pTTL int64
	if pTTL, err = lockFunc(clientID, expiration); err != nil {
		return
	}
	if pTTL == 0 {
		breakLoop = true
		return
	}
	once.Do(func() {
		go func() {
			// 开启订阅模式
			pubSub := r.client.Subscribe(ctx, channelName(r.key))
			for {
				select {
				case <-releasePubSub:
					// 读取完后关闭订阅模式
					if err := pubSub.Close(); err != nil {
						log.Println(err)
					}
					return
				case <-pubSub.Channel():
					receivePubSub <- struct{}{}
				}
			}
		}()
	})
	t := time.NewTimer(time.Duration(pTTL) * time.Millisecond)
	defer t.Stop()
	select {
	case <-ctx.Done():
		breakLoop = true
		err = ctx.Err()
	case <-t.C:
		// 针对“redis 中存在未维护的锁”，即当锁自然过期后，并不会发布通知的锁
	case <-receivePubSub:
		// 收到解锁通知，则尝试抢锁
	}
	return
}

func (r RWMutex) lock(clientID string, expiration int64) (int64, error) {
	pTTL, err := r.client.Eval(context.Background(), r.lockScript,
		[]string{r.key}, clientID, expiration).Result()
	if err == redis.Nil {
		return 0, nil
	}

	if err != nil {
		return 0, errors.WithStack(err)
	}

	return pTTL.(int64), nil
}

func (r RWMutex) RLock() error {
	ctx, cancel := context.WithTimeout(context.Background(), r.waitTimeout)
	defer cancel()

	var (
		// 单位：ms
		expiration    = int64(r.expiration / time.Millisecond)
		clientID      = r.clientIDPrefix
		receivePubSub = make(chan struct{})
		releasePubSub = make(chan struct{})
		once          sync.Once
		breakLoop     bool
		err           error
	)
	for !breakLoop {
		breakLoop, err = r.tryLockLoop(
			ctx,
			&once,
			receivePubSub,
			releasePubSub,
			clientID,
			expiration,
			r.rLock,
		)
	}
	close(releasePubSub)
	if err != nil {
		return err
	}
	// 加锁成功，开个协程，定时续锁
	go func() {
		ticker := time.NewTicker(r.expiration / 3)
		defer ticker.Stop()
		for range ticker.C {
			res, err := r.client.Eval(context.Background(), r.renewalScript,
				[]string{r.key}, expiration, clientID).Int64()
			if err != nil || res == 0 {
				return
			}
		}
	}()
	return nil
}

func (r RWMutex) rLock(clientID string, expiration int64) (int64, error) {
	pTTL, err := r.client.Eval(context.Background(), r.rLockScript, []string{r.key},
		clientID, expiration).Result()
	if err == redis.Nil {
		return 0, nil
	}

	if err != nil {
		return 0, errors.WithStack(err)
	}

	return pTTL.(int64), nil
}

func (r RWMutex) Unlock() error {
	clientID := r.clientIDPrefix
	res, err := r.client.Eval(context.Background(), r.unlockScript,
		[]string{r.key, channelName(r.key)}, clientID, 1).Int64()
	if err != nil {
		return errors.WithStack(err)
	}
	if res == 0 {
		return errors.Errorf("unknown client: %s", clientID)
	}
	return nil
}

func (r RWMutex) RUnlock() error {
	return r.Unlock()
}
