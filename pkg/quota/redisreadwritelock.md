# redis 实现的读写锁，有简单的读锁不阻塞，写锁、读锁互相阻塞 功能

## 以下脚本只有返回 OK 才表示加锁成功，只有返回 1 时代表解锁成功，不同的返回代表着不通失败原因的标识

## 读锁加锁逻辑
 - 脚本参数
   - KEYS[1]: 锁对应的key
   - KEYS[2]: 线程持有锁过期时间记录的key 前缀
   - KEYS[3]: 当前锁等待获取写锁的记录key 前缀
   - ARGV[1]: 当前线程获取锁的过期时间
   - ARGV[2]: 当前线程ID
 - 脚本解析
```sql
	-- 判断是否有在等待读锁的线程，如果有就获取锁失败，避免一直有读锁获取时，写锁一直获取不到
	local waitWrite = redis.call('get', KEYS[3]..':wait_write')
		if waitWrite ~= false then
		return redis.call('pttl', KEYS[1])
	end
        -- 获取锁信息
	local mode = redis.call('hget', KEYS[1], 'mode')
	-- 如果目前没有任何线程持有锁 就直接加锁
	if (mode == false) then 
	  -- 设置锁 类型为读锁
	  redis.call('hset', KEYS[1], 'mode', 'read')
	  -- 设置当前持线程的id到锁 map中
	  redis.call('hset', KEYS[1], ARGV[2], 1)
	  -- 设置当前线程的持锁过期标识 和过期时间
	  redis.call('set', KEYS[2]..':1', 1)
	  redis.call('pexpire', KEYS[2]..':1', ARGV[1])
	  -- 设置当前锁的过期时间
	  redis.call('pexpire', KEYS[1], ARGV[1])
	  -- 返回加锁成功
	  return 'OK'
	end
	-- 判断当前锁是否为读锁
	if (mode == 'read') then
	  -- 当前读锁map增加一条线程id 记录 （这里支持线程重入读锁）
          local ind = redis.call('hincrby', KEYS[1], ARGV[2], 1) 
          -- 设置当前线程的持锁过期标识 和过期时间
	  local key = KEYS[2] .. ':' .. ind
	  redis.call('set', key, 1)
	  redis.call('pexpire', key, ARGV[1])
	  -- 设置当前锁的过期时间，取锁的过期时间和当前线程持有锁的过期时间中的最大值
	  local remainTime = redis.call('pttl', KEYS[1])
	  redis.call('pexpire', KEYS[1], math.max(remainTime, ARGV[1]))
	  -- 返回加锁成功
	  return 'OK'
	end
	-- 当前锁为写锁，加锁失败，返回锁过期时间
	return redis.call('pttl', KEYS[1])
```

## 读锁解锁逻辑
- 脚本参数
    - KEYS[1]: 锁对应的key
    - KEYS[2]: 线程持有锁过期时间记录的key 前缀
    - KEYS[3]: 当前锁key对应的前缀
    - ARGV[1]: 当前线程ID
 - 脚本解析
