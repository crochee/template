package middleware

import (
	"github.com/gin-gonic/gin"
)

// RequestLogger 设置请求日志
func RequestLogger(c *gin.Context) {
	//ctx := c.Request.Context()
	//c.Request = c.Request.WithContext(logger.With(ctx, log))
	c.Next()
}
