package gormx

import (
	"context"

	"go.uber.org/zap"

	"template/pkg/logger"
)

func NewGormWriterFrom(ctx context.Context) interface {
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
} {
	l := logger.From(ctx).WithOptions(zap.WithCaller(false))
	return l.Sugar()
}
