package model

import (
	"go_template/pkg/storage/mysql"
)

type DcsResourceChangeFlow struct {
	mysql.Base

	ResourceID string `json:"resource_id" gorm:"type:varchar(255);not null;index:idx_resource_id;comment:资源id，按照一定的业务规则，生成唯一标识"`
	// nolint:lll
	OrderType      string `json:"order_type" gorm:"not null;type:varchar(255);comment:工单类型,CREATE代表创建,RENEW代表续订,UPGRADE代表升配,DOWNGRADE代表降配,DESTROY代表销毁"`
	PurchaseUnit   int    `json:"purchase_unit" gorm:"not null;comment:购买单位"`
	PurchaseNumber int    `json:"purchase_number" gorm:"not null;comment:购买数量"`
	Reason         string `json:"reason" gorm:"comment:资源包开通失败原因"`
}

type ChangeFlowCreateOpts struct {
	ResourceID     string
	OrderType      string
	PurchaseUnit   int
	PurchaseNumber int
	Reason         string
}

type ChangeFlowUpdateOpts struct {
}
