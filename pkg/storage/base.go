package storage

import (
	"strconv"
	"time"

	"gorm.io/gorm"

	"template/pkg/idx"
	"template/pkg/utils/v"
)

type SnowID struct {
	ID uint64 `json:"id,string" gorm:"primary_key:id"`
}

func (b *SnowID) BeforeCreate(db *gorm.DB) error {
	if b.ID == 0 {
		snowID, err := idx.NextID()
		if err != nil {
			return err
		}
		b.ID = snowID
	}
	return nil
}

type Time struct {
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;not null;comment:创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;not null;comment:更新时间"`
	DeletedAt DeletedAt `json:"-" gorm:"column:deleted_at;index;comment:删除时间"`
}

type Base struct {
	SnowID
	Time
}

func (b *Base) BeforeCreate(db *gorm.DB) error {
	return b.SnowID.BeforeCreate(db)
}

func (b *Base) PK() string {
	return strconv.FormatUint(b.ID, v.Decimal)
}

func (b *Base) WithPK(id string) {
	if id == "" {
		return
	}
	parseUint, err := strconv.ParseUint(id, v.Decimal, 64)
	if err != nil {
		panic(err)
	}
	b.ID = parseUint
}

type BaseCaller struct {
	Base
	Caller
}

func (b *BaseCaller) BeforeCreate(db *gorm.DB) error {
	if err := b.Caller.BeforeCreate(db); err != nil {
		return err
	}
	return b.Base.BeforeCreate(db)
}

type AccountCaller struct {
	BaseCaller
	AccountID string `gorm:"column:account_id;type:varchar(120);comment:ACCOUNT ID" json:"account_id"`
	UserID    string `gorm:"column:user_id;type:varchar(120);comment:用户ID" json:"user_id"`
}
