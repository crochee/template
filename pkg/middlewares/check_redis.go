package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"template/pkg/code"
	"template/pkg/env"
	"template/pkg/redis"
	"template/pkg/resp"
)

// 根据Redis服务状态决定是否拦截非GET请求
// 由于用户配额服务强依赖Redis服务，所以当Redis服务故障时，只放行GET请求
func CheckRedis(c *gin.Context) {
	// 私有云直接跳过
	if env.IsPrivate() {
		c.Next()
		return
	}

	// 未开启配额功能，直接跳过
	if !viper.GetBool("user_quota.enable") {
		c.Next()
		return
	}

	// 如果Redis服务不可用且请求为非GET请求，拦截请求并返回错误信息
	if !redis.IsRedisActive() && c.Request.Method != http.MethodGet {
		resp.Error(c, code.ErrForbidden.WithMessage("服务维护中，请稍后再试..."))
		return
	}

	// before request
	c.Next()
	// after request
}
