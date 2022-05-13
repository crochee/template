package logger

import (
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type option struct {
	skip    int
	level   zapcore.Level
	encoder func(zapcore.EncoderConfig) zapcore.Encoder
	writer  io.Writer
	fields  []zap.Field
}

type Option func(*option)

func WithSkip(skip int) Option {
	return func(o *option) {
		o.skip = skip
	}
}

func WithLevel(level string) Option {
	return func(o *option) {
		o.level = newLevel(level)
	}
}

func WithEncoder(encoder func(zapcore.EncoderConfig) zapcore.Encoder) Option {
	return func(o *option) {
		o.encoder = encoder
	}
}

func WithWriter(w io.Writer) Option {
	return func(o *option) {
		o.writer = w
	}
}

func WithFields(fields ...zap.Field) Option {
	return func(o *option) {
		o.fields = fields
	}
}
