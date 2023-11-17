package quota

import (
	"time"
)

const (
	LockKey          = "dcs-lock-%s"
	LockLeaseTime    = time.Second * 30
	PreWriteLockTime = time.Second * 10 // 第一个读锁预先获得锁的等待时间
)

// Redis Key
const (
	ResourceKey           = "dcs:resource:{%s}:%s" // account
	ResourceUsedField     = "used"                 // resource used
	ResourceQuotaField    = "quota"                // resource quota
	ResourceUpdateTimeKey = "dcs:resource:{%s}:updatetime"
	QuotaAccountZSetKey   = "dcs:quota:account" // 存储含有配额数据用户的ZSet key， 权重为更新时间
)

const (
	QuotaAccountExprieTime = time.Hour * 24
)

// Redis 脚本响应状态
const (
	OperationSuccess    = "OK"
	OperationFailure    = "Fail"
	ResourceDataInvalid = "Invalid"
)

// ResourceName
const (
	CPUResource    = "cpu"
	MemoryResource = "memory"
	// ……
)
