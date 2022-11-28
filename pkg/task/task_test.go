package task

import (
	"context"
	"log"
	"time"

	"github.com/pkg/errors"
)

func NewFirst() Task {
	return &taskFirst{
		DefaultTaskInfo(""),
	}
}

type taskFirst struct {
	StoreInfo
}

func (t taskFirst) Commit(ctx context.Context, input interface{}, callbacks ...Callback) error {
	log.Println("first commit")
	return nil
}

func (t taskFirst) Rollback(ctx context.Context, input interface{}, callbacks ...Callback) error {
	log.Println("first rollback")
	return nil
}

func NewSecond() Task {
	return &taskSecond{
		DefaultTaskInfo(""),
	}
}

type taskSecond struct {
	StoreInfo
}

func (t taskSecond) Commit(ctx context.Context, input interface{}, callbacks ...Callback) error {
	time.Sleep(time.Second)
	return errors.New("second commit failed")
}

func (t taskSecond) Rollback(ctx context.Context, input interface{}, callbacks ...Callback) error {
	log.Println("second rollback")
	return nil
}
