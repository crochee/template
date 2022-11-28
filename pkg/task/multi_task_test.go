package task

import (
	"context"
	"testing"

	"github.com/crochee/devt/pkg/logger"
)

func TestPipelineTask(t *testing.T) {
	ctx := logger.With(context.Background(), logger.New())
	st := PipelineTask(WithTasks(NewFirst()))
	t.Log(Execute(ctx, st, nil))
}

func TestParallelTask(t *testing.T) {
	ctx := logger.With(context.Background(), logger.New())
	st := ParallelTask(WithTasks(NewFirst(), NewSecond()))
	t.Log(Execute(ctx, st, nil))
}
