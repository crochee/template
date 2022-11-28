package mysql

import (
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/crochee/devt/pkg/idx"
	"github.com/crochee/devt/pkg/utils/v"
)

type Time struct {
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;not null;default:current_timestamp();comment:创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;not null;default:current_timestamp() on update current_timestamp();comment:更新时间"`
	DeletedAt DeletedAt `json:"-" gorm:"column:deleted_at;index;comment:删除时间"`
}

type Base struct {
	ID uint64 `json:"id,string" gorm:"primary_key:id"`
	Time
}

type AccountUser struct {
	Base
	UserId    string `gorm:"column:user_id;type:varchar(120);comment:用户ID" json:"user_id"`
	AccountId string `gorm:"column:account_id;type:varchar(120);comment:ACCOUNT ID" json:"account_id"`
}

type ProjectAndOrganization struct {
	AccountUser
	OrganizationId string `gorm:"column:organization_id;type:varchar(120);comment:组织机构ID" json:"organization_id"`
	ProjectId      string `gorm:"column:project_id;type:varchar(120);comment:项目ID" json:"project_id"`
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

func (b *Base) WithPK(id string) {
	if id == "" {
		return
	}
	parseUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		panic(err)
	}
	b.ID = parseUint
}
