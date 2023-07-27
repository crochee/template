package logger

import (
	"io"
	"path/filepath"
	"sync"

	"github.com/mattn/go-colorable"
	"gopkg.in/natefinch/lumberjack.v2"
)

// 日志滚动转存默认参数
const (
	LoggerMaxSize    = 500 // 日志默认最大size，单位MB
	LoggerMaxAge     = 30  // 日志默认最大保留时间，单位day
	LoggerMaxBackups = 10  // 日志最大备份数量，单位个
)

type writerOption struct {
	maxSize    int
	maxAge     int
	maxBackups int
	localTime  bool
	compress   bool
}

type WriterOption func(*writerOption)

func WithMaxSize(size int) WriterOption {
	return func(o *writerOption) {
		o.maxSize = size
	}
}

func WithMaxAge(age int) WriterOption {
	return func(o *writerOption) {
		o.maxAge = age
	}
}

func WithMaxBackups(backups int) WriterOption {
	return func(o *writerOption) {
		o.maxBackups = backups
	}
}

func WithLocalTime(localTime bool) WriterOption {
	return func(o *writerOption) {
		o.localTime = localTime
	}
}

func WithCompress(compress bool) WriterOption {
	return func(o *writerOption) {
		o.compress = compress
	}
}

var loggerWriterManagers = NewWriterManager()

// SetWriter return a io.Writer
func SetWriter(console bool, path string, opts ...WriterOption) io.Writer {
	o := &writerOption{
		maxSize:    LoggerMaxSize,
		maxAge:     LoggerMaxAge,
		maxBackups: LoggerMaxBackups,
		localTime:  false,
		compress:   true,
	}
	for _, opt := range opts {
		opt(o)
	}

	var writers []io.Writer
	path = filepath.Clean(path)

	if console {
		writers = append(writers, colorable.NewColorableStdout())
	}
	if path != "" {
		loggerWriter := loggerWriterManagers.GetLogger(path, o.maxSize, o.maxAge, o.maxBackups, o.localTime, o.compress)
		writers = append(writers, loggerWriter)
	}
	return io.MultiWriter(writers...)
}

type writerManager struct {
	mu      sync.RWMutex
	loggers map[string]*lumberjack.Logger
}

func NewWriterManager() *writerManager {
	return &writerManager{
		loggers: make(map[string]*lumberjack.Logger),
	}
}

func (w *writerManager) GetLogger(path string, maxSize, maxAge, maxBackups int,
	localTime, compress bool) io.WriteCloser {
	w.mu.RLock()
	writer, ok := w.loggers[path]
	w.mu.RUnlock()
	if ok {
		return writer
	}
	w.mu.Lock()
	writer = &lumberjack.Logger{
		Filename:   path,
		MaxSize:    maxSize,
		MaxAge:     maxAge,
		MaxBackups: maxBackups,
		LocalTime:  localTime,
		Compress:   compress,
	}
	w.loggers[path] = writer
	w.mu.Unlock()
	return writer
}
