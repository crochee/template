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

func NewSQLBuilder(opts ...SQLBuilder) SQLBuilder {
	return sqlBuilders(opts)
}

type sqlBuilders []SQLBuilder

func (s sqlBuilders) Build(ctx context.Context, query *gorm.DB, opts ...SQLOption) *gorm.DB {
	for _, builder := range s {
		query = builder.Build(ctx, query, opts...)
	}
	return query
}
