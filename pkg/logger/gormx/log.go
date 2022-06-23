package gormx

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

func NewLog(l *zap.Logger, debug bool, cfg logger.Config) logger.Interface {
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
		Logger:       l.WithOptions(zap.WithCaller(false)),
		Config:       cfg,
		debug:        debug,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

type gormLog struct {
	*zap.Logger
	logger.Config
	debug                               bool
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

func (g *gormLog) LogMode(level logger.LogLevel) logger.Interface {
	g.LogLevel = level
	return g
}

func (g *gormLog) Info(_ context.Context, msg string, data ...interface{}) {
	if g.LogLevel >= logger.Info {
		g.Logger.Sugar().
			Infof(g.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (g *gormLog) Warn(_ context.Context, msg string, data ...interface{}) {
	if g.LogLevel >= logger.Warn {
		g.Logger.Sugar().
			Warnf(g.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (g *gormLog) Error(_ context.Context, msg string, data ...interface{}) {
	if g.LogLevel >= logger.Error {
		g.Logger.Sugar().
			Errorf(g.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

const NanosecondPerMillisecond = 1e6

func (g *gormLog) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	if g.LogLevel <= logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && g.LogLevel >= logger.Error:
		s, rows := fc()
		if rows == -1 {
			g.Logger.Sugar().
				Errorf(g.traceErrStr, utils.FileWithLineNum(), err,
					float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, "-", s)
		} else {
			g.Logger.Sugar().
				Errorf(g.traceErrStr, utils.FileWithLineNum(), err,
					float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, rows, s)
		}
	case elapsed > g.SlowThreshold && g.SlowThreshold != 0 && g.LogLevel >= logger.Warn:
		s, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", g.SlowThreshold)
		if rows == -1 {
			g.Logger.Sugar().
				Warnf(g.traceWarnStr, utils.FileWithLineNum(), slowLog,
					float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, "-", s)
		} else {
			g.Logger.Sugar().
				Warnf(g.traceWarnStr, utils.FileWithLineNum(), slowLog,
					float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, rows, s)
		}
	case g.LogLevel == logger.Info:
		s, rows := fc()
		if strings.Contains(s, "SELECT") {
			if g.debug {
				if rows == -1 {
					g.Logger.Sugar().
						Infof(g.traceStr, utils.FileWithLineNum(),
							float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, "-", s)
				} else {
					g.Logger.Sugar().
						Infof(g.traceStr, utils.FileWithLineNum(),
							float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, rows, s)
				}
			}
		} else {
			if rows == -1 {
				g.Logger.Sugar().
					Infof(g.traceStr, utils.FileWithLineNum(),
						float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, "-", s)
			} else {
				g.Logger.Sugar().
					Infof(g.traceStr, utils.FileWithLineNum(),
						float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, rows, s)
			}
		}
	}
}
