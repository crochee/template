package quota

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/sync/singleflight"

	wsserver "template/pkg/msg/server"
	"template/pkg/quota/errcode"
	"template/pkg/redis"
)

func CreateReadLock(ctx context.Context, key string, userId uint64, leaseTime time.Duration) Locker {
	return &redisReadLock{
		ctx:       ctx,
		key:       key,
		leaseTime: leaseTime,
		userId:    userId,
	}
}

func CreateWriteLock(ctx context.Context, key string, userId uint64, leaseTime time.Duration) Locker {
	return &redisWriteLock{
		ctx:       ctx,
		key:       key,
		leaseTime: leaseTime,
		userId:    userId,
	}
}

type Locker interface {
	Lock() error
	Unlock() error
}

type redisReadLock struct {
	ctx       context.Context
	key       string
	leaseTime time.Duration
	userId    uint64
}

func (r *redisReadLock) Lock() error {
	r.ctx, _ = context.WithTimeout(context.Background(), r.leaseTime)
	script := `
	local waitWrite = redis.call('get', KEYS[3]..':wait_write')
		if waitWrite ~= false then
		return redis.call('pttl', KEYS[1])
	end

	local mode = redis.call('hget', KEYS[1], 'mode')
	if (mode == false) then 
	  redis.call('hset', KEYS[1], 'mode', 'read')
	  redis.call('hset', KEYS[1], ARGV[2], 1)
	  redis.call('set', KEYS[2]..':1', 1)
	  redis.call('pexpire', KEYS[2]..':1', ARGV[1])
	  redis.call('pexpire', KEYS[1], ARGV[1])
	  return 'OK'
	end
	if (mode == 'read') then
      local ind = redis.call('hincrby', KEYS[1], ARGV[2], 1) 
	  local key = KEYS[2] .. ':' .. ind
	  redis.call('set', key, 1)
	  redis.call('pexpire', key, ARGV[1])
	  local remainTime = redis.call('pttl', KEYS[1])
	  redis.call('pexpire', KEYS[1], math.max(remainTime, ARGV[1]))
	  return 'OK'
	end
	return redis.call('pttl', KEYS[1])
`
	client := redis.NewRedisClient()
	timeout := time.Now().Add(r.leaseTime)
	for {
		res, err := client.Eval(script, []string{r.key, fmt.Sprintf("{%s}:%d:rwlock_timeout", r.key, r.userId), "{" + r.key + "}"}, r.leaseTime.Milliseconds(), r.userId)
		if err != nil {
			return err
		}
		if result, ok := res.(string); ok && result == OperationSuccess {
			return nil
		}
		if time.Now().After(timeout) {
			// 获取锁超时
			break
		}
		_ = WaitUnlock(r.ctx, r.key)
	}
	wsserver.Errorf(r.ctx, errcode.ErrCodeWaitLockTimeout, "get read lock timeout,key:%s", r.key)
	return errcode.ErrCodeWaitLockTimeout
}

func (r redisReadLock) Unlock() error {
	script := `
		local mode = redis.call('hget', KEYS[1], 'mode')
		if (mode == false) then 
			return 3
		end
		if mode == 'write' then  
			return -1 
		end
		local lockExists = redis.call('hexists', KEYS[1], ARGV[1])
		if (lockExists == 0) then 
			return 2
		end
		local counter = redis.call('hincrby', KEYS[1], ARGV[1], -1) 
		if (counter == 0) then 
			redis.call('hdel', KEYS[1], ARGV[1]) 
		end
		redis.call('del', KEYS[2] .. ':' .. (counter+1))
		if (redis.call('hlen', KEYS[1]) > 1) then 
			local maxRemainTime = 0 
			local keys = redis.call('hkeys', KEYS[1]) 
			for n, key in ipairs(keys) do  
				counter = tonumber(redis.call('hget', KEYS[1], key)) 
				if type(counter) == 'number' then  
					for i=counter, 1, -1 do  
						local remainTime = redis.call('pttl', KEYS[3] .. ':' .. key .. ':rwlock_timeout:' .. i) 
						maxRemainTime = math.max(remainTime, maxRemainTime) 
					end 
				end 
			end
			if maxRemainTime > 0 then 
				redis.call('pexpire', KEYS[1], maxRemainTime)
				return 0
			end
		end
		redis.call('del', KEYS[1])
		redis.call('PUBLISH',KEYS[3]..':channel',ARGV[1]..'_del_read')
		return 1
	`
	client := redis.NewRedisClient()
	res, err := client.Eval(script, []string{r.key, fmt.Sprintf("{%s}:%d:rwlock_timeout", r.key, r.userId), "{" + r.key + "}"}, r.userId)
	if err != nil {
		return err
	}
	result, _ := res.(int64)
	if result == -1 || result == 2 || result == 3 {
		wsserver.Errorf(r.ctx, err, "read lock unlock fail, not held lock: %s,type: %d", r.key, result)
		return nil
	}
	return nil
}

