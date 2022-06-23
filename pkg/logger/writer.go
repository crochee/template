package logger

import (
	"io"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	LogSizeM   = 20
	MaxZip     = 50
	MaxLogDays = 30
)

// SetWriter return a io.Writer
func SetWriter(path string) io.Writer {
	if path == "" {
		return os.Stdout
	}
	return &lumberjack.Logger{
		Filename:   filepath.Clean(path),
		MaxSize:    LogSizeM,   // 单个日志文件最大MaxSize*M大小
		MaxAge:     MaxLogDays, // days
		MaxBackups: MaxZip,     // 备份数量
		Compress:   false,      // 不压缩
		LocalTime:  true,       // 备份名采用本地时间
	}
}
