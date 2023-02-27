// Package msg
package msg

import (
	"context"
	"io"
	"sync"
	"time"
)

type Queue interface {
	io.Closer
	Length() int
	Read() <-chan *Metadata
	Write(*Metadata)
	ListAndClear() []*Metadata
}

func NewQueue(capacity int64) Queue {
	return &standardQueue{
		queue: make(chan *Metadata, capacity),
	}
}

type standardQueue struct {
	queue chan *Metadata
}

func (s *standardQueue) Length() int {
	return len(s.queue)
}

func (s *standardQueue) Read() <-chan *Metadata {
	return s.queue
}

func (s *standardQueue) Write(metadata *Metadata) {
	s.queue <- metadata
}

func (s *standardQueue) ListAndClear() []*Metadata {
	result := make([]*Metadata, 0, s.Length())
	for s.Length() > 0 {
		result = append(result, <-s.Read())
	}
	return result
}

func (s *standardQueue) Close() error {
	close(s.queue)
	return nil
}

type QueuePool interface {
	Get(key string) Queue
	SendOn(func([]*Metadata))
}

func NewQueuePool(ctx context.Context, interval, expiration time.Duration, newQueue func() Queue) QueuePool {
	if expiration < 0 {
		expiration = 30 * time.Minute
	}
	if interval <= 0 {
		interval = expiration
	}
	object := &queuePool{
		getQueue:          newQueue,
		interval:          interval,
		defaultExpiration: expiration,
		queueList:         make(map[string]*item, 10),
		finishHandler:     noopFinish,
	}
	go object.run(ctx)
	return object
}

type item struct {
	object     Queue
	expiration int64
}

type queuePool struct {
	getQueue          func() Queue
	interval          time.Duration
	defaultExpiration time.Duration
	pool              sync.Pool
	queueList         map[string]*item
	rwMutex           sync.RWMutex
	finishHandler     func([]*Metadata)
}

func (q *queuePool) Get(key string) Queue {
	q.rwMutex.Lock()
	itemValue, found := q.queueList[key]
	if !found {
		// add
		value, ok := q.pool.Get().(*item)
		if !ok {
			value = &item{
				object:     q.getQueue(),
				expiration: time.Now().Add(q.defaultExpiration).UnixNano(),
			}
		} else {
			value.expiration = time.Now().Add(q.defaultExpiration).UnixNano()
		}
		q.queueList[key] = value
		q.rwMutex.Unlock()
		return value.object
	}
	q.rwMutex.Unlock()
	if itemValue.expiration > 0 {
		if time.Now().UnixNano() > itemValue.expiration {
			q.finishHandler(itemValue.object.ListAndClear())
			itemValue.expiration = time.Now().Add(q.defaultExpiration).UnixNano()
			return itemValue.object
		}
	}
	return itemValue.object
}

func (q *queuePool) SendOn(f func([]*Metadata)) {
	q.finishHandler = f
}

func (q *queuePool) run(ctx context.Context) {
	ticker := time.NewTicker(q.interval)
	for {
		select {
		case <-ticker.C:
			q.deleteExpired()
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

// Delete all expired items from the cache.
func (q *queuePool) deleteExpired() {
	now := time.Now().UnixNano()
	q.rwMutex.Lock()
	for k, v := range q.queueList {
		if v.expiration > 0 && now > v.expiration {
			delete(q.queueList, k)
			q.finishHandler(v.object.ListAndClear())
			v.expiration = 0
			q.pool.Put(v)
		}
	}
	q.rwMutex.Unlock()
}

func noopFinish([]*Metadata) {
}
