package area

import (
	"context"

	"go.uber.org/zap"

	"template/internal/gateway"
	"template/internal/model"
	"template/internal/request"
	"template/internal/response"
	"template/internal/store"
	"template/pkg/logger"
)

type AreaSrv interface {
	List(ctx context.Context, req *request.QueryAreaListReq) (*response.QueryAreaListRes, error)
}

func NewAreaSrv(store store.Store, client gateway.Client) AreaSrv {
	return areaSrv{
		store:  store,
		client: client,
	}
}

type areaSrv struct {
	store  store.Store
	client gateway.Client
}

func (a areaSrv) List(ctx context.Context, req *request.QueryAreaListReq) (*response.QueryAreaListRes, error) {
	areas, err := a.store.Area().List(ctx, req)
	if err != nil {
		logger.From(ctx).Error("The database failed to query the area list",
			zap.Any("param", req), zap.Error(err))
		return nil, err
	}

	results := response.QueryAreaListRes{
		Pagination: model.Pagination{
			PageNum:  req.PageNum,
			PageSize: req.PageSize,
			Total:    req.Total,
		}}
	results.List = make([]*response.ListQueryAreaList, len(areas))
	for i, area := range areas {
		results.List[i] = &response.ListQueryAreaList{
			ID:            area.PK(),
			CreatedAt:     area.CreatedAt,
			UpdatedAt:     area.UpdatedAt,
			AreaName:      area.AreaName,
			AreaCode:      area.AreaCode,
			AreaDesc:      area.AreaDesc,
			Status:        area.Status,
			CountryName:   area.CountryName,
			BigRegionCode: area.BigRegionCode,
			BigRegionName: area.BigRegionName,
			ProvinceCode:  area.ProvinceCode,
			ProvinceName:  area.ProvinceName,
		}
	}

	return &results, nil
}
