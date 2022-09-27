package logger

import (
	"io"
	"path/filepath"
	"sync"

	"github.com/mattn/go-colorable"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	loggerWriter     io.WriteCloser
	loggerWriterOnce sync.Once
)

// SetWriter return a io.Writer
func SetWriter(console bool) io.Writer {
	var writerList []io.Writer
	path := viper.GetString("log.path")
	if console {
		writerList = append(writerList, colorable.NewColorableStdout())
		if path == "" {
			return io.MultiWriter(writerList...)
		}
	}
	loggerWriterOnce.Do(func() {
		loggerWriter = &lumberjack.Logger{
			Filename:   filepath.Clean(path),
			MaxBackups: 30,  // files
			MaxSize:    500, // megabytes
			MaxAge:     30,  // days
			Compress:   true,
		}
	})
	writerList = append(writerList, loggerWriter)
	return io.MultiWriter(writerList...)
}
