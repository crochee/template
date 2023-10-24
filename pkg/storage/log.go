package storage

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

func NewLog(writerFrom func(context.Context) interface {
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
}, opts ...func(*logger.Config)) logger.Interface {
	cfg := logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logger.Info,
		IgnoreRecordNotFoundError: false,
		Colorful:                  true,
	}
	for _, o := range opts {
		o(&cfg)
	}

	var (
		infoStr      = "%s\n[info] "
		warnStr      = "%s\n[warn] "
		errStr       = "%s\n[error] "
		traceStr     = "%s\n[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
	)

	if cfg.Colorful {
		infoStr = logger.Green + "%s\n" + logger.Reset + logger.Green + "[info] " + logger.Reset
		warnStr = logger.BlueBold + "%s\n" + logger.Reset + logger.Magenta + "[warn] " + logger.Reset
		errStr = logger.Magenta + "%s\n" + logger.Reset + logger.Red + "[error] " + logger.Reset
		traceStr = logger.Green + "%s\n" + logger.Reset + logger.Yellow + "[%.3fms] " + logger.BlueBold +
			"[rows:%v]" + logger.Reset + " %s"
		traceWarnStr = logger.Green + "%s " + logger.Yellow + "%s\n" + logger.Reset + logger.RedBold + "[%.3fms] " +
			logger.Yellow + "[rows:%v]" + logger.Magenta + " %s" + logger.Reset
		traceErrStr = logger.RedBold + "%s " + logger.MagentaBold + "%s\n" + logger.Reset + logger.Yellow +
			"[%.3fms] " + logger.BlueBold + "[rows:%v]" + logger.Reset + " %s"
	}
	return &gormLog{
		writerFrom:   writerFrom,
		Config:       cfg,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

type gormLog struct {
	writerFrom func(context.Context) interface {
		Infof(string, ...interface{})
		Warnf(string, ...interface{})
		Errorf(string, ...interface{})
	}
	logger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

func (g *gormLog) LogMode(level logger.LogLevel) logger.Interface {
	g.LogLevel = level
	return g
}

func (g *gormLog) Info(ctx context.Context, msg string, data ...interface{}) {
	if g.LogLevel >= logger.Info {
		g.writerFrom(ctx).Infof(g.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (g *gormLog) Warn(ctx context.Context, msg string, data ...interface{}) {
	if g.LogLevel >= logger.Warn {
		g.writerFrom(ctx).Warnf(g.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (g *gormLog) Error(ctx context.Context, msg string, data ...interface{}) {
	if g.LogLevel >= logger.Error {
		g.writerFrom(ctx).Errorf(g.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

const NanosecondPerMillisecond = 1e6

func (g *gormLog) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if g.LogLevel <= logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && g.LogLevel >= logger.Error:
		s, rows := fc()
		if rows == -1 {
			g.writerFrom(ctx).
				Errorf(g.traceErrStr, utils.FileWithLineNum(), err,
					float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, "-", s)
		} else {
			g.writerFrom(ctx).
				Errorf(g.traceErrStr, utils.FileWithLineNum(), err,
					float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, rows, s)
		}
	case elapsed > g.SlowThreshold && g.SlowThreshold != 0 && g.LogLevel >= logger.Warn:
		s, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", g.SlowThreshold)
		if rows == -1 {
			g.writerFrom(ctx).
				Warnf(g.traceWarnStr, utils.FileWithLineNum(), slowLog,
					float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, "-", s)
		} else {
			g.writerFrom(ctx).
				Warnf(g.traceWarnStr, utils.FileWithLineNum(), slowLog,
					float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, rows, s)
		}
	case g.LogLevel == logger.Info:
		s, rows := fc()
		if rows == -1 {
			g.writerFrom(ctx).
				Infof(g.traceStr, utils.FileWithLineNum(),
					float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, "-", s)
		} else {
			g.writerFrom(ctx).
				Infof(g.traceStr, utils.FileWithLineNum(),
					float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, rows, s)
		}
	}
}
