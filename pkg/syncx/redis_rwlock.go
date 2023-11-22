package syncx

import (
	"context"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type RWMutex struct {
	pubSub *redis.PubSub

	lockScript    string
	rLockScript   string
	renewalScript string
	unlockScript  string

	key string

	option
}

func NewRWMutex(key string, opts ...Option) *RWMutex {
	o := option{
		expiration:     10 * time.Second,
		waitTimeout:    30 * time.Second,
		clientIDPrefix: uuid.NewV4().String(),
	}
	for _, opt := range opts {
		opt(&o)
	}

	m := &RWMutex{
		pubSub: &redis.PubSub{},
		lockScript: `
	-- KEYS[1] 锁名
	-- ARGV[1] 协程唯一标识：客户端标识+协程ID
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
	-- ARGV[1] 协程唯一标识：客户端标识+协程ID
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
	-- ARGV[1] 协程唯一标识：客户端标识+协程ID
	-- ARGV[2] 解锁时发布的消息
	local t = redis.call('type',KEYS[1])["ok"]
	if  t == "hash" then
		if redis.call('hexists',KEYS[1],ARGV[1]) == 0 then
			return 0
		end
		if redis.call('hincrby',KEYS[1],ARGV[1],-1) == 0 then
			redis.call('hdel',KEYS[1],ARGV[1])
			if (redis.call('hlen',KEYS[1]) > 0 )then
				return 2
			end
			redis.call('del',KEYS[1])
			return redis.call('publish',KEYS[2],ARGV[2])
		else
			return 1
		end
	elseif t == "none" then
			redis.call('publish',KEYS[2],ARGV[2])
			return 1
    elseif t == "string" then
        if redis.call('get',KEYS[1]) == ARGV[1] then
			redis.call('del',KEYS[1])
			return redis.call('publish',KEYS[2],ARGV[2])
        end
        return 0
	else
		return 0
	end
`,
		key:    key,
		option: o,
	}
	m.pubSub = m.client.Subscribe(context.Background(), channelName(m.key))

	runtime.SetFinalizer(m, func(m *RWMutex) {
		if err := m.pubSub.Unsubscribe(context.Background(), channelName(m.key)); err != nil {
			log.Println(err)
		}
		if err := m.pubSub.Close(); err != nil {
			log.Println(err)
		}
	})
	return m
}

func (r RWMutex) Lock() error {
	// 单位：ms
	expiration := int64(r.expiration / time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), r.waitTimeout)
	defer cancel()

	clientID := r.clientIDPrefix + ":" + goroutineNum()
	ch := make(chan result)
	var once sync.Once
	err := r.tryLock(ctx, &once, ch, clientID, expiration)
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

func (r RWMutex) tryLock(ctx context.Context, once *sync.Once, ch chan result, clientID string, expiration int64) error {
	pTTL, err := r.lock(clientID, expiration)
	if err != nil {
		return err
	}
	if pTTL == 0 {
		return nil
	}
	once.Do(func() {
		go func() {
			msg, err := r.pubSub.ReceiveMessage(ctx)
			ch <- result{val: msg, err: err}
		}()
	})
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Duration(pTTL) * time.Millisecond):
		// 针对“redis 中存在未维护的锁”，即当锁自然过期后，并不会发布通知的锁
		return r.tryLock(ctx, once, ch, clientID, expiration)
	case value := <-ch:
		if value.err != nil {
			return value.err
		}
		// 收到解锁通知，则尝试抢锁
		return r.tryLock(ctx, once, ch, clientID, expiration)
	}
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
	// 单位：ms
	expiration := int64(r.expiration / time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), r.waitTimeout)
	defer cancel()

	clientID := r.clientIDPrefix + ":" + goroutineNum()
	ch := make(chan result)
	var once sync.Once
	err := r.tryRLock(ctx, &once, ch, clientID, expiration)
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

func (r RWMutex) tryRLock(ctx context.Context, once *sync.Once, ch chan result, clientID string, expiration int64) error {
	pTTL, err := r.rLock(clientID, expiration)
	if err != nil {
		return err
	}
	if pTTL == 0 {
		return nil
	}
	once.Do(func() {
		go func() {
			msg, err := r.pubSub.ReceiveMessage(ctx)
			ch <- result{val: msg, err: err}
		}()
	})

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Duration(pTTL) * time.Millisecond):
		// 针对“redis 中存在未维护的锁”，即当锁自然过期后，并不会发布通知的锁
		return r.tryRLock(ctx, once, ch, clientID, expiration)
	case value := <-ch:
		if value.err != nil {
			return value.err
		}
		// 收到解锁通知，则尝试抢锁
		return r.tryRLock(ctx, once, ch, clientID, expiration)
	}
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
	clientID := r.clientIDPrefix + ":" + goroutineNum()
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
