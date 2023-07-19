package clean

import (
	"context"
	"strings"

	"time"

	"template/pkg/storage"
)

type Policy struct {
	Retention time.Duration // 软删除数据保留时间
	Capacity  int           // 单表容量限制 单位M
	Rows      int           // 单表条数限制
	WhiteList []string      // 不需要操作的表名单
	Name      string        // 数据库名称
}

type PolicyGetter interface {
	Get(ctx context.Context) (*Policy, error)
}

func (p *Policy) Get(ctx context.Context) (*Policy, error) {
	return p, nil
}

// CleanPolicyConfig 清理任务的策略配置
type CleanPolicyConfig struct {
	ID        uint64        `json:"id" gorm:"primary_key:id;comment:主键id"`
	Retention time.Duration `json:"retention"  gorm:"retention;not null;comment:软删除数据的保留时间"`
	Capacity  int           `json:"capacity"  gorm:"capacity;not null;comment:单表容量限制 单位M"`
	Rows      int           `json:"rows" gorm:"rows;not null;comment:单表条数限制"`
	WhiteList string        `json:"white_list" gorm:"white_list;not null;comment:不需要操作的表名单"`

	Deleted storage.Deleted `json:"-" gorm:"not null;index:idx_deleted;comment:软删除记录id"`
	storage.Base
}

// TableName get sql table name.获取数据库表名
func (*CleanPolicyConfig) TableName() string {
	return "clean_policy_config"
}

func Register(ctx context.Context, connector func(context.Context, ...storage.Opt) *storage.DB) error {
	return connector(ctx).Set("gorm:table_options",
		"ENGINE=InnoDB COMMENT='清理任务的策略配置' DEFAULT CHARSET='utf8mb4'").
		AutoMigrate(&CleanPolicyConfig{})
}

func NewDBPolicyGetter(name string, connector func(context.Context, ...storage.Opt) *storage.DB) *dbPolicyGetter {
	return &dbPolicyGetter{
		connector: connector,
		name:      name,
	}
}

type dbPolicyGetter struct {
	connector func(context.Context, ...storage.Opt) *storage.DB
	name      string
}

func (d dbPolicyGetter) Get(ctx context.Context) (*Policy, error) {
	c := &CleanPolicyConfig{}
	if err := d.connector(ctx).Model(c).First(c).Error; err != nil {
		return nil, err
	}
	return &Policy{
		Retention: c.Retention,
		Capacity:  c.Capacity,
		Rows:      c.Rows,
		WhiteList: strings.Split(c.WhiteList, ";"),
		Name:      d.name,
	}, nil
}
