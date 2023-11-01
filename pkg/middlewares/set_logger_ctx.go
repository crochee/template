package middlewares

import (
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"template/pkg/logger"
	"template/pkg/utils/v"
)

// SetRequestLogger 设置请求日志
func SetZeroLogger(c *gin.Context) {
	// 请求头X-Trace-ID不能为空，为空时需要自动生成
	traceID := c.GetHeader(v.HeaderTraceID)
	if traceID == "" {
		traceID = "req-" + uuid.NewV4().String()
		c.Request.Header.Set(v.HeaderTraceID, traceID) // 请求头
	}
	// 设置响应头
	if c.Writer.Header().Get(v.HeaderTraceID) == "" {
		c.Writer.Header().Set(v.HeaderTraceID, traceID) // 响应头
	}
	logContext := logger.FromContext(c.Request.Context()).With().Str("trace_id", traceID)

	if adminID := c.GetHeader(v.HeaderAdminID); adminID != "" {
		logContext = logContext.Str("admin_id", adminID)
	}
	if accountID := c.GetHeader(v.HeaderAccountID); accountID != "" {
		logContext = logContext.Str("account_id", accountID)
	}
	if userID := c.GetHeader(v.HeaderUserID); userID != "" {
		logContext = logContext.Str("user_id", userID)
	}
	c.Request = c.Request.WithContext(logger.WithContext(c.Request.Context(), &logger.Logger{
		Logger: logContext.Str("client_ip", c.ClientIP()).Logger()},
	))
	c.Next()
}

// SetZapLogger 设置请求日志
func SetZapLogger(c *gin.Context) {
	var fields []zapcore.Field
	if traceID := c.GetHeader(v.HeaderTraceID); traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}
	if accountID := c.GetHeader(v.HeaderAccountID); accountID != "" {
		fields = append(fields, zap.String("account_id", accountID))
	}
	if userID := c.GetHeader(v.HeaderUserID); userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}
	if source := c.GetHeader(v.HeaderSource); source != "" {
		fields = append(fields, zap.String("source", source))
	}
	if len(fields) > 0 {
		l := logger.From(c.Request.Context()).With(fields...)
		c.Request = c.Request.WithContext(logger.With(c.Request.Context(), l))
	}
	c.Next()
}
