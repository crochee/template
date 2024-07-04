package syncx

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type Mutex struct {
	client *redis.ClusterClient

	lockScript    string
	renewalScript string
	unlockScript  string

	key string

	option
}

func NewMutex(key string, client *redis.ClusterClient, opts ...Option) *Mutex {
	o := option{
		expiration:     10 * time.Second,
		waitTimeout:    30 * time.Second,
		clientIDPrefix: uuid.NewV4().String(),
	}
	for _, opt := range opts {
		opt(&o)
	}

	m := &Mutex{
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
		renewalScript: `
	-- KEYS[1] 锁名
	-- ARGV[1] 过期时间
	-- ARGV[2] 客户端协程唯一标识
	if redis.call('get',KEYS[1])==ARGV[2] then
		return redis.call('pexpire',KEYS[1],ARGV[1])
	end
	return 0
`,
		unlockScript: `
	-- KEYS[1] 锁名
	-- KEYS[2] 发布订阅的channel
	-- ARGV[1] 协程唯一标识：客户端标识
	-- ARGV[2] 解锁时发布的消息
	if redis.call('exists',KEYS[1]) == 1 then
		if (redis.call('get',KEYS[1]) == ARGV[1]) then
			redis.call('del',KEYS[1])
		else
			return 0
		end
	end
	redis.call('publish',KEYS[2],ARGV[2])
    return 1
`,
		key:    key,
		option: o,
	}
	return m
}

func (m *Mutex) Lock() error {
	ctx, cancel := context.WithTimeout(context.Background(), m.waitTimeout)
	defer cancel()
	var (
		// 单位：ms
		expiration    = int64(m.expiration / time.Millisecond)
		clientID      = m.clientIDPrefix
		receivePubSub = make(chan struct{})
		releasePubSub = make(chan struct{})
		once          sync.Once
		breakLoop     bool
		err           error
	)
	for !breakLoop {
		breakLoop, err = m.tryLockLoop(
			ctx,
			&once,
			receivePubSub,
			releasePubSub,
			clientID,
			expiration,
			m.lock,
		)
	}
	close(releasePubSub)
	if err != nil {
		return err
	}
	// 加锁成功，开个协程，定时续锁
	go func() {
		ticker := time.NewTicker(m.expiration / 3)
		defer ticker.Stop()
		for range ticker.C {
			res, err := m.client.Eval(context.Background(), m.renewalScript, []string{m.key}, expiration, clientID).
				Int64()
			if err != nil || res == 0 {
				return
			}
		}
	}()
	return nil
}

type result struct {
	val interface{}
	err error
}

// ErrNotObtained is returned when a lock cannot be obtained.
var ErrNotObtained = errors.New("redislock: not obtained")

func (m Mutex) tryLockLoop(ctx context.Context,
	once *sync.Once,
	receivePubSub chan struct{},
	releasePubSub chan struct{},
	clientID string,
	expiration int64,
	lockFunc func(clientID string, expiration int64) (int64, error),
) (breakLoop bool, err error) {
	// 尝试加锁
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
			pubSub := m.client.Subscribe(ctx, channelName(m.key))
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
		// 申请锁的耗时如果大于等于最大等待时间，则申请锁失败.
		breakLoop = true
		err = ctx.Err()
	case <-t.C:
		// 针对“redis 中存在未维护的锁”，即当锁自然过期后，并不会发布通知的锁
	case <-receivePubSub:
		// 收到解锁通知，则尝试抢锁
	}
	return
}

func (m Mutex) TryLock() error {
	// 单位：ms
	expiration := int64(m.expiration / time.Millisecond)

	clientID := m.clientIDPrefix
	// 尝试加锁
	pTTL, err := m.lock(clientID, expiration)
	if err != nil {
		return err
	}
	if pTTL != 0 {
		return fmt.Errorf("key %s already locked, please try again after %d ms,%w",
			m.key, pTTL, ErrNotObtained)
	}
	// 加锁成功，开个协程，定时续锁
	go func() {
		ticker := time.NewTicker(m.expiration / 3)
		defer ticker.Stop()
		for range ticker.C {
			res, err := m.client.Eval(context.Background(), m.renewalScript, []string{m.key}, expiration, clientID).
				Int64()
			if err != nil || res == 0 {
				return
			}
		}
	}()
	return nil
}

func (m Mutex) lock(clientID string, expiration int64) (int64, error) {
	pTTL, err := m.client.Eval(context.Background(), m.lockScript,
		[]string{m.key}, clientID, expiration).
		Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return pTTL.(int64), nil
}

func (m Mutex) Unlock() error {
	clientID := m.clientIDPrefix
	res, err := m.client.Eval(context.Background(), m.unlockScript, []string{m.key, channelName(m.key)}, clientID, 1).
		Int64()
	if err != nil {
		return errors.WithStack(err)
	}
	if res == 0 {
		return errors.Errorf("unknown client: %s", clientID)
	}
	return nil
}
