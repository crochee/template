package model

import (
	"go_template/pkg/storage/mysql"
)

const (
	NoControl                    uint8 = 0
	ExpirationRestrictedControl  uint8 = 1
	DestructionRestrictedControl uint8 = 2
)

type DcsAuthorControl struct {
	mysql.Base
	AccountID     string `json:"account_id" gorm:"type:varchar(255);index:idx_account_id_deleted,unique;not null;comment:主账号ID"`
	AuthorControl uint8  `json:"author_control" gorm:"type:tinyint;not null;default:0;comment:权限控制：0不控制，1到期受限控制，2销毁受限控制"`

	Deleted mysql.Deleted `json:"deleted" gorm:"not null;index:idx_account_id_deleted,unique;comment:软删除记录id"`
}

type AuthorControlCreateOpts struct {
	AccountID     string
	AuthorControl uint8
}
