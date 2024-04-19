package syncx

import (
	"time"
)

func channelName(key string) string {
	return "redisson_lock__channel" + ":{" + key + "}"
}

type option struct {
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
