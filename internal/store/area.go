package store

import (
	"context"

	"template/internal/model"
	"template/internal/request"
)

type AreaStore interface {
	List(ctx context.Context, req *request.QueryAreaListReq) ([]*model.Area, error)
}
