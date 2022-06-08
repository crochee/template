package mysql

import (
	"context"

	"github.com/pkg/errors"

	"go_template/internal/code"
	"go_template/internal/model"
	"go_template/pkg/logger"
	"go_template/pkg/storage/mysql"
)

func newResourcePkg(db *mysql.DB) *resourcePkg {
	return &resourcePkg{
		DB: db,
	}
}

type resourcePkg struct {
	*mysql.DB
}

func (r resourcePkg) Create(ctx context.Context, opts *model.ResourcePkgCreateOpts) (string, error) {
	rp := &model.DcsResourcePkg{
		ResourceID:    opts.ResourceID,
		ChargeType:    opts.ChargeType,
		AccountID:     opts.AccountID,
		UserID:        opts.UserID,
		ProductID:     opts.ProductID,
		PkgStatus:     opts.PkgStatus,
		OrderID:       opts.OrderID,
		OrderType:     opts.OrderType,
		ResourceType:  opts.ResourceType,
		Configuration: opts.Configuration,
		ActiveTime:    opts.ActiveTime,
		InactiveTime:  opts.InactiveTime,
	}
	if err := r.With(ctx).Model(rp).Create(rp); err != nil {
		return "", errors.WithStack(code.ErrNoAccount.WithResult(err))
	}
	return rp.PK(), nil
}

func (r resourcePkg) Delete(ctx context.Context, resourceID string) error {
	rp := &model.DcsResourcePkg{}
	if err := r.With(ctx).Model(rp).Where("resource_id= ?", resourceID).Delete(rp); err != nil {
		return errors.WithStack(code.ErrNoAccount.WithResult(err))
	}
	return nil
}

func (r resourcePkg) Update(ctx context.Context, resourceID string, opts map[string]interface{}) error {
	query := r.With(ctx).Model(&model.DcsResourcePkg{}).
		Where("resource_id= ?", resourceID).Updates(opts)
	if err := query.Error; err != nil {
		return errors.WithStack(code.ErrNoAccount.WithResult(err))
	}
	if query.RowsAffected == 0 {
		return errors.WithStack(code.ErrNoUpdate)
	}
	return nil
}

func (r resourcePkg) UpdateWhenNotFail(ctx context.Context, resourceID string, opts map[string]interface{}) error {
	query := r.With(ctx).Model(&model.DcsResourcePkg{}).
		Where("resource_id= ? AND pkg_status != ?", resourceID, model.OpenFail).Updates(opts)
	if err := query.Error; err != nil {
		return errors.WithStack(code.ErrNoAccount.WithResult(err))
	}
	if query.RowsAffected == 0 {
		return errors.WithStack(code.ErrNoUpdate)
	}
	return nil
}

func (r resourcePkg) UpdateWhenSuccess(ctx context.Context, resourceID string, opts map[string]interface{}) error {
	query := r.With(ctx).Model(&model.DcsResourcePkg{}).
		Where("resource_id= ? AND pkg_status = ?", resourceID, model.OpenSuccess).Updates(opts)
	if err := query.Error; err != nil {
		return errors.WithStack(code.ErrNoAccount.WithResult(err))
	}
	if query.RowsAffected == 0 {
		return errors.WithStack(code.ErrNoUpdate)
	}
	return nil
}

func (r resourcePkg) ExistSuccess(ctx context.Context, accountID string) error {
	var count int64
	if err := r.With(ctx).Model(&model.DcsResourcePkg{}).Select("id").
		Where("charge_type= ? and pkg_status= ? and account_id= ?",
			model.ChargeByVolume, model.OpenSuccess, accountID).Count(&count).Error; err != nil {
		logger.From(ctx).Error(err.Error())
		return errors.WithStack(code.ErrNoAccount.WithResult(err))
	}
	if count > 0 {
		return errors.WithStack(code.ErrNoAccount.WithMessage("请勿重复开通服务"))
	}
	return nil
}
