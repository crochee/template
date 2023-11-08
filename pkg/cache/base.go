package cache

import (
	"context"
	"errors"
	"time"
)

var ErrNil = errors.New("nil")

type CacheInterface interface {
	Get(ctx context.Context, key string) (*Value, error)
	Set(ctx context.Context, key string, value *Value, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	Del(ctx context.Context, key string) error
}

type Value struct {
	Status      int
	ContentType string
	Body        []byte
}
