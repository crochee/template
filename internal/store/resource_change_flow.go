package store

import (
	"context"

	"go_template/internal/model"
)

type ChangeFlowStore interface {
	Create(ctx context.Context, opts *model.ChangeFlowCreateOpts) (string, error)
}
