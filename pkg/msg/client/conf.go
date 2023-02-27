package client

import (
	"context"
	"net/http"
	"time"

	"template/internal/gateway/base"
	clientx "template/pkg/client"
)

type ProduceConf interface {
	Get(ctx context.Context) (*ProduceConfig, error)
	Update(ctx context.Context, cfg *ProduceConfig) error
}

func NewProduceConf(endpoint string) ProduceConf {
	return &produceConf{
		IRequest: clientx.NewResource().
			AddEndpoint(endpoint).
			AddPath("config"),
	}
}

type produceConf struct {
	clientx.IRequest
}

func (f *produceConf) Get(ctx context.Context) (*ProduceConfig, error) {
	var result ProduceConfig
	if err := f.To().WithRequest(base.DCSRequest{}).
		WithResponse(base.Parser{}).
		Method(http.MethodGet).
		Prefix("v1", "produce").
		Do(ctx, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

type ProduceConfig struct {
	URI        string        `json:"uri"`
	RoutingKey string        `json:"routing_key"`
	Exchange   string        `json:"exchange"`
	Limit      uint64        `json:"limit"`
	Timeout    time.Duration `json:"timeout"`
	Enable     uint32        `json:"enable"`
	Topic      string        `json:"topic"`
}

func (f *produceConf) Update(ctx context.Context, cfg *ProduceConfig) error {
	if err := f.To().WithRequest(base.DCSRequest{}).
		WithResponse(base.Parser{}).
		Method(http.MethodPatch).
		Prefix("v1", "produce").
		Body(cfg).
		DoNop(ctx); err != nil {
		return err
	}
	return nil
}
