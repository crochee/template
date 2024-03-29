package middlewares

import (
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"

	"template/pkg/utils/v"
)

// SetHeaderContext set some public parameters to header and context
func SetHeaderContext(setCtx func(*gin.Context)) func(c *gin.Context) {
	return func(c *gin.Context) {
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
		setCtx(c)
		c.Next()
	}
}
