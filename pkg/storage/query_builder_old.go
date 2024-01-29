package storage

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"template/pkg/timex"
	"template/pkg/utils"
)

// PrimaryFilter 过滤
// 私有云层次结构和DCS概念映射关系
// account -- 平台租户id（masterTenantId）
// project -- 项目
// user    --- 用户
type PrimaryFilter struct {
	AccountID      string   `header:"X-Account-ID" json:"account_id"`
	AccountIDs     []string `form:"account_id"`
	UserIDs        []string `form:"user_id"`
	OrganizationID string   `form:"organization_id" json:"organization_id"`
	ProjectID      string   `form:"project_id" json:"project_id"`
	CallerID       string   `form:"caller_id" json:"caller_id"`
	CallerCode     string   `form:"caller_code" json:"caller_code"`
}

var (
	getAccountID      func(ctx context.Context) string
	getAdminID        func(ctx context.Context) string
	getAPIFromConsole func(ctx context.Context) bool
	isPrivate         func(ctx context.Context) bool
)

func SetFunc(
	getAccountIDFunc func(ctx context.Context) string,
	getAdminIDFunc func(ctx context.Context) string,
	getAPIFromConsoleFunc func(ctx context.Context) bool,
	isPrivateFunc func(ctx context.Context) bool,
) {
	getAccountID = getAccountIDFunc
	getAdminID = getAdminIDFunc
	getAPIFromConsole = getAPIFromConsoleFunc
	isPrivate = isPrivateFunc
}

func (p *PrimaryFilter) Build(ctx context.Context, query *gorm.DB, opts ...SQLOption) *gorm.DB {
	o := &SqlOption{}
	if getAccountID != nil {
		// 从context中获取accountID并判断赋值
		accountID := getAccountID(ctx)
		if accountID != "" {
			o.AccountID = accountID
		}
	}

	for _, opt := range opts {
		opt(o)
	}
	// 若经过opt执行后，值不为空则重新取值
	if o.AccountID != "" {
		p.AccountID = o.AccountID
	}

	query = p.BuildPrivate(ctx, query, opts...)
	query = p.BuildStaff(ctx, query, opts...)
	query = p.BuildPublic(ctx, query, opts...)

	// Filter by organization id
	if p.OrganizationID != "" {
		query = query.Where("organization_id = ?", p.OrganizationID)
	}

	// Filter by project id
	if p.ProjectID != "" {
		if strings.Contains(p.ProjectID, ",") {
			// 批量的输入格式为1,2,3,4,建议单次批量不超过200,否则应该分批处理
			query = query.Where("project_id IN(?)", strings.Split(p.ProjectID, ","))
		} else {
			query = query.Where("project_id = ?", p.ProjectID)
		}
	}

	// Filter by caller id
	if p.CallerID != "" {
		query = query.Where("caller_id = ?", p.CallerID)
	}
	// Filter by caller name
	if p.CallerCode != "" {
		query = query.Where("caller_code = ?", p.CallerCode)
	}

	return query
}

func (p *PrimaryFilter) BuildPrivate(ctx context.Context, query *gorm.DB, opts ...SQLOption) *gorm.DB {
	if isPrivate != nil && isPrivate(ctx) {
		// 【私有云】
		if getAPIFromConsole != nil && !getAPIFromConsole(ctx) {
			// 不是来源于pokerface的请求参数
			// 适用于pass下的私有云
			// 如果header’s X-Account-ID不为空，则需要把X-Account-ID追加到查询条件中
			if p.AccountID != "" {
				if !utils.InSlice(p.AccountIDs, p.AccountID) {
					p.AccountIDs = append(p.AccountIDs, p.AccountID)
				}
			}
		}
		// Filter by account list
		accountIDsLength := len(p.AccountIDs)
		if accountIDsLength > 0 {
			if accountIDsLength == 1 {
				query = query.Where("account_id = ?", p.AccountIDs[0])
			} else {
				query = query.Where("account_id IN (?)", p.AccountIDs)
			}
		}
	}
	return query
}

func (p *PrimaryFilter) BuildStaff(ctx context.Context, query *gorm.DB, opts ...SQLOption) *gorm.DB {
	if getAdminID != nil && getAdminID(ctx) != "" {
		// 【云警运维账号】
		// 云警运维账号不会传X-Account-ID, 会用到account_ids
		// Filter by account list
		accountIDsLength := len(p.AccountIDs)
		if accountIDsLength > 0 {
			if accountIDsLength == 1 {
				query = query.Where("account_id = ?", p.AccountIDs[0])
			} else {
				query = query.Where("account_id IN (?)", p.AccountIDs)
			}
		}
	}
	return query
}

