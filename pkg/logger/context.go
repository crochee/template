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

type zeroLogKey struct{}

// WithContext Adds log to context.Context.
func WithContext(ctx context.Context, log *Logger) context.Context {
	return context.WithValue(ctx, zeroLogKey{}, log)
}

// FromContext Gets the log from context.Context.
func FromContext(ctx context.Context) *Logger {
	l, ok := ctx.Value(zeroLogKey{}).(*Logger)
	if !ok {
		l = &Logger{Logger: zerolog.Nop()}
	}
	return l
}
