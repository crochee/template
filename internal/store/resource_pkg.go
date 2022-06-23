package store

import (
	"context"

	"github.com/crochee/devt/internal/model"
)

type ResourcePkgStore interface {
	Create(ctx context.Context, opts *model.ResourcePkgCreateOpts) (string, error)
	Delete(ctx context.Context, resourceID string) error
	Update(ctx context.Context, resourceID string, opts map[string]interface{}) error
	UpdateWhenNotFail(ctx context.Context, resourceID string, opts map[string]interface{}) error
	UpdateWhenSuccess(ctx context.Context, resourceID string, opts map[string]interface{}) error
	ExistSuccess(ctx context.Context, accountID string) error
}
