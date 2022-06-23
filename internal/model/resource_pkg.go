package model

import (
	"time"

	"github.com/crochee/devt/pkg/storage/mysql"
)

type DcsResourcePkg struct {
	mysql.Base
	ResourceID string `json:"resource_id" gorm:"type:varchar(255);index:idx_resource_id_deleted,unique;not null;comment:资源id，按照一定的业务规则，生成唯一标识"`
	ChargeType uint8  `json:"charge_type" gorm:"not null;comment:资源包开通类型，1包年包月，2按量计费"`
	AccountID  string `json:"account_id" gorm:"not null;index:idx_account_id;type:varchar(255);comment:主账号ID"`
	UserID     string `json:"user_id" gorm:"not null;type:varchar(255);comment:用户ID"`
	ProductID  string `json:"product_id" gorm:"type:varchar(100);comment:产品编号"`
	PkgStatus  uint8  `json:"pkg_status" gorm:"not null;comment:资源包开通状态 0开通失败,1开通成功,2过期"`
	OrderID    string `json:"order_id" gorm:"not null;type:varchar(255);comment:工单唯一标识"`
	// nolint:lll
	OrderType     string    `json:"order_type" gorm:"not null;type:varchar(255);comment:工单类型,CREATE代表创建,RENEW代表续订,UPGRADE代表升配,DOWNGRADE代表降配,DESTROY代表销毁"`
	ResourceType  int       `json:"resource_type" gorm:"not null;comment:工单资源类型，5784代表边缘云开通唯一标识"`
	Configuration string    `json:"configuration" gorm:"not null;type:json;comment:工单资源构建json"`
	ActiveTime    time.Time `json:"active_time" gorm:"comment:生效时间"`
	InactiveTime  time.Time `json:"inactive_time" gorm:"comment:失效时间"`

	Deleted mysql.Deleted `json:"deleted" gorm:"not null;index:idx_resource_id_deleted,unique;comment:软删除记录id"`
}

func (DcsResourcePkg) TableName() string {
	return "dcs_resource_pkg"
}

const (
	ChargeByVolume = 2

	OpenFail uint8 = iota
	OpenSuccess
	PkgExpire
)

type ResourcePkgCreateOpts struct {
	ResourceID    string
	ChargeType    uint8
	AccountID     string
	UserID        string
	ProductID     string
	PkgStatus     uint8
	OrderID       string
	OrderType     string
	ResourceType  int
	Configuration string
	ActiveTime    time.Time
	InactiveTime  time.Time
}

type ResourcePkgUpdateOpts struct {
}
