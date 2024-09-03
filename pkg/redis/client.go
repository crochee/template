package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"

	"template/pkg/json"
)

var (
	onceOption sync.Once
	rClient    *Client
)

// NewRedisClient 默认获取一个全局的 redis 客户端
func NewRedisClient() *Client {
	if rClient == nil {
		onceOption.Do(func() {
			rClient = NewClient()
		})
	}
	return rClient
}

type Option struct {
	AddrList []string
	Password string
}

func New(ctx context.Context, opts ...func(*Option)) (*redis.ClusterClient, error) {
	o := Option{}
	for _, opt := range opts {
		opt(&o)
	}
	cli := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    o.AddrList,
		Password: o.Password,
	})
	if err := cli.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed.Error:%w", err)
	}
	return cli, nil
}

type Client struct {
	ctx context.Context
	cli *redis.ClusterClient
}

var NilErr = redis.Nil

func (c *Client) Close() error {
	return c.cli.Close()
}

func (c *Client) Del(keys ...string) error {
	return c.cli.Del(c.ctx, keys...).Err()
}

func (c *Client) Pipeline() redis.Pipeliner {
	return c.cli.Pipeline()
}

func (c *Client) Exists(keys ...string) (int64, error) {
	return c.cli.Exists(c.ctx, keys...).Result()
}

func (c *Client) Expire(key string, expiration time.Duration) (bool, error) {
	return c.cli.Expire(c.ctx, key, expiration).Result()
}

func (c *Client) Eval(script string, keys []string, args ...interface{}) (interface{}, error) {
	return c.cli.Eval(c.ctx, script, keys, args...).Result()
}

func (c *Client) Get(key string) (string, error) {
	return c.cli.Get(c.ctx, key).Result()
}

func (c *Client) GetInterface(key string, obj interface{}) error {
	data, err := c.cli.Get(c.ctx, key).Bytes()
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, obj)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetByte(key string) ([]byte, error) {
	return c.cli.Get(c.ctx, key).Bytes()
}

func (c *Client) Subscribe(channels ...string) *redis.PubSub {
	return c.cli.Subscribe(c.ctx, channels...)
}

func (c *Client) PSubscribe(channels ...string) *redis.PubSub {
	return c.cli.PSubscribe(c.ctx, channels...)
}

func (c *Client) BRPop(timeout time.Duration, keys ...string) ([]string, error) {
	return c.cli.BRPop(c.ctx, timeout, keys...).Result()
}

func (c *Client) BRPopLPush(source, destination string, timeout time.Duration) (string, error) {
	return c.cli.BRPopLPush(c.ctx, source, destination, timeout).Result()
}

func (c *Client) FlushDB() error {
	return c.cli.FlushDB(c.ctx).Err()
}

func (c *Client) HDel(key string, fields ...string) error {
	return c.cli.HDel(c.ctx, key, fields...).Err()
}

func (c *Client) HGet(key, field string) (string, error) {
	return c.cli.HGet(c.ctx, key, field).Result()
}

func (c *Client) HGetAll(key string) (map[string]string, error) {
	return c.cli.HGetAll(c.ctx, key).Result()
}

func (c *Client) HLen(key string) (int64, error) {
	return c.cli.HLen(c.ctx, key).Result()
}

func (c *Client) HIncrBy(key, field string, incr int64) (int64, error) {
	return c.cli.HIncrBy(c.ctx, key, field, incr).Result()
}

func (c *Client) HKeys(key string) ([]string, error) {
	return c.cli.HKeys(c.ctx, key).Result()
}

func (c *Client) HMGet(key string, fields ...string) ([]interface{}, error) {
	return c.cli.HMGet(c.ctx, key, fields...).Result()
}

func (c *Client) HSet(key string, values ...interface{}) (int64, error) {
	return c.cli.HSet(c.ctx, key, values...).Result()
}

func (c *Client) Incr(key string) (int64, error) {
	return c.cli.Incr(c.ctx, key).Result()
}

func (c *Client) Keys(pattern string) ([]string, error) {
	return c.cli.Keys(c.ctx, pattern).Result()
}

func (c *Client) LLen(key string) (int64, error) {
	return c.cli.LLen(c.ctx, key).Result()
}

func (c *Client) LPush(key string, values ...interface{}) (int64, error) {
	return c.cli.LPush(c.ctx, key, values...).Result()
}

func (c *Client) LRange(key string, start, stop int64) ([]string, error) {
	return c.cli.LRange(c.ctx, key, start, stop).Result()
}

func (c *Client) LTrim(key string, start, stop int64) (string, error) {
	return c.cli.LTrim(c.ctx, key, start, stop).Result()
}

func (c *Client) MGet(keys ...string) ([]interface{}, error) {
	return c.cli.MGet(c.ctx, keys...).Result()
}

func (c *Client) MSet(values ...interface{}) (string, error) {
	return c.cli.MSet(c.ctx, values...).Result()
}

func (c *Client) Ping() (string, error) {
	return c.cli.Ping(c.ctx).Result()
}

func (c *Client) Publish(channel string, message interface{}) (int64, error) {
	return c.cli.Publish(c.ctx, channel, message).Result()
}

func (c *Client) RPop(key string) (string, error) {
	return c.cli.RPop(c.ctx, key).Result()
}

func (c *Client) RPopLPush(source, destination string) (string, error) {
	return c.cli.RPopLPush(c.ctx, source, destination).Result()
}

func (c *Client) RPush(key string, values ...interface{}) (int64, error) {
	return c.cli.RPush(c.ctx, key, values...).Result()
}

func (c *Client) SAdd(key string, members ...interface{}) (int64, error) {
	return c.cli.SAdd(c.ctx, key, members...).Result()
}

func (c *Client) SIsMember(key string, member interface{}) (bool, error) {
	return c.cli.SIsMember(c.ctx, key, member).Result()
}

func (c *Client) Set(key string, value interface{}, expiration time.Duration) error {
	return c.cli.Set(c.ctx, key, value, expiration).Err()
}

func (c *Client) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.cli.SetNX(c.ctx, key, value, expiration).Result()
}

func (c *Client) SMembers(key string) ([]string, error) {
	return c.cli.SMembers(c.ctx, key).Result()
}

func (c *Client) SRem(key string, members ...interface{}) (int64, error) {
	return c.cli.SRem(c.ctx, key, members...).Result()
}

func (c *Client) TTL(key string) (time.Duration, error) {
	return c.cli.TTL(c.ctx, key).Result()
}

func (c *Client) ZAdd(key string, members ...*redis.Z) (int64, error) {
	return c.cli.ZAdd(c.ctx, key, members...).Result()
}

func (c *Client) ZRemRangeByScore(key, max, min string) (int64, error) {
	return c.cli.ZRemRangeByScore(c.ctx, key, min, max).Result()
}

func (c *Client) ZRange(key string, start, end int64) ([]string, error) {
	return c.cli.ZRange(c.ctx, key, start, end).Result()
}

func (c *Client) ClusterInfo() (string, error) {
	return c.cli.ClusterInfo(c.ctx).Result()
}

func (c *Client) Client() *redis.ClusterClient {
	return c.cli
}

func (c *Client) Context() context.Context {
	return c.ctx
}

func NewClient() *Client {
	opts := redis.ClusterOptions{Addrs: viper.GetStringSlice("redis.addrs")}
	password := viper.GetString("redis.password")
	if password != "" {
		opts.Password = password
	}
	cli := redis.NewClusterClient(&opts)
	return &Client{
		ctx: context.TODO(),
		cli: cli,
	}
}
