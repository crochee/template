package syncx

import (
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestMutex_Lock(t *testing.T) {
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

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, index int) {
			defer wg.Done()
			mutex := NewMutex("gs2g", WithClient(client))
			if err := mutex.Lock(); err != nil {
				t.Log(err)
				return
			}
			t.Log("lock success", index)
			time.Sleep(5 * time.Second)
			if err := mutex.Unlock(); err != nil {
				t.Log(err)
				return
			}
			t.Log("release success", index)
			if err := mutex.Lock(); err != nil {
				t.Log(err)
				return
			}
			t.Log("lock twice success", index)
			time.Sleep(3 * time.Second)
			if err := mutex.Unlock(); err != nil {
				t.Log(err)
				return
			}
			t.Log("release twice success", index)
		}(&wg, i)
	}
	wg.Wait()
}
