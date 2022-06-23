package store

import (
	"context"

	"github.com/crochee/devt/internal/model"
)

type AuthorControlStore interface {
	Create(ctx context.Context, opts *model.AuthorControlCreateOpts) (string, error)
	Update(ctx context.Context, accountID string, opts map[string]interface{}) error
	Get(ctx context.Context, accountID string, selects ...string) (*model.DcsAuthorControl, error)
}
