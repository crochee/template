package utils

import (
	"net/url"

	"github.com/gin-gonic/gin"
)

// GetHeader return the decoded header value from gin.Context
func GetHeader(c *gin.Context, key string) string {
	value := c.GetHeader(key)
	decoded, err := url.QueryUnescape(value)
	if err != nil {
		decoded = value
	}
	return decoded
}
