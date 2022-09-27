package dcs

import (
	"context"
	"net/http"
	"time"

	"go_template/internal/gateway/base"
	"go_template/internal/request"
	"go_template/pkg/client"
	"go_template/pkg/utils"
)

type AreaSrv interface {
	List(ctx context.Context, req request.QueryAreaListReq) (*GetAreasDetailRsp, error)
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

func (a AreaClient) List(ctx context.Context, req request.QueryAreaListReq) (*GetAreasDetailRsp, error) {
	var result GetAreasDetailRsp
	if err := a.To().WithRequest(base.DCSRequest{}).
		WithResponse(base.Parser{}).
		Method(http.MethodGet).
		Prefix("v2").
		Param("page_num", utils.ToString(req.PageNum)).
		Param("page_size", utils.ToString(req.PageSize)).
		Param("flavor_types", req.FlavorTypes).
		Param("sys_volume_types", req.SysVolumeTypes).
		Param("volume_types", req.VolumeTypes).
		Do(ctx, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
