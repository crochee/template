package cache

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"

	_redis "template/pkg/redis"
)

func DefaultCache() CacheInterface {
	store, err := _redis.New(context.Background(), func(option *_redis.Option) {
		option.AddrList = viper.GetStringSlice("redis.addrs")
		option.Password = viper.GetString("redis.password")
	})
	if err != nil {
		panic(err)
	}
	return &Cache{Store: store}
}

type Cache struct {
	Store *redis.ClusterClient
}

func (c *Cache) Get(ctx context.Context, key string) (*Value, error) {
	data, err := c.Store.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrNil
		}
		return nil, err
	}
	var value Value
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	if err = dec.Decode(&value); err != nil {
		return nil, err
	}
	return &value, nil
}

func (c *Cache) Set(ctx context.Context, key string, value *Value, expiration time.Duration) error {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	if err := enc.Encode(*value); err != nil {
		return err
	}
	return c.Store.Set(ctx, key, b.String(), expiration).Err()
}

func (c *Cache) TTL(ctx context.Context, key string) (time.Duration, error) {
	expiration, err := c.Store.TTL(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, ErrNil
		}
		return 0, err
	}
	return expiration, nil
}

func (c *Cache) Del(ctx context.Context, key string) error {
	return c.Store.Del(ctx, key).Err()
}