```sql
        -- 判断当前锁是否存在
        local mode = redis.call('hget', KEYS[1], 'mode')
		-- 不存在直接返回
        if (mode == false) then 
			return 3
		end
        -- 如果是写锁也直接返回
		if mode == 'write' then  
			return -1 
		end
		-- 判断当前线程的持有读锁记录是否存在，如果不存在直接返回
		local lockExists = redis.call('hexists', KEYS[1], ARGV[1])
		if (lockExists == 0) then 
			return 2
		end
		-- 当前线程记录值 减一
		local counter = redis.call('hincrby', KEYS[1], ARGV[1], -1) 
		-- 如果是当前线程最后一条记录，直接 删除
        if (counter == 0) then 
			redis.call('hdel', KEYS[1], ARGV[1]) 
		end
        -- 删除当前线程持有读锁的过期时间记录
		redis.call('del', KEYS[2] .. ':' .. (counter+1))
        -- 判断当前读锁是否还有 其它线程占用
		if (redis.call('hlen', KEYS[1]) > 1) then 
			local maxRemainTime = 0 
			local keys = redis.call('hkeys', KEYS[1]) 
            -- 遍历当前还占用读锁的线程的过期时间记录的过期时间，取过期时间的最大值
			for n, key in ipairs(keys) do  
				counter = tonumber(redis.call('hget', KEYS[1], key)) 
				if type(counter) == 'number' then  
					for i=counter, 1, -1 do  
						local remainTime = redis.call('pttl', KEYS[3] .. ':' .. key .. ':rwlock_timeout:' .. i) 
						maxRemainTime = math.max(remainTime, maxRemainTime) 
					end 
				end 
			end
            -- 如果过期时间大于0, 奖当前过期时间刷新到读锁map中，并返回
			if maxRemainTime > 0 then 
				redis.call('pexpire', KEYS[1], maxRemainTime)
				return 0
			end
		end
        -- 当前读锁已无线程占用，直接删除， 获取锁成功
		redis.call('del', KEYS[1])
        -- 推送锁删除信息到对应的队列
		redis.call('PUBLISH',KEYS[3]..':channel',ARGV[1]..'_del_read')
		return 1
```
- 这里有个问题，就是如果读锁记录不被释放，且一直都有读锁记录获取，那么锁的存储map 长度会越来越大。 不过一般这种情况 可以避免
## 写锁加锁逻辑
- 脚本参数
    - KEYS[1]: 锁对应的key
    - KEYS[2]: 当前锁key对应的前缀
    - ARGV[1]: 当前获取锁过期时间
    - ARGV[2]: 当前线程ID
    - ARGV[3]: 优先获取写锁的占位记录过期时间
 - 脚本解析
```sql
        -- 判断锁是否存在
        local mode = redis.call('hget', KEYS[1], 'mode')
		-- 不存在直接加锁
        if (mode == false) then
          -- 判断当前线程是否 第一个等待加写锁的线程
          local waitWrite = redis.call('get', KEYS[2]..':wait_write')
		  -- 当前线程不是第一个等待加写锁的线程， 返回获取锁失败
          if waitWrite ~= false and waitWrite ~= ARGV[2] then
			return -1
		  end
		  -- 当前线程是第一个等待加锁的线程, 删除等待加锁的记录
          if waitWrite ~= false then 
			redis.call('del', KEYS[2]..':wait_write')
		  end
          -- 加锁，记录写锁类型，当前线程的标识ID,并设置过期时间
		  redis.call('hset', KEYS[1], 'mode', 'write')
		  redis.call('hset', KEYS[1], ARGV[2], 1)
		  redis.call('pexpire', KEYS[1], ARGV[1])
          -- 返回加锁成功
		  return 'OK'
		end
        -- 当前锁已存在且为读锁
		if (mode == 'read') then
            -- 判断当前锁是否存在 等待加写锁的标识
			local waitWrite = redis.call('get', KEYS[2]..':wait_write')
			-- 不存在，则为当前线程创建 等待加写锁的记录标识，并设置过期时间
            if waitWrite == false then
				redis.call('set', KEYS[2]..':wait_write',ARGV[2])
				redis.call('pexpire', KEYS[2]..':wait_write',ARGV[3])
				return 1
			end
            -- 等待加写锁的线程就是当前线程，刷新标识过期时间
			if waitWrite == ARGV[2] then
				redis.call('pexpire', KEYS[2]..':wait_write',ARGV[3])
			end
		end
		-- 返回加锁失败
		return 2
```

## 写锁解锁逻辑
- 脚本参数
    - KEYS[1]: 锁对应的key
    - KEYS[2]: 当前锁key对应的前缀
    - ARGV[1]: 当前线程ID
- 脚本解析
```sql
        -- 判断当前锁是否写锁
        if (redis.call('hget', KEYS[1], 'mode') == 'write') then
			-- 判断持有写锁者是否当前线程
            if (tonumber(redis.call('hget', KEYS[1], ARGV[1])) == 1) then
				-- 是当前线程则删除锁，发送解锁成功消息到对应的消息队列
                redis.call('del', KEYS[1])
				redis.call('PUBLISH',KEYS[2]..':channel',ARGV[1]..'_del_write')
				return 1
			end
		end
		-- 不是写锁，或锁不存在，返回解锁失败
		return -1
```