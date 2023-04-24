package server

import (
	"context"
	"database/sql/driver"
	"fmt"

	"gorm.io/datatypes"

	"template/pkg/storage"
)

type ClientFunc func(context.Context, ...storage.Opt) *storage.DB

var dbClient ClientFunc

func RegisterClient(ctx context.Context, f ClientFunc) error {
	dbClient = f
	// register model
	p := &ProduceConsumeConf{}
	return dbClient(ctx).Set("gorm:table_options",
		"ENGINE=InnoDB COMMENT='生产消费配置信息表' DEFAULT CHARSET='utf8mb4'").
		AutoMigrate(p)
}

// ProduceConsumeConf 生产消费配置信息表
type ProduceConsumeConf struct {
	ID         uint64     `json:"id" gorm:"primary_key:id;comment:主键id"`
	Topic      string     `json:"topic" gorm:"type:varchar(255);not null;comment:消息队列的topic"`
	URI        string     `json:"uri" gorm:"type:varchar(255);not null;comment:rabbit的地址"`
	Exchange   string     `json:"exchange" gorm:"type:varchar(255);not null;comment:rabbit的exchange"`
	RoutingKey string     `json:"routing_key" gorm:"type:varchar(255);not null;comment:rabbit的routing_key"`
	Enable     uint32     `json:"enable" gorm:"not null;comment:消息队列的【启动与关闭】- 1激活,2关闭"`
	ClientType ClientType `json:"client_type" gorm:"not null;index:idx_client_type_deleted,unique;comment:链接的类型"`

	ExchangeArgs datatypes.JSON `json:"exchange_args" gorm:"column:exchange_args;type:json;comment:交换机参数"`
	QueueArgs    datatypes.JSON `json:"queue_args" gorm:"column:queue_args;type:json;comment:队列参数"`
	BindArgs     datatypes.JSON `json:"bind_args" gorm:"column:bind_args;type:json;comment:绑定参数"`

	Deleted storage.Deleted `json:"-" gorm:"not null;index:idx_client_type_deleted,unique;comment:软删除记录id"`
	storage.Base
}

const (
	FaultProducer          ClientType = "fault_producer"
	FaultConsumer          ClientType = "fault_consumer"
	APIAsyncWodenConsumer  ClientType = "api_async_woden_consumer"
	APIAsyncRudderProducer ClientType = "api_async_rudder_producer"
	APIAsyncDeadLetter     ClientType = "api_async_dead_letter"
)

// TableName get sql table name.获取数据库表名
func (*ProduceConsumeConf) TableName() string {
	return "produce_consume_conf"
}

type ClientType string

// Scan implements the Scanner interface.
func (c *ClientType) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("%v isn't []byte", value)
	}
	*c = ClientType(v)
	return nil
}

// Value implements the driver Valuer interface.
func (c *ClientType) Value() (driver.Value, error) {
	return string(*c), nil
}

func (c ClientType) Eq(v string) bool {
	return string(c) == v
}
