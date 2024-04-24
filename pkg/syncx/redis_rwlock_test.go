package syncx

import (
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestNewRWMutex(t *testing.T) {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			":7001",
			":7002",
			":7003",
			":7004",
			":7005",
			":7000"},
	})
	defer client.Close()
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, index int) {
			defer wg.Done()
			mutex := NewRWMutex("g23", client, WithWaitTimeout(30*time.Minute))
			if err := mutex.Lock(); err != nil {
				t.Log(err)
				return
			}
			t.Log("Lock success", index)
			time.Sleep(5 * time.Second)
			if err := mutex.Unlock(); err != nil {
				t.Log(err)
				return
			}
			t.Log("Unlock success", index)
			if err := mutex.Lock(); err != nil {
				t.Log(err)
				return
			}
			t.Log("Lock twice success", index)
			time.Sleep(3 * time.Second)
			if err := mutex.Unlock(); err != nil {
				t.Log(err)
				return
			}
			t.Log("Unlock twice success", index)
			if err := mutex.RLock(); err != nil {
				t.Log(err)
				return
			}
			t.Log("Rlock success", index)
			time.Sleep(3 * time.Second)
			if err := mutex.RUnlock(); err != nil {
				t.Log(err)
				return
			}
			t.Log("RUnlock success", index)
		}(&wg, i)
	}
	wg.Wait()
}
