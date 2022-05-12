package logger

import (
	"context"

	"github.com/rs/zerolog"
)

type logKey struct{}

func From(ctx context.Context) *zerolog.Logger {
	l, ok := ctx.Value(logKey{}).(*zerolog.Logger)
	if !ok {
		nop := zerolog.Nop()
		return &nop
	}
	return l
}

func With(ctx context.Context, l *zerolog.Logger) context.Context {
	return context.WithValue(ctx, logKey{}, l)
}
