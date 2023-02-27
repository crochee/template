package response

import (
	"time"

	"template/internal/model"
)

type QueryAreaListRes struct {
	model.Pagination
	List []*ListQueryAreaList `json:"list"`
}

type ListQueryAreaList struct {
	// ID
	ID string `json:"id"`
	// 创建时间
	CreatedAt time.Time `json:"created_at" time_format:"2006-01-02 15:04:05" time_location:"Asia/Shanghai"`
	// 更新时间
	UpdatedAt time.Time `json:"updated_at" time_format:"2006-01-02 15:04:05" time_location:"Asia/Shanghai"`
	// 区域名称
	AreaName string `json:"area_name"`
	// 区域编码
	AreaCode string `json:"area_code"`
	// 区域描述
	AreaDesc string `json:"area_desc"`
	// 区域使用状态
	Status string `json:"status"`
	// 国家名称
	CountryName string `json:"country_name"`
	// 大区编码
	BigRegionCode string `json:"big_region_code"`
	// 大区名称
	BigRegionName string `json:"big_region_name"`
	// 省份编码
	ProvinceCode string `json:"province_code"`
	// 省份名称
	ProvinceName string `json:"province_name"`
}
