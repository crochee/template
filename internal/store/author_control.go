package store

import (
	"context"

	"go_template/internal/model"
)

type AuthorControlStore interface {
	Create(ctx context.Context, opts *model.AuthorControlCreateOpts) (string, error)
	Update(ctx context.Context, accountID string, opts map[string]interface{}) error
	Get(ctx context.Context, accountID string, selects ...string) (*model.DcsAuthorControl, error)
}
