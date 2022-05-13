package logger

import (
	"os"

	"go-template/pkg/logger/console"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	DEBUG  = "DEBUG"
	INFO   = "INFO"
	WARN   = "WARN"
	ERROR  = "ERROR"
	DPanic = "DPANIC"
	PANIC  = "PANIC"
	FATAL  = "FATAL"
)

func New(opts ...Option) *zap.Logger {
	o := &option{
		level:   zapcore.InfoLevel,
		encoder: console.NewConsoleEncoder,
		writer:  os.Stdout,
	}
	for _, opt := range opts {
		opt(o)
	}

	core := zapcore.NewCore(
		o.encoder(newEncoderConfig()),
		zap.CombineWriteSyncers(zapcore.AddSync(o.writer)),
		o.level,
	).With(o.fields) // 自带node 信息
	// 大于error增加堆栈信息
	return zap.New(core).WithOptions(zap.AddCaller(), zap.AddCallerSkip(o.skip),
		zap.AddStacktrace(zapcore.DPanicLevel))
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
	l := zap.InfoLevel
	if temp, ok := map[string]zapcore.Level{
		DEBUG:  zap.DebugLevel,
		INFO:   zap.InfoLevel,
		WARN:   zap.WarnLevel,
		ERROR:  zap.ErrorLevel,
		DPanic: zap.DPanicLevel,
		PANIC:  zap.PanicLevel,
		FATAL:  zap.FatalLevel,
	}[level]; ok {
		l = temp
	}
	return l
}
