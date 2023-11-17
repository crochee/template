package gormx

import (
	"context"

	"go.uber.org/zap"

	"template/pkg/logger"
)

type Logger interface {
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
}

func NewZapGormWriterFrom(ctx context.Context) Logger {
	l := logger.From(ctx).WithOptions(zap.WithCaller(false))
	return l.Sugar()
}

func NewZeroGormWriterFrom(ctx context.Context) Logger {
	return logger.FromContext(ctx)
}

type nop struct{}

func (nop) Debugf(string, ...interface{}) {}

func (nop) Infof(string, ...interface{}) {}

func (nop) Warnf(string, ...interface{}) {}

func (nop) Errorf(string, ...interface{}) {}

func Nop(context.Context) Logger {
	return nop{}
}
