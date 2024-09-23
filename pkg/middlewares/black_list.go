package middlewares

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"template/internal/util/v"
	"template/pkg/code"
	"template/pkg/resp"
)

func BlackList(blackList map[string]struct{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 不存在HeaderAccountID直接放行
		source := c.GetHeader(v.HeaderAccountID)
		if source == "" {
			c.Next()
			return
		}

		if _, exists := blackList[source]; exists {
			resp.Error(
				c,
				code.ErrForbidden.WithResult(fmt.Sprintf("source %s is in black list", source)),
			)
			return
		}
		c.Next()

	}
}
