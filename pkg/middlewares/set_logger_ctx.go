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
		traceID = uuid.NewV4().String()
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
	// NOTE:为避免openapi用户拿别人的aksk调用，此处需要将accountid,userid重写入header
	if accountID := c.GetHeader("Accountid"); accountID != "" {
		c.Request.Header.Set(v.HeaderAccountID, accountID)
	}
	if userID := c.GetHeader("Userid"); userID != "" {
		c.Request.Header.Set(v.HeaderUserID, userID)
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
	traceID := c.GetHeader(v.HeaderTraceID)
	if traceID == "" {
		traceID = uuid.NewV4().String()
		c.Request.Header.Set(v.HeaderTraceID, traceID) // 请求头
	}
	// 设置响应头
	if c.Writer.Header().Get(v.HeaderTraceID) == "" {
		c.Writer.Header().Set(v.HeaderTraceID, traceID) // 响应头
	}
	var fields = []zapcore.Field{zap.String("trace_id", traceID)}
	// NOTE:为避免openapi用户拿别人的aksk调用，此处需要将accountid,userid重写入header
	if accountID := c.GetHeader("Accountid"); accountID != "" {
		c.Request.Header.Set(v.HeaderAccountID, accountID)
	}
	if userID := c.GetHeader("Userid"); userID != "" {
		c.Request.Header.Set(v.HeaderUserID, userID)
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
	l := logger.From(c.Request.Context()).With(fields...)
	c.Request = c.Request.WithContext(logger.With(c.Request.Context(), l))
	c.Next()
}
