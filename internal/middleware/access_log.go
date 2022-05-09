package middleware

import (
	"strconv"
	"strings"
	"time"

	"github.com/crochee/lirity/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Log request logx
func Log(ctx *gin.Context) {
	// Start timer
	start := time.Now()
	path := ctx.Request.URL.Path
	raw := ctx.Request.URL.RawQuery
	// Process request
	ctx.Next()
	// Log only when path is not being skipped

	param := gin.LogFormatterParams{
		Request: ctx.Request,
		Keys:    ctx.Keys,
	}
	// Stop timer
	param.TimeStamp = time.Now()
	param.Latency = param.TimeStamp.Sub(start)

	param.ClientIP = ctx.ClientIP()
	param.Method = ctx.Request.Method
	param.StatusCode = ctx.Writer.Status()
	param.ErrorMessage = ctx.Errors.ByType(gin.ErrorTypePrivate).String()

	param.BodySize = ctx.Writer.Size()
	param.Keys = ctx.Keys
	if raw != "" {
		var buf strings.Builder
		buf.WriteString(path)
		buf.WriteByte('?')
		buf.WriteString(raw)
		path = buf.String()
	}
	param.Path = path
	logger.From(ctx.Request.Context()).Info(defaultLogFormatter(&param), zap.Any("keys", param.Keys))
}

// defaultLogFormatter is the default log format function Logger middleware uses.
func defaultLogFormatter(param *gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}
	if param.Latency > time.Minute {
		// Truncate in a golang < 1.8 safe way
		param.Latency = param.Latency - param.Latency%time.Second
	}
	var buf strings.Builder
	buf.WriteString(statusColor)
	buf.WriteByte(' ')
	buf.WriteString(strconv.Itoa(param.StatusCode))
	buf.WriteByte(' ')
	buf.WriteString(resetColor)
	buf.WriteString("| ")
	buf.WriteString(param.Latency.String())
	buf.WriteString(" | ")
	buf.WriteString(param.ClientIP)
	buf.WriteString(" |")
	buf.WriteString(methodColor)
	buf.WriteByte(' ')
	buf.WriteString(param.Method)
	buf.WriteByte(' ')
	buf.WriteString(resetColor)
	buf.WriteByte('|')
	buf.WriteString(strconv.Itoa(param.BodySize))
	buf.WriteString("| ")
	buf.WriteString(param.Path)
	if param.ErrorMessage != "" {
		buf.WriteString(" | ")
		buf.WriteString(param.ErrorMessage)
	}
	return buf.String()
}
