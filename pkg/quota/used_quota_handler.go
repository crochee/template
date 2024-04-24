package quota

import (
	"context"
	"fmt"

	"golang.org/x/sync/singleflight"
)

type UsedQuotaHandler interface {
	// QueryUsed 查询用户的配额数据和使用量是数据
	QueryUsed(ctx context.Context, associatedID string) (used int, err error)
	// QueryQuota 查询用户的配额数据和使用量是数据
	QueryQuota(ctx context.Context, associatedID string) (quota int, err error)
}

func NewSingleFlightUsedQuotaHandler(
	name string,
	handler UsedQuotaHandler,
) *singleflightUsedQuotaHandler {
	return &singleflightUsedQuotaHandler{
		name:    name,
		handler: handler,
	}
}

var singleFlight singleflight.Group

type singleflightUsedQuotaHandler struct {
	name    string
	handler UsedQuotaHandler
}

func (s *singleflightUsedQuotaHandler) key(queryType, associatedID string) string {
	return fmt.Sprintf("%s:%s:%s", s.name, queryType, associatedID)
}

// QueryUsed 查询用户的配额数据和使用量是数据
func (s *singleflightUsedQuotaHandler) QueryUsed(
	ctx context.Context,
	associatedID string,
) (int, error) {
	value, err, _ := singleFlight.Do(s.key("used", associatedID), func() (interface{}, error) {
		return s.handler.QueryUsed(ctx, associatedID)
	})
	return value.(int), err
}

// QueryQuota 查询用户的配额数据和使用量是数据
func (s *singleflightUsedQuotaHandler) QueryQuota(
	ctx context.Context,
	associatedID string,
) (int, error) {
	value, err, _ := singleFlight.Do(s.key("quota", associatedID), func() (interface{}, error) {
		return s.handler.QueryQuota(ctx, associatedID)
	})
	return value.(int), err
}
