package redisx

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"

	"template/pkg/json"
	"template/pkg/registry"
)

//go:generate mockgen -source=./event.go -destination=./event_mock.go -package=redisx

type Encoder interface {
	Encode(action Action, info *registry.Info, expireAt int64) (string, error)
	Decode(data string) (Action, *registry.Info, int64, error)
}

type EventFlow interface {
	Change(ctx context.Context, msg *redis.Message)
}

type KeyGenerator interface {
	Create(info *registry.Info) string
}

type Event struct {
	ExpireAt    int64
	Action      Action
	UUID        string
	ServiceName string
	Addr        string
	Weight      int
	Tags        map[string]interface{}
}

type DefaultEncoder struct {
}

func (d DefaultEncoder) Encode(action Action, info *registry.Info, expireAt int64) (string, error) {
	meta, err := json.Marshal(&Event{
		ExpireAt:    expireAt,
		Action:      action,
		UUID:        info.UUID,
		ServiceName: info.ServiceName,
		Addr:        info.Addr,
		Weight:      info.Weight,
		Tags:        info.Tags,
	})
	if err != nil {
		return "", err
	}
	return string(meta), nil
}

func (d DefaultEncoder) Decode(data string) (Action, *registry.Info, int64, error) {
	var result Event
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return "", nil, 0, err
	}
	return result.Action, &registry.Info{
		UUID:        result.UUID,
		ServiceName: result.ServiceName,
		Addr:        result.Addr,
		Weight:      result.Weight,
		Tags:        result.Tags,
	}, result.ExpireAt, nil
}

type DefaultKeyGenerator struct {
}

func (d DefaultKeyGenerator) Create(info *registry.Info) string {
	return fmt.Sprintf("%s/provider", info.ServiceName)
}
