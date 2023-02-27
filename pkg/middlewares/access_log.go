package middlewares

import (
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"template/pkg/logger"
	"template/pkg/utils"
)

// Log request logx
func Log(c *gin.Context) {
	// Start timer
	start := time.Now()
	path := c.Request.URL.Path
	raw := c.Request.URL.RawQuery
	httpRequest, err := httputil.DumpRequest(c.Request, true)
	if err != nil {
		logger.From(c.Request.Context()).Error("fail to DumpRequest", zap.Error(err))
	}

	// Process request
	c.Next()
	// Log only when path is not being skipped

	param := gin.LogFormatterParams{
		Request: c.Request,
		Keys:    c.Keys,
	}
	// Stop timer
	param.TimeStamp = time.Now()
	param.Latency = param.TimeStamp.Sub(start)

	param.ClientIP = c.ClientIP()
	param.Method = c.Request.Method
	param.StatusCode = c.Writer.Status()
	param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()
	param.BodySize = c.Writer.Size()

	if raw != "" {
		var buf strings.Builder
		buf.WriteString(path)
		buf.WriteByte('?')
		buf.WriteString(raw)
		path = buf.String()
	}
	param.Path = path
	logger.From(c.Request.Context()).Info(defaultLogFormatter(&param), zap.String("Request", utils.String(httpRequest)))
}

// defaultLogFormatter is the default log format function Logger middlewares uses.
func defaultLogFormatter(param *gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}
	if param.Latency > time.Minute {
		// Truncate in a golang < 1.8 safe way
		param.Latency -= param.Latency % time.Second
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
