package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"

	"template/pkg/timex"
)

var _ = func() struct{} {
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().In(timex.CST)
	}
	return struct{}{}
}()

type Logger struct {
	zerolog.Logger
}

func (l *Logger) Disable() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.Info().Msgf(format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Warn().Msgf(format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Warn().Msgf(format, args...)
}

// NewZeroLogger returns a zerolog logger with as much context as possible
func NewZeroLogger(opts ...Option) *Logger {
	opt := &option{
		level:  zerolog.InfoLevel.String(),
		writer: os.Stdout,
	}
	for _, o := range opts {
		o(opt)
	}
	l := newZeroLevel(opt.level)
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.SetGlobalLevel(l)
	return &Logger{
		Logger: zerolog.New(opt.writer).
			Level(l).With().
			Str("service_name", opt.serverName).
			Timestamp().Caller().Logger(),
	}
}

func newZeroLevel(level string) zerolog.Level {
	l, err := zerolog.ParseLevel(level)
	if err != nil {
		l = zerolog.InfoLevel
	}
	return l
}