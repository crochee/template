package mysql

import (
	"context"

	"github.com/pkg/errors"

	"go_template/internal/code"
	"go_template/internal/model"
	"go_template/pkg/storage/mysql"
)

func newResourceChangeFlow(db *mysql.DB) *resourceChangeFlow {
	return &resourceChangeFlow{
		DB: db,
	}
}

type resourceChangeFlow struct {
	*mysql.DB
}

func (r resourceChangeFlow) Create(ctx context.Context, opts *model.ChangeFlowCreateOpts) (string, error) {
	cf := &model.DcsResourceChangeFlow{
		ResourceID:     opts.ResourceID,
		OrderType:      opts.OrderType,
		PurchaseUnit:   opts.PurchaseUnit,
		PurchaseNumber: opts.PurchaseNumber,
		Reason:         opts.Reason,
	}
	if err := r.With(ctx).Model(cf).Create(cf).Error; err != nil {
		return "", errors.WithStack(code.ErrNoAccount.WithResult(err))
	}
	return cf.PK(), nil
}
