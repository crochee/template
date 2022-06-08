package mysql

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"go_template/internal/code"
	"go_template/internal/model"
	"go_template/pkg/storage/mysql"
)

func newAuthorControl(db *mysql.DB) *authorControl {
	return &authorControl{
		DB: db,
	}
}

type authorControl struct {
	*mysql.DB
}

func (a authorControl) Create(ctx context.Context, opts *model.AuthorControlCreateOpts) (string, error) {
	ac := &model.DcsAuthorControl{
		AccountID:     opts.AccountID,
		AuthorControl: opts.AuthorControl,
	}
	if err := a.WithContext(ctx).Model(ac).Create(ac).Error; err != nil {
		return "", errors.WithStack(code.ErrNoAccount.WithResult(err))
	}
	return ac.PK(), nil
}

func (a authorControl) Update(ctx context.Context, accountID string, opts map[string]interface{}) error {
	query := a.WithContext(ctx).Model(&model.DcsAuthorControl{}).
		Where("account_id =?", accountID).
		Updates(opts)
	if err := query.Error; err != nil {
		return errors.WithStack(code.ErrNoAccount.WithResult(err))
	}
	if query.RowsAffected == 0 {
		return errors.WithStack(code.ErrNoUpdate)
	}
	return nil
}

func (a authorControl) Get(ctx context.Context, accountID string, selects ...string) (*model.DcsAuthorControl, error) {
	var value model.DcsAuthorControl
	query := a.WithContext(ctx).Model(&model.DcsAuthorControl{}).Where("account_id =?", accountID)
	if len(selects) != 0 {
		query = query.Select(selects[0])
	}
	if err := query.First(&value).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, errors.WithStack(code.ErrNoAccount.WithResult(err))
	}
	return &value, nil
}
