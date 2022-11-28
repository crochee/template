package task

import (
	"context"
	"testing"
	"time"

	"github.com/crochee/devt/pkg/logger"
)

func TestRetryTask(t *testing.T) {
	ctx := logger.With(context.Background(), logger.New())
	t.Log(Execute(ctx, RetryTask(NewFirst(), WithAttempt(3), WithInterval(time.Second)), nil))
	t.Log(Execute(ctx, RetryTask(NewSecond(), WithPolicy(PolicyRevert), WithAttempt(3)), nil))
}
