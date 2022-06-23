package mysql

import (
	"context"

	"github.com/pkg/errors"

	"github.com/crochee/devt/internal/code"
	"github.com/crochee/devt/internal/model"
	"github.com/crochee/devt/pkg/storage/mysql"
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
