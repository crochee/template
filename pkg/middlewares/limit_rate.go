package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/mailgun/ttlmap"

	"template/internal/util/v"
	"template/pkg/code"
	"template/pkg/limit"
	"template/pkg/resp"
)

// RateLimit limit request rate
func RateLimit(
	rlf func(key string) limit.RateLimiter,
) func(c *gin.Context) {
	buckets, err := ttlmap.NewConcurrent(65536)
	if err != nil {
		panic(err)
	}
	return func(c *gin.Context) {
		// 不存在HeaderAccountID直接放行
		source := c.GetHeader(v.HeaderAccountID)
		if source == "" {
			c.Next()
			return
		}
		var bucket limit.RateLimiter
		if rlSource, exists := buckets.Get(source); exists {
			bucket = rlSource.(limit.RateLimiter)
		} else {
			bucket = rlf(source)
		}
		// We Set even in the case where the source already exists,
		// because we want to update the expiryTime everytime we get the source,
		// as the expiryTime is supposed to reflect the activity (or lack thereof) on that source.
		if err := buckets.Set(source, bucket, 60); err != nil {
			resp.Error(c, code.ErrInternalServerError.WithResult("Could not insert/update bucket"))
			return
		}
		err := bucket.Wait(c.Request.Context())
		if err != nil {
			resp.Error(c, code.ErrTooManyRequests.WithResult(err.Error()))
			return
		}
		c.Next()
	}
}
