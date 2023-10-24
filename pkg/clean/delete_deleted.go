package clean

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"

	"template/pkg/idx"
	"template/pkg/logger"
	"template/pkg/storage"
)

type option struct {
	connector func(context.Context) *storage.DB
	policy    PolicyGetter
}

type Option func(*option)

func WithPolicy(policy PolicyGetter) Option {
	return func(o *option) {
		o.policy = policy
	}
}

func WithConnector(connector func(context.Context) *storage.DB) Option {
	return func(o *option) {
		o.connector = connector
	}
}

func NewDeleteDeletedJob(opts ...Option) *deleteDeletedJob {
	o := option{
		policy: &Policy{
			Retention: 60 * 24 * time.Hour, //  60天
			Capacity:  200,                 //  200M
			Rows:      1000000,             // 100w
			WhiteList: []string{},
		},
	}
	for _, opt := range opts {
		opt(&o)
	}
	return &deleteDeletedJob{
		option: o,
	}
}

type deleteDeletedJob struct {
	option
}

func (d deleteDeletedJob) Description() string {
	return "delete deleted job"
}

// Key returns the unique key for the Job.
func (d deleteDeletedJob) Key() string {
	id, err := idx.NextID()
	if err != nil {
		return "1"
	}
	return strconv.FormatUint(id, 10)
}

// Execute is called by a SchedulerRuntime when the Trigger associated with this job fires.
func (d deleteDeletedJob) Execute(ctx context.Context) {
	if d.connector == nil {
		logger.From(ctx).Error("connector is nil")
		return
	}
	policy, err := d.policy.Get(ctx)
	if err != nil {
		logger.From(ctx).Error("", zap.Error(err))
		return
	}
	if policy.Name == "" {
		logger.From(ctx).Error("name is  nil")
		return
	}
	deadlineTime := time.Now().UTC().Add(-policy.Retention)
	var hasDeletedTables []string
	if err := d.connector(ctx).
		Raw("SELECT table_name  FROM information_schema.columns  WHERE table_schema = ? AND COLUMN_NAME ='deleted'",
			policy.Name).Find(&hasDeletedTables).Error; err != nil {
		logger.From(ctx).Error("", zap.Error(err))
		return
	}
	var hasDeletedAtTables []string
	if err := d.connector(ctx).
		Raw("SELECT table_name  FROM information_schema.columns  WHERE table_schema = ? AND COLUMN_NAME ='deleted_at'",
			policy.Name).Find(&hasDeletedAtTables).Error; err != nil {
		logger.From(ctx).Error("", zap.Error(err))
		return
	}

	var notOverLimitTables []string
	if err := d.connector(ctx).
		Raw(`SELECT a.table_name FROM (
            SELECT
            table_name,
            table_rows,
            TRUNCATE ( data_length / 1024 / 1024, 2 ) AS data_cap
            FROM information_schema.tables WHERE table_schema = ?) AS a WHERE a.data_cap <  ?  AND  a.table_rows < ?;`,
			policy.Name, policy.Capacity, policy.Rows).Find(&notOverLimitTables).Error; err != nil {
		logger.From(ctx).Error("", zap.Error(err))
		return
	}
	var needDeletedAtTables []string
	for _, v := range hasDeletedAtTables {
		var contain bool
		for _, value := range hasDeletedTables {
			if v == value {
				contain = true
			}
		}
		if !contain {
			needDeletedAtTables = append(needDeletedAtTables, v)
		}
	}
	//   判断需要跳过清理的表名
	skipTableName := func(name string) bool {
		for _, v := range policy.WhiteList {
			if v == name {
				return true
			}
		}
		for _, v := range notOverLimitTables {
			if v == name {
				return true
			}
		}
		return false
	}
	//   start  delete
	for _, name := range hasDeletedTables {
		if skipTableName(name) {
			continue
		}
		if err := d.connector(ctx).Exec(fmt.Sprintf("DELETE FROM %s.%s  WHERE deleted != 0  AND deleted_at <  ?",
			policy.Name, name),
			deadlineTime).Error; err != nil {
			logger.From(ctx).Error("", zap.Error(err))
			return
		}
	}
	for _, name := range needDeletedAtTables {
		if skipTableName(name) {
			continue
		}
		if err := d.connector(ctx).Exec(fmt.Sprintf("DELETE FROM %s.%s  WHERE deleted_at IS NOT NULL  AND deleted_at <  ?",
			policy.Name, name),
			deadlineTime).Error; err != nil {
			logger.From(ctx).Error("", zap.Error(err))
			return
		}
	}
}
