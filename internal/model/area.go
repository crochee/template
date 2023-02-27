package model

import "template/pkg/storage"

// Area 区域信息表
type Area struct {
	storage.Base
	AreaName      string `gorm:"column:area_name;type:varchar(255);comment:区域名称;NOT NULL" json:"area_name"`
	AreaCode      string `gorm:"column:area_code;type:varchar(255);comment:区域编码;NOT NULL;uniqueIndex:idx_area_code" json:"area_code"` // nolint:lll
	AreaDesc      string `gorm:"column:area_desc;type:text;comment:区域描述" json:"area_desc"`
	Status        string `gorm:"column:status;type:varchar(16);comment:区域使用状态" json:"status"`
	CountryName   string `gorm:"column:country_name;type:varchar(30);default:中国;comment:国家名称;NOT NULL" json:"country_name"`
	BigRegionCode string `gorm:"column:big_region_code;type:varchar(30);comment:大区编码" json:"big_region_code"`
	BigRegionName string `gorm:"column:big_region_name;type:varchar(30);comment:大区名称" json:"big_region_name"`
	ProvinceCode  string `gorm:"column:province_code;type:varchar(30);comment:省份编码" json:"province_code"`
	ProvinceName  string `gorm:"column:province_name;type:varchar(30);comment:省份名称" json:"province_name"`

	Deleted storage.Deleted `gorm:"column:deleted;type:bigint(20) unsigned" json:"-"`
}
