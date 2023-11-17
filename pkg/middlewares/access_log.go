package middlewares

import (
	"context"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"template/pkg/logger/gormx"
	"template/pkg/utils"
)

// AccessLog request logx
func AccessLog(from func(context.Context) gormx.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		httpRequest, err := httputil.DumpRequest(c.Request, true)
		if err != nil {
			from(c.Request.Context()).Errorf("fail to DumpRequest %+v", err)
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

		code := param.StatusCode
		switch {
		case code >= http.StatusOK && code < http.StatusMultipleChoices:
			from(c.Request.Context()).Infof("%s,Request=%s", defaultLogFormatter(&param), utils.String(httpRequest))
		case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
			from(c.Request.Context()).Warnf("%s,Request=%s", defaultLogFormatter(&param), utils.String(httpRequest))
		case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
			from(c.Request.Context()).Errorf("%s,Request=%s", defaultLogFormatter(&param), utils.String(httpRequest))
		default:
			from(c.Request.Context()).Errorf("%s,Request=%s", defaultLogFormatter(&param), utils.String(httpRequest))
		}
	}
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
