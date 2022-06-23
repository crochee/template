package mysql

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/crochee/devt/pkg/idx"
)

type Time struct {
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;not null;default:current_timestamp();comment:创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;not null;default:current_timestamp() on update current_timestamp();comment:更新时间"`
	DeletedAt DeletedAt `json:"-" gorm:"column:deleted_at;index;comment:删除时间"`
}

type PrimaryKeyID string

// Scan implements the Scanner interface.
func (p *PrimaryKeyID) Scan(value interface{}) error {
	v, ok := value.(uint64)
	if !ok {
		return fmt.Errorf("%v isn't u64", value)
	}
	*p = PrimaryKeyID(strconv.FormatUint(v, 10))
	return nil
}

// Value implements the driver Valuer interface.
func (p *PrimaryKeyID) Value() (driver.Value, error) {
	return strconv.ParseUint(string(*p), 10, 64)
}

type Base struct {
	ID PrimaryKeyID `json:"id,string" gorm:"primary_key:id"`
	Time
}

func (b *Base) BeforeCreate(db *gorm.DB) error {
	if len(b.ID) == 0 {
		snowID, err := idx.NextID()
		if err != nil {
			return err
		}
		b.ID = PrimaryKeyID(strconv.FormatUint(snowID, 10))
	}
	return nil
}

func (b *Base) PK() string {
	return string(b.ID)
}
