package request

import (
	"template/internal/model"
)

type QueryAreaListReq struct {
	model.ListQuery
	AreaName      string `json:"area_name" form:"area_name"`
	AreaCode      string `json:"area_code" form:"area_code"`
	BigRegionCode string `json:"big_region_code" form:"big_region_code"`
	BigRegionName string `json:"big_region_name" form:"big_region_name"`
	ProvinceCode  string `json:"province_code" form:"province_code"`
	ProvinceName  string `json:"province_name" form:"province_name"`
}
