package store

import (
	"context"

	"github.com/crochee/devt/internal/model"
)

type ChangeFlowStore interface {
	Create(ctx context.Context, opts *model.ChangeFlowCreateOpts) (string, error)
}
