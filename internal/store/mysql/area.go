package mysql

import (
	"context"

	"github.com/pkg/errors"

	"template/internal/model"
	"template/internal/request"
	"template/pkg/code"
	"template/pkg/storage"
)

func newArea(db *storage.DB) *area {
	return &area{
		DB: db,
	}
}

type area struct {
	*storage.DB
}

func (a area) List(ctx context.Context, req *request.QueryAreaListReq) ([]*model.Area, error) {
	var objs []*model.Area
	query := a.With(ctx).Model(&model.Area{})
	if req.AreaName != "" {
		query = query.Where("area_name = ?", req.AreaName)
	}
	if req.AreaCode != "" {
		query = query.Where("area_code = ?", req.AreaCode)
	}
	if req.BigRegionCode != "" {
		query = query.Where("big_region_code = ?", req.BigRegionCode)
	}
	if req.BigRegionName != "" {
		query = query.Where("big_region_name = ?", req.BigRegionName)
	}
	if req.ProvinceCode != "" {
		query = query.Where("province_code = ?", req.ProvinceCode)
	}
	if req.ProvinceName != "" {
		query = query.Where("province_name = ?", req.ProvinceName)
	}
	if err := req.Build(ctx, query).Find(&objs).Error; err != nil {
		return nil, errors.WithStack(code.ErrInternalServerError.WithResult(err.Error()))
	}
	return objs, nil
}
