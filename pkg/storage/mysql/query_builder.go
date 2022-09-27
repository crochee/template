package mysql

import (
	"context"

	"gorm.io/gorm"
)

type sqlOption struct {
	accountID string
}

type SQLOption func(*sqlOption)

func WithAccountID(accountID string) SQLOption {
	return func(option *sqlOption) {
		option.accountID = accountID
	}
}

// SQLBuilder 将参数组装成 gorm.DB 即预处理的sql语句
type SQLBuilder interface {
	Build(ctx context.Context, query *gorm.DB, opts ...SQLOption) *gorm.DB
}

type Header struct {
	AccountID string `header:"X-Account-ID" json:"account_id"`
}

func (h Header) Build(ctx context.Context, query *gorm.DB, opts ...SQLOption) *gorm.DB {
	o := &sqlOption{}

	for _, f := range opts {
		f(o)
	}
	if o.accountID != "" {
		h.AccountID = o.accountID
	}

	if h.AccountID != "" {
		query = query.Where("account_id = ?", h.AccountID)
	}
	return query
}
