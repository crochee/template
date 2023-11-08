package quota

const (
	// 预占 脚本
	PreAppropriationScript = `
		for i, key in ipairs(KEYS) do
			local used = redis.call('HGET', key, 'used')
			local quota = redis.call('HGET', key, 'quota')
			if tonumber(used) == nil or tonumber(quota) == nil then
				return 'Invalid-'..i
			end
			if tonumber(used) + tonumber(ARGV[i]) > tonumber(quota) then 
				return 'Fail-'..i
			end
		end
		for i, key in ipairs(KEYS) do
			redis.call('HINCRBY',key, 'used', tonumber(ARGV[i]))
			if redis.call('TTL',key) == -1 then
				redis.call('DEL',key)
			end
		end 
		return 'OK'
	`

	// 回滚脚本
	RollbackScript = `
		for i, key in ipairs(KEYS) do
			local used = redis.call('HGET', key, 'used')
			if tonumber(used) == nil then
				return 'Invalid-'..i
			end
			if tonumber(used) + tonumber(ARGV[i]) < 0 then 
				return 'Fail-'..i
			end
		end
		for i, key in ipairs(KEYS) do
			redis.call('HINCRBY',key, 'used', -tonumber(ARGV[i]))
			if redis.call('TTL',key) == -1 then
				redis.call('DEL',key)
			end
		end 
		return 'OK'
	`
)
