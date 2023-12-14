package logger

import (
	"context"

	"github.com/rs/zerolog"
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

// WithContext Adds log to context.Context.
func WithContext(ctx context.Context, log *Logger) context.Context {
	return (&log.Logger).WithContext(ctx)
}

// FromContext Gets the log from context.Context.
func FromContext(ctx context.Context) *Logger {
	l := zerolog.Ctx(ctx)
	return &Logger{Logger: *l}
}
