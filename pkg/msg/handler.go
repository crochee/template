// Package msg
package msg

import (
	"bytes"
	"runtime"
	"strconv"
	"time"
)

const skipOffset = 2 // skip getCallerFrame and Callers

type CallerHandler func(int) string

func CallerFunc(skip int) string {
	_, file, line, ok := runtime.Caller(skip + skipOffset)
	if !ok {
		return "???"
	}
	var buf bytes.Buffer
	buf.WriteString(file)
	buf.WriteByte(':')
	buf.WriteString(strconv.Itoa(line))
	return buf.String()
}

type ServiceNameHandler func() string

func ServiceName() string {
	return "DCSName"
}

type NowHandler func() time.Time
