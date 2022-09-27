package mysql

import (
	"strconv"
	"time"

	"gorm.io/gorm"

	"go_template/pkg/idx"
	"go_template/pkg/utils/v"
)

type Time struct {
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;not null;default:current_timestamp();comment:创建时间"`
	// nolint:lll
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;not null;default:current_timestamp() on update current_timestamp();comment:更新时间"`
	DeletedAt DeletedAt `json:"-" gorm:"column:deleted_at;index;comment:删除时间"`
}

type Base struct {
	ID uint64 `json:"id,string" gorm:"primary_key:id"`
	Time
}

func (b *Base) BeforeCreate(db *gorm.DB) error {
	if b.ID == 0 {
		snowID, err := idx.NextID()
		if err != nil {
			return err
		}
		b.ID = snowID
	}
	return nil
}

func (b *Base) PK() string {
	return strconv.FormatUint(b.ID, v.Decimal)
}
