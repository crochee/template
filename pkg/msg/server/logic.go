package server

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type ProduceConfig struct {
	// rabbit 的URI
	// in: query
	URI string `json:"uri"`
	// rabbit 的routing_key
	// in: query
	RoutingKey string `json:"routing_key"`
	// rabbit 的exchange
	// in: query
	Exchange string `json:"exchange"`
	// 消息队列限制
	// in: query
	Limit uint64 `json:"limit"`
	// 超时时间
	// in: query
	Timeout time.Duration `json:"timeout"`
	// 消息队列的【启动与关闭】- 1激活,2关闭
	// in: query
	Enable uint32 `json:"enable"`
	// 消息队列的topic
	// in: query
	Topic string `json:"topic"`
}

// GetProduce 获取当前服务配置
func GetProduce(ctx context.Context) (*ProduceConfig, error) {
	param := GetConfig()
	tempPcc := &ProduceConsumeConf{
		Topic:      param.Topic,
		URI:        param.URI,
		Exchange:   param.Exchange,
		RoutingKey: param.RoutingKey,
		Enable:     param.Enable,
		ClientType: FaultProducer,
	}
	pcc := &ProduceConsumeConf{}
	if err := dbClient(ctx).Model(pcc).Where("`client_type` = ?", FaultProducer).First(pcc).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if err = dbClient(ctx).Model(pcc).Create(tempPcc).Error; err != nil {
			return nil, err
		}
	} else {
		updates := make(map[string]interface{})
		if pcc.Topic != tempPcc.Topic {
			updates["topic"] = tempPcc.Topic
		}
		if pcc.URI != tempPcc.URI {
			updates["uri"] = tempPcc.URI
		}
		if pcc.Exchange != tempPcc.Exchange {
			updates["exchange"] = tempPcc.Exchange
		}
		if pcc.RoutingKey != tempPcc.RoutingKey {
			updates["routing_key"] = tempPcc.RoutingKey
		}
		if pcc.Enable != tempPcc.Enable {
			updates["enable"] = tempPcc.Enable
		}
		if err = dbClient(ctx).Model(pcc).Where("id = ?", pcc.ID).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return &ProduceConfig{
		URI:        param.URI,
		RoutingKey: param.RoutingKey,
		Exchange:   param.Exchange,
		Limit:      param.Limit,
		Timeout:    param.Timeout,
		Enable:     param.Enable,
		Topic:      param.Topic,
	}, nil
}

func UpdateProduce(ctx context.Context, param *ProduceConfig) error {
	// 更新内存rabbitMQ配置
	if err := UpdateConfig(&ConfigParam{
		URI:        param.URI,
		RoutingKey: param.RoutingKey,
		Exchange:   param.Exchange,
		Limit:      param.Limit,
		Enable:     param.Enable,
		Topic:      param.Topic,
		Timeout:    param.Timeout,
	}); err != nil {
		return err
	}
	_, err := GetProduce(ctx)
	return err
}

func InitProduceConfig(ctx context.Context) (*ProduceConfig, error) {
	pcc := &ProduceConsumeConf{}
	if err := dbClient(ctx).Model(pcc).Where("client_type = ?", FaultProducer).First(pcc).Error; err != nil {
		return nil, err
	}
	return &ProduceConfig{
		URI:        pcc.URI,
		RoutingKey: pcc.RoutingKey,
		Exchange:   pcc.Exchange,
		Enable:     pcc.Enable,
		Topic:      pcc.Topic,
	}, nil
}

// GetConfigWithType get mq config or other with client type
func GetConfigWithType(ctx context.Context, clientType ClientType) (*ProduceConsumeConf, error) {
	pcc := &ProduceConsumeConf{}
	if err := dbClient(ctx).Model(pcc).Where("client_type = ?", string(clientType)).First(pcc).Error; err != nil {
		return nil, err
	}
	return pcc, nil
}

// CreateOrUpdateConfigWithType insert or update mq config or other with client type
func CreateOrUpdateConfigWithType(ctx context.Context, pcc *ProduceConsumeConf) error {
	p, err := GetConfigWithType(ctx, pcc.ClientType)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dbClient(ctx).Model(pcc).Create(pcc).Error
		}
		return err
	}
	pcc.ID = p.ID
	return dbClient(ctx).Model(pcc).Updates(pcc).Error
}