type redisWriteLock struct {
	ctx       context.Context
	key       string
	leaseTime time.Duration
	userId    uint64
}

func (r redisWriteLock) Lock() error {
	r.ctx, _ = context.WithTimeout(context.Background(), r.leaseTime)
	script := `
		local mode = redis.call('hget', KEYS[1], 'mode')
		if (mode == false) then
          local waitWrite = redis.call('get', KEYS[2]..':wait_write')
		  if waitWrite ~= false and waitWrite ~= ARGV[2] then
			return -1
		  end
		  if waitWrite ~= false then 
			redis.call('del', KEYS[2]..':wait_write')
		  end
		  redis.call('hset', KEYS[1], 'mode', 'write')
		  redis.call('hset', KEYS[1], ARGV[2], 1)
		  redis.call('pexpire', KEYS[1], ARGV[1])
		  return 'OK'
		end
		if (mode == 'read') then
			local waitWrite = redis.call('get', KEYS[2]..':wait_write')
			if waitWrite == false then
				redis.call('set', KEYS[2]..':wait_write',ARGV[2])
				redis.call('pexpire', KEYS[2]..':wait_write',ARGV[3])
				return 1
			end
			if waitWrite == ARGV[2] then
				redis.call('pexpire', KEYS[2]..':wait_write',ARGV[3])
			end
		end
		return 2
	`

	client := redis.NewRedisClient()
	timeout := time.Now().Add(r.leaseTime)
	for {
		res, err := client.Eval(script, []string{r.key, "{" + r.key + "}"}, r.leaseTime.Milliseconds(), r.userId, PreWriteLockTime.Milliseconds())
		if err != err {
			return err
		}
		if result, ok := res.(string); ok && result == OperationSuccess {
			return nil
		}
		if time.Now().After(timeout) {
			// 获取锁超时
			break
		}
		_ = WaitUnlock(r.ctx, r.key)
	}
	wsserver.Errorf(r.ctx, errcode.ErrCodeWaitLockTimeout, "get write lock timeout,key:%s", r.key)
	return errcode.ErrCodeWaitLockTimeout
}

func (r redisWriteLock) Unlock() error {
	script := `
		if (redis.call('hget', KEYS[1], 'mode') == 'write') then
			if (tonumber(redis.call('hget', KEYS[1], ARGV[1])) == 1) then
				redis.call('del', KEYS[1])
				redis.call('PUBLISH',KEYS[2]..':channel',ARGV[1]..'_del_write')
				return 1
			end
		end
		return -1
	`
	client := redis.NewRedisClient()
	res, err := client.Eval(script, []string{r.key, "{" + r.key + "}"}, r.userId)
	if err != nil {
		return err
	}
	result, _ := res.(int64)
	if result == -1 {
		wsserver.Errorf(r.ctx, err, "write lock unlock fail, not held lock: %s", r.key)
		return nil
	}
	return nil
}

var group singleflight.Group

func WaitUnlock(ctx context.Context, key string) error {
	_, err, _ := group.Do(key, func() (interface{}, error) {
		ctx, cancel := context.WithCancel(ctx)
		client := redis.NewRedisClient()
		pub := client.Subscribe("{" + key + "}:channel")
		defer pub.Close()
		go func() {
			// 先订阅 再确认下锁还在不在
			// 如果放锁速度很快，那就会收不到订阅，就要这里结束订阅通知
			time.Sleep(time.Millisecond * 10)
			res, _ := client.Exists(key)
			if res != 1 {
				cancel()
			}
		}()
		_, err := pub.ReceiveMessage(ctx)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	return nil
}

func ExitLock(ctx context.Context, key string) (string, error) {
	client := redis.NewRedisClient()
	lockMode, err := client.HGet(key, "mode")
	if err != nil {
		if errors.Is(err, redis.NilErr) {
			return "", nil
		}
	}
	return lockMode, err
}
