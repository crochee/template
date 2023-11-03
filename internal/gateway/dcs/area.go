package dcs

import (
	"context"
	"net/http"
	"time"

	"template/pkg/client"
)

type AreaSrv interface {
	List(ctx context.Context, param *QueryAreaListParam) (*GetAreasDetailRsp, error)
}

type AreaClient struct {
	client.IRequest
}

type GetAreasDetailRsp struct {
	List     []*Area `json:"list"`
	PageNum  int     `json:"page_num"`
	PageSize int     `json:"page_size"`
	Total    int     `json:"total"`
}

type Area struct {
	// ID
	ID uint64 `json:"id"`
	// IAM主账号id
	AccountID string `json:"account_id"`
	// IAM主账号id
	UserID string `json:"user_id"`
	// 区域名称
	AreaName string `json:"area_name"`
	// 区域编码
	AreaCode string `json:"area_code"`
	// 区域描述
	AreaDesc string `json:"area_desc"`
	// 国家名称
	CountryName string `json:"country_name"`
	// 大区域编码
	BigRegionCode string `json:"big_region_code"`
	// 大区域名称
	BigRegionName string `json:"big_region_name"`
	// 省份编码
	ProvinceCode string `json:"province_code"`
	// 省份名称
	ProvinceName string `json:"province_name"`
	// 创建时间
	CreatedAt time.Time `json:"created_at"`
	// 更新时间
	UpdatedAt time.Time `json:"updated_at"`
	// 区域下站点数量
	SiteCount int `json:"site_count"`
	// 状态
	Status string `json:"status"`
	// 站点信息
	Sites []Site `json:"site"`
}

type Site struct {
	// 边缘站点ID
	ID uint64 `json:"site_id"`
	// 边缘站点名称
	Name string `json:"site_name"`
	// 站点网络类型
	SiteNets []SiteNet `json:"site_nets"`
}

type SiteNet struct {
	// 站点网络类型ID
	ID uint64 `json:"id"`
	// 网络类型, ChinaUnicom
	NetType string `json:"net_type"`
}

type QueryAreaListParam struct {
	// 查询第几页
	// Example: 1
	PageNum int `form:"page_num,default=1" json:"page_num" binding:"omitempty,min=0"`
	// 查询每页显示条目
	// Example: 100
	PageSize int `form:"page_size,default=20" json:"page_size" binding:"omitempty,min=-1"`
	// 多个主机类型时通过，进行分割，如 flavor_types=S,M,L,KS,KM,EN,B
	FlavorTypes string `json:"flavor_types" form:"flavor_types" binding:"omitempty,oneof=S M L KS KM EN B"`
	// 系统盘类型，多个值通过英文逗号进行分割，枚举值： efficiency, ssd
	SysVolumeTypes string `json:"sys_volume_types" form:"sys_volume_types" binding:"omitempty,oneof=efficiency ssd"`
	// 数据盘类型，多个值通过英文逗号进行分割，枚举值： efficiency, ssd
	VolumeTypes string `json:"volume_types" form:"volume_types" binding:"omitempty,oneof=efficiency ssd"`
	AreaName    string `json:"search_by_name"`
	AreaCode    string `json:"area_code"`
}

func (a AreaClient) List(ctx context.Context, param *QueryAreaListParam) (*GetAreasDetailRsp, error) {
	var result GetAreasDetailRsp
	if err := a.To().WithRequest(client.ModifiableRequest{
		ModifyRequest: func(*http.Request) {
		},
		Req: client.NewCoPartner("ak", "sk"),
	}).
		WithResponse(client.Parser{}).
		Method(http.MethodGet).
		Prefix("v2").
		Param("page_num", param.PageNum).
		Param("page_size", param.PageSize).
		Param("flavor_types", param.FlavorTypes).
		Param("sys_volume_types", param.SysVolumeTypes).
		Param("volume_types", param.VolumeTypes).
		Param("search_by_name", param.AreaName).
		Param("area_code", param.AreaCode).
		Do(ctx, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
