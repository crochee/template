package task

import (
	"context"
	"testing"
	"time"

	"anchor/pkg/logger"
)

func TestRetryTask(t *testing.T) {
	ctx := logger.With(context.Background(), logger.New())
	t.Log(Execute(ctx, RetryTask(NewFunc(funcT), WithInterval(time.Second), WithAttempt(3)), nil))
	t.Log(Execute(ctx, RetryTask(NewFirst(), WithAttempt(3), WithInterval(time.Second)), nil))
	t.Log(Execute(ctx, RetryTask(NewSecond(), WithAttempt(3), WithInterval(time.Second)), nil))
}
