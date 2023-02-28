package task

import (
	"context"
	"testing"
	"time"

	"template/pkg/logger"
)

func TestPipelineTask(t *testing.T) {
	ctx := logger.With(context.Background(), logger.New())
	st := PipelineTask(WithTasks(RetryTask(
		NewFunc(funcT), WithInterval(time.Second), WithAttempt(3)),
		RetryTask(NewSecond(), WithInterval(time.Second), WithAttempt(3)),
	))
	t.Log(Execute(ctx, st, nil))
}

func TestParallelTask(t *testing.T) {
	ctx := logger.With(context.Background(), logger.New())
	st := ParallelTask(WithTasks(
		RetryTask(NewFunc(funcT), WithInterval(time.Second), WithAttempt(3)),
		RetryTask(NewSecond(), WithInterval(time.Second), WithAttempt(3)),
	))
	t.Log(Execute(ctx, st, nil))
}
