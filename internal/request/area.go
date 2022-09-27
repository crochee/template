package request

import (
	"go_template/internal/model"
)

type QueryAreaListReq struct {
	model.Pagination
	// 多个主机类型时通过，进行分割，如 flavor_types=S,M,L,KS,KM,EN,B
	FlavorTypes string `json:"flavor_types" form:"flavor_types" binding:"omitempty,oneof=S M L KS KM EN B"`
	// 系统盘类型，多个值通过英文逗号进行分割，枚举值： efficiency, ssd
	SysVolumeTypes string `json:"sys_volume_types" form:"sys_volume_types" binding:"omitempty,oneof=efficiency ssd"`
	// 数据盘类型，多个值通过英文逗号进行分割，枚举值： efficiency, ssd
	VolumeTypes string `json:"volume_types" form:"volume_types" binding:"omitempty,oneof=efficiency ssd"`
}
