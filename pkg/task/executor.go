package task

import (
	"context"

	"go.uber.org/multierr"
	"go.uber.org/zap"

	"template/pkg/logger"
)

// Execute 在执行阶段不允许任何错误
func Execute(ctx context.Context, task Task, input interface{}, callbacks ...Callback) error {
	err := task.Commit(ctx, input, callbacks...)
	if err == nil {
		return nil
	}
	return multierr.Append(err, task.Rollback(ctx, input, callbacks...))
}

// InertExecute 惰性执行器,如果发生错误，它将打印提示，但是会忽略commit的错误
func InertExecute(ctx context.Context, task Task, input interface{}, callbacks ...Callback) error {
	err := task.Commit(ctx, input, callbacks...)
	if err == nil {
		return nil
	}
	logger.From(ctx).Warn("commit failed", zap.Error(err))
	return task.Rollback(ctx, input, callbacks...)
}
