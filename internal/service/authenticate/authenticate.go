package authenticate

import (
	"context"

	"github.com/pkg/errors"

	"anchor/internal/code"
	"anchor/internal/model"
	"anchor/internal/store"
)

type AuthenticateSrv interface {
	Authenticate(ctx context.Context, accountID string) error
}

func NewAuthenticateSrv() AuthenticateSrv {
	return authSrv{}
}

type authSrv struct {
}

func (a authSrv) Authenticate(ctx context.Context, accountID string) error {
	ac, err := store.Client().Auth().Get(ctx, accountID, "author_control")
	if err != nil {
		return err
	}
	if ac.AuthorControl == model.ExpirationRestrictedControl {
		return errors.WithStack(code.ErrForbidden.WithResult("用户资源包已到期"))
	}
	if ac.AuthorControl == model.DestructionRestrictedControl {
		return errors.WithStack(code.ErrForbidden.WithResult("用户资源包已删除"))
	}
	return nil
}
