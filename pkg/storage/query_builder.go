package storage

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"template/pkg/utils/v"
)

type SqlOption struct {
	AccountID string
	UserID    string
}

type SQLOption func(*SqlOption)

func WithAccountID(accountID string) SQLOption {
	return func(option *SqlOption) {
		option.AccountID = accountID
	}
}

func WithUserID(userID string) SQLOption {
	return func(option *SqlOption) {
		option.UserID = userID
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

// Pagination 分页
type Pagination struct {
	// 查询第几页
	// Example: 1
	PageNum int `form:"page_num,default=1" json:"page_num" binding:"omitempty,min=0"`
	// 查询每页显示条目
	// Example: 100
	PageSize int `form:"page_size,default=20" json:"page_size" binding:"omitempty,min=-1"`
	// 总计条目
	// Example: 300
	Total int64 `form:"-" json:"total"`
}

func (p *Pagination) Build(_ context.Context, query *gorm.DB, _ ...SQLOption) *gorm.DB {
	if p.Total == 0 {
		query.Count(&p.Total)
	}
	// -1表示全量查询
	if p.PageSize < 0 {
		return query
	}
	if p.PageNum == 0 {
		p.PageNum = v.DefaultPageNum
	}
	if p.PageSize == 0 {
		p.PageSize = v.DefaultPageSize
	}
	return query.Limit(p.PageSize).Offset((p.PageNum - 1) * p.PageSize)
}

// Sort 排序
type Sort struct {
	// 排序信息【格式:字段 排序方式】,desc-降序,asc-升序,默认降序排列,例如:[created_at asc]
	// 支持多字段同时排序,例如:[created_at desc,created_at asc]
	// 给多个字段排序 created_at, id asc => order by created_at desc, id asc
	SortField string `form:"sort" json:"-" binding:"omitempty,order"`
}

func (s *Sort) Build(_ context.Context, query *gorm.DB, _ ...SQLOption) *gorm.DB {
	// SortField 给多个字段排序
	// created_at, id asc => order by created_at desc, id asc
	defaultCreatedAtSort := true
	if s.SortField != "" {
		sorts := strings.Split(s.SortField, ",")
		for _, field := range sorts {
			if strings.Contains(field, "created_at") {
				defaultCreatedAtSort = false
			}
			// 如果排序没有明确要按asc或desc来排序，则按照默认排序(倒序)
			if !strings.HasSuffix(field, "asc") && !strings.HasSuffix(field, "desc") {
				query = query.Order(fmt.Sprintf("%s desc", field))
				continue
			}
			query = query.Order(field)
		}
	}
	if defaultCreatedAtSort {
		// 默认按created_at倒序排列
		query = query.Order("created_at desc")
	}
	return query
}

type ListQuery struct {
	Pagination
	Sort
	Select []string `json:"-"`
	IDs    []string `form:"id" binding:"omitempty,number"`
}

func (l *ListQuery) Build(ctx context.Context, query *gorm.DB, opts ...SQLOption) *gorm.DB {
	if length := len(l.Select); length > 0 {
		if length == 1 {
			query = query.Select(l.Select[0])
		} else {
			args := make([]interface{}, 0, length-1)
			for _, selectValue := range l.Select[1:] {
				args = append(args, selectValue)
			}
			query = query.Select(l.Select[0], args...)
		}
	}

	if length := len(l.IDs); length > 0 {
		if length == 1 {
			query = query.Where("id = ?", l.IDs[0])
		} else {
			query = query.Where("id IN (?)", l.IDs)
		}
	}

	query = l.Sort.Build(ctx, query, opts...)
	return l.Pagination.Build(ctx, query, opts...)
}

type WhereOpts struct {
	ID        string
	AccountID string
	Select    []string
}

func (w *WhereOpts) Build(ctx context.Context, query *gorm.DB, opts ...SQLOption) *gorm.DB {
	if w.ID != "" {
		query = query.Where("id = ?", w.ID)
	}
	if w.AccountID != "" {
		query = query.Where("account_id = ?", w.AccountID)
	}
	if length := len(w.Select); length > 0 {
		if length == 1 {
			query = query.Select(w.Select[0])
		} else {
			args := make([]interface{}, 0, length-1)
			for _, selectValue := range w.Select[1:] {
				args = append(args, selectValue)
			}
			query = query.Select(w.Select[0], args...)
		}
	}
	return query
}
