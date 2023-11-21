package syncx

import (
	"time"

	"github.com/go-redis/redis/v8"
)

type option struct {
	client         *redis.ClusterClient
	expiration     time.Duration
	waitTimeout    time.Duration
	clientIDPrefix string
}

type Option func(*option)

func WithExpireDuration(dur time.Duration) Option {
	return func(o *option) {
		o.expiration = dur
	}
}

func WithWaitTimeout(timeout time.Duration) Option {
	return func(o *option) {
		o.waitTimeout = timeout
	}
}

func WithClientIDPrefix(prefix string) Option {
	return func(o *option) {
		o.clientIDPrefix = prefix
	}
}

func WithClient(client *redis.ClusterClient) Option {
	return func(o *option) {
		o.client = client
	}
}
