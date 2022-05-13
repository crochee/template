package store

import (
	"context"

	"go-template/internal/model"
)

type CreateOpts struct {
}

type DeleteOpts struct {
}

type UpdateOpts struct {
}

type GetOpts struct {
}

type ListOpts struct {
}

type UserHandler interface {
	Create(ctx context.Context, opts *CreateOpts) (string, error)
	Delete(ctx context.Context, id string, opts *DeleteOpts) error
	Update(ctx context.Context, id string, opts *UpdateOpts) error
	Get(ctx context.Context, id string, opts *GetOpts) (*model.User, error)
	List(ctx context.Context, opts *ListOpts) ([]*model.User, error)
}
