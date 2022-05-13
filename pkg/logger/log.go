package logger

import (
	"context"

	"go.uber.org/zap"
)

type logKey struct{}

func From(ctx context.Context) *zap.Logger {
	l, ok := ctx.Value(logKey{}).(*zap.Logger)
	if !ok {
		return zap.NewNop()
	}
	return l
}

func With(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, logKey{}, l)
}
