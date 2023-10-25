package gormx

import (
	"context"

	"go.uber.org/zap"

	"template/pkg/logger"
)

func NewZapGormWriterFrom(ctx context.Context) interface {
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
} {
	l := logger.From(ctx).WithOptions(zap.WithCaller(false))
	return l.Sugar()
}

func NewZeroGormWriterFrom(ctx context.Context) interface {
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
} {
	return logger.FromContext(ctx)
}
