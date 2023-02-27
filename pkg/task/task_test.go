package task

import (
	"context"
	"log"
	"time"

	"github.com/pkg/errors"

	"anchor/pkg/logger"
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
	log.Println("second commit")
	return errors.New("second commit failed")
}

func (t taskSecond) Rollback(ctx context.Context, input interface{}, callbacks ...Callback) error {
	log.Println("second rollback")
	return nil
}

func funcT(ctx context.Context, input interface{}) error {
	logger.From(ctx).Info("test")
	log.Println("test")
	return errors.New("test Func")
}
