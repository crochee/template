package user

import (
	"context"

	"github.com/crochee/lirity/logger"

	"github.com/crochee/devt/internal/store"
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

type User struct {
}

type UserSrv interface {
	Create(ctx context.Context, opts *CreateOpts) (string, error)
	Delete(ctx context.Context, id string, opts *DeleteOpts) error
	Update(ctx context.Context, id string, opts *UpdateOpts) error
	Get(ctx context.Context, id string, opts *GetOpts) (*User, error)
	List(ctx context.Context, opts *ListOpts) ([]*User, error)
}

func NewUserService() UserSrv {
	return &userService{}
}

var _ UserSrv = (*userService)(nil)

type userService struct {
}

func (u *userService) Create(ctx context.Context, opts *CreateOpts) (string, error) {
	return store.Client().Users().Create(ctx, &store.CreateOpts{})
}

func (u *userService) Delete(ctx context.Context, id string, opts *DeleteOpts) error {
	return store.Client().Users().Delete(ctx, id, &store.DeleteOpts{})
}

func (u *userService) Update(ctx context.Context, id string, opts *UpdateOpts) error {
	return store.Client().Users().Update(ctx, id, &store.UpdateOpts{})
}

func (u *userService) Get(ctx context.Context, id string, opts *GetOpts) (*User, error) {
	user, err := store.Client().Users().Get(ctx, id, &store.GetOpts{})
	if err != nil {
		return nil, err
	}
	logger.From(ctx).Sugar().Debugf("%#v", user)
	return &User{}, nil
}

func (u *userService) List(ctx context.Context, opts *ListOpts) ([]*User, error) {
	user, err := store.Client().Users().List(ctx, &store.ListOpts{})
	if err != nil {
		return nil, err
	}
	logger.From(ctx).Sugar().Debugf("%#v", user)
	return []*User{}, nil
}
