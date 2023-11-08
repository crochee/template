# WOSLO

# 配额功能

## 使用方法
```go
package main
import (
	"fmt"
	"time"
	"golang.org/x/net/context"
	"woslo/quota"
)

type mockHandler struct {
	Used  int
	Quota int
}

func (m *mockHandler) QueryUsed(ctx context.Context, account string) (used int, err error) {
	return m.Used, nil
}

func (m *mockHandler) QueryQuota(ctx context.Context, account string) (quota int, err error) {
	return m.Quota, nil
}


func main() {
	// 初始化 配额资源数据
	quota.InitResourceQuotaManager(time.Second, time.Minute*3,
		quota.WithUsedQuotaHandler("CPUA", &mockHandler{Used: 10, Quota: 20}),
		quota.WithUsedQuotaHandler("CPUB", &mockHandler{Used: 0, Quota: 20}),
		quota.WithUsedQuotaHandler("CPUC", &mockHandler{Used: 18, Quota: 20}),
		quota.WithUsedQuotaHandler("CPUD", &mockHandler{Used: 20, Quota: 20}),
	)
	ctx := context.Background()
	quota.InitResourceQuotaData(ctx, func(ctx context.Context) ([]string, error) {
        // 查询account 列表函数
		return []string{"account1","account2"},nil
	})

	// 创建资源
	finish,err := quota.PrepareOccupying(ctx,"account1",map[string]uint{
		"CPUA":1,
		"CPUB":2,
    })
	if err != nil {
		// 资源不足
		return
    }
    
    // 业务逻辑
    // 资源数据入库
	if err != nil {
       // 资源入库失败
		_ = finish.Rollback(ctx)
	}
    
    
    _ = finish.Finally(ctx)
    
    
    // 删除资源逻辑
	_ = quota.CleanUsed(ctx,"account1",map[string]uint{
		"CPUA":1,
		"CPUB":2,
	})
	
	// 异步创建失败回滚资源逻辑
	errTime := time.Now()   // 错误发生时间； 
	_ = quota.Rollback(ctx,"account1",map[string]uint{
		"CPUA":1,
		"CPUB":2,
	},errTime)
}
```

## 数据校验接口
 ### GET /used_quota  校验配额数据是否正确
 ### POST /used_quota  刷新配额数据

# 分布式读写锁

## 使用方法
```go
package main
import (
	"fmt"
	"time"
	"woslo/quota"
	"golang.org/x/net/context"
)

func main() {
	ctx := context.Background()
	// 读锁
	mutex := quota.CreateReadLock(ctx,"mutexKey", 123456,time.Minute)
	err := mutex.Lock()
	if err != nil {
		fmt.Printf("read lock error: %v\n", err)
		return
	}

	time.Sleep(2 * time.Millisecond)
	err = mutex.Unlock()
	if err != nil {
		fmt.Printf("read unlock error: %v\n", err)
		return
	}

	// 写锁
	mutex = quota.CreateWriteLock(ctx,"mutexKey", 123456,time.Minute)
	err = mutex.Lock()
	if err != nil {
		fmt.Printf("write lock error: %v\n", err)
		return
	}

	time.Sleep(2 * time.Millisecond)
	err = mutex.Unlock()
	if err != nil {
		fmt.Printf("write unlock error: %v\n", err)
		return
	}
}
```