package stack

import (
	"bytes"
	"context"
	"runtime"

	"github.com/gin-gonic/gin"

	"template/pkg/code"
	"template/pkg/logger"
	"template/pkg/resp"
)

func panicInfo() string {
	s := []byte("/src/runtime/panic.go")
	e := []byte("\ngoroutine ")

	line := []byte("\n")
	stack := make([]byte, 1<<16)
	length := runtime.Stack(stack, true)
	start := bytes.Index(stack, s)
	stack = stack[start:length]
	start = bytes.Index(stack, line) + 1
	stack = stack[start:]

	end := bytes.LastIndex(stack, line)
	if end != -1 {
		stack = stack[:end]
	}

	end = bytes.Index(stack, e)
	if end != -1 {
		stack = stack[:end]
	}

	stack = bytes.TrimRight(stack, "\n")
	return string(stack)
}

func RecoverPanic(ctx context.Context, c *gin.Context) {
	if err := recover(); err != nil {
		log := logger.FromContext(ctx)
		log.Error().Interface("ERR", err).Msg("recover panic from goroutine")
		log.Error().Str("ERR", "").Msg(panicInfo())

		if c != nil {
			resp.Error(c, code.ErrInternalServerError)
		}
	}
}
