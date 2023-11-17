package logger

import (
	"os"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"template/pkg/logger/console"
	"template/pkg/timex"
)

func New(opts ...Option) *zap.Logger {
	o := &option{
		level:   zapcore.InfoLevel.String(),
		encoder: console.NewConsoleEncoder,
		writer:  os.Stdout,
	}
	for _, opt := range opts {
		opt(o)
	}

	core := zapcore.NewCore(
		o.encoder(newEncoderConfig()),
		zap.CombineWriteSyncers(zapcore.AddSync(o.writer)),
		NewChangeLevel(o.level),
	).With(append(o.fields, zap.String("service_name", o.serverName))) // 自带node 信息
	// 大于error增加堆栈信息
	return zap.New(core).WithOptions(zap.AddCaller(), zap.AddCallerSkip(o.skip),
		zap.AddStacktrace(zapcore.DPanicLevel), zap.WithClock(systemClock{}))
}

func newEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		MessageKey:     "Message",
		LevelKey:       "Level",
		TimeKey:        "Time",
		NameKey:        "Logger",
		CallerKey:      "Caller",
		StacktraceKey:  "Stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
}

func newLevel(level string) zapcore.Level {
	l, err := zapcore.ParseLevel(level)
	if err != nil {
		l = zap.InfoLevel
	}
	return l
}

// systemClock implements default Clock that uses system time.
type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now().In(timex.CST)
}

func (systemClock) NewTicker(duration time.Duration) *time.Ticker {
	return time.NewTicker(duration)
}

func NewChangeLevel(level string) *changeLevel {
	return &changeLevel{
		level: newLevel(level),
	}
}

type changeLevel struct {
	level zapcore.Level
}

func (ch *changeLevel) Enabled(lvl zapcore.Level) bool {
	if atomic.LoadUint32(&debug) == 1 {
		return true
	}
	return lvl >= ch.level
}
