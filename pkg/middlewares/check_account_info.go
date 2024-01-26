package middlewares

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"template/pkg/code"
	"template/pkg/resp"
	"template/pkg/utils/v"
)

const msgStr = " header is must be required"

func CheckAccountInfo(c *gin.Context) {
	// 云警管理员，直接跳过
	if c.GetHeader(v.HeaderAdminID) != "" {
		c.Next()
		return
	}

	if c.GetHeader(v.HeaderAccountID) == "" {
		resp.Error(c, code.ErrCodeForbidden.WithMessage(fmt.Sprintf("%s %s", v.HeaderAccountID, msgStr)))
		return
	}
	if c.GetHeader(v.HeaderUserID) == "" {
		resp.Error(c, code.ErrCodeForbidden.WithMessage(fmt.Sprintf("%s %s", v.HeaderUserID, msgStr)))
		return
	}
	c.Next()
}
