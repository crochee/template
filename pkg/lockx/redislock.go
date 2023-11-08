package lockx

import (
	"context"
	"time"

	"github.com/bsm/redislock"

	"template/pkg/redis"
)

type Locker struct {
	ctx    context.Context
	client *redis.Client
	locker *redislock.Lock
}

func (l *Locker) Key() string {
	return l.locker.Key()
}

func (l *Locker) Token() string {
	return l.locker.Token()
}

func (l *Locker) TTL() (time.Duration, error) {
	return l.locker.TTL(l.ctx)
}

func (l *Locker) Refresh(ttl time.Duration) error {
	return l.locker.Refresh(l.ctx, ttl, nil)
}

func (l *Locker) Release() error {
	defer l.client.Close()
	return l.locker.Release(l.ctx)
}

func NewLock(key string, ttl time.Duration) (*Locker, error) {
	client := redis.NewClient()
	locker, err := redislock.Obtain(client.Context(), client.Client(), key, ttl, nil)
	if err != nil {
		return nil, err
	}

	l := &Locker{
		ctx:    client.Context(),
		client: client,
		locker: locker,
	}
	return l, nil
}