func (p *PrimaryFilter) BuildPublic(ctx context.Context, query *gorm.DB, opts ...SQLOption) *gorm.DB {
	if isPrivate != nil && !isPrivate(ctx) && getAdminID != nil && getAdminID(ctx) == "" {
		// 【行业云】
		// 如果header’s X-Account-ID不为空，则需要把X-Account-ID追加到查询条件中
		if p.AccountID != "" {
			query = query.Where("account_id = ?", p.AccountID)
		}

		// Filter by user list
		// NOTE(huangt): private cloud doesn't filter by user id.
		userIDsLength := len(p.UserIDs)
		if userIDsLength > 0 {
			if userIDsLength == 1 {
				query = query.Where("user_id = ?", p.UserIDs[0])
			} else {
				query = query.Where("user_id IN(?)", p.UserIDs)
			}
		}
	}
	return query
}

// PrimarySearch 模糊过滤
type PrimarySearch struct {
	// 根据ID进行模糊过滤
	SearchByID string `form:"search_by_id" json:"search_by_id"`
	// 根据名称进行模糊过滤
	SearchByName string `form:"search_by_name" json:"search_by_name"`
}

func (p *PrimarySearch) Build(_ context.Context, query *gorm.DB, _ ...SQLOption) *gorm.DB {
	if p.SearchByID != "" {
		query = query.Where("id like ?", fmt.Sprintf("%%%s%%", EscapeQueryString(p.SearchByID)))
	}
	if p.SearchByName != "" {
		query = query.Where("name like ?", fmt.Sprintf("%%%s%%", EscapeQueryString(p.SearchByName)))
	}
	return query
}

// PrimarySort 排序
type PrimarySort struct {
	// 根据创建时间排序,默认倒排,目前已经新增新的排序方式， 用sort字段来控制排序字段以及倒排还正排，后期考虑遗弃这个方法
	OrderCreatedAt string `form:"order_created_at" json:"order_created_at"`
	// 给多个字段排序 created_at, id asc => order by created_at desc, id asc
	SortField string `form:"sort" json:"sort" binding:"omitempty,order"`
}

func (p *PrimarySort) Build(_ context.Context, query *gorm.DB, _ ...SQLOption) *gorm.DB {
	// SortField 给多个字段排序
	// created_at, id asc => order by created_at desc, id asc
	if p.SortField != "" {
		sorts := strings.Split(p.SortField, ",")
		for _, field := range sorts {
			// 如果排序没有明确要按asc或desc来排序，则按照默认排序(倒序)
			if !strings.HasSuffix(field, "asc") && !strings.HasSuffix(field, "desc") {
				query = query.Order(fmt.Sprintf("%s desc", field))
				continue
			}
			query = query.Order(field)
		}
	}

	if p.OrderCreatedAt == "asc" {
		query = query.Order("created_at asc")
	} else {
		// 默认按created_at倒序排列
		query = query.Order("created_at desc")
	}
	return query
}

// CreatedScope 创建时间范围
type CreatedScope struct {
	// 创建时间查询开始值，时间格式：2006-01-02 15:04:05
	StartCreated string `form:"start_created" json:"start_created" binding:"omitempty,datetime=2006-01-02 15:04:05"`
	// 创建时间查询结束值，时间格式：2006-01-02 15:04:05
	EndCreated string `form:"end_created" json:"end_created" binding:"omitempty,datetime=2006-01-02 15:04:05"`
}

func (c *CreatedScope) Build(_ context.Context, query *gorm.DB, _ ...SQLOption) *gorm.DB {
	// 如果有一条记录创建时间是{2022-11-11 07:17:42.230}，当查询条件start_created & end_created都等于{2022-11-11 07:17:42}时，
	// 无法过滤出这条记录，因此此种情况下直接设置end_created为空
	if c.StartCreated == c.EndCreated {
		c.EndCreated = timex.GetAddTime(c.EndCreated)
	}
	if c.StartCreated != "" {
		query = query.Where("created_at >= ?", timex.GetUTC(c.StartCreated))
	}
	if c.EndCreated != "" {
		query = query.Where("created_at <= ?", timex.GetUTC(c.EndCreated))
	}
	return query
}

type MiniPrimary struct {
	PrimaryFilter
	PrimarySearch
	PrimarySort
	CreatedScope
}

func (p *MiniPrimary) Build(ctx context.Context, query *gorm.DB, opts ...SQLOption) *gorm.DB {
	query = p.PrimaryFilter.Build(ctx, query, opts...)
	query = p.PrimarySearch.Build(ctx, query, opts...)
	query = p.CreatedScope.Build(ctx, query, opts...)
	return p.PrimarySort.Build(ctx, query, opts...)
}

type Primary struct {
	MiniPrimary
	Pagination
}

func (p *Primary) Build(ctx context.Context, query *gorm.DB, opts ...SQLOption) *gorm.DB {
	query = p.MiniPrimary.Build(ctx, query, opts...)
	return p.Pagination.Build(ctx, query, opts...)
}
