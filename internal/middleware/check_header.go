package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"template/pkg/code"
	"template/pkg/resp"
	"template/pkg/utils/v"
)

// CheckHeaders 检查请求头传参正确性
func CheckHeaders(c *gin.Context) {
	xAccountID := c.GetHeader(v.HeaderAccountID)
	if xAccountID == "" {
		resp.Error(c, code.ErrForbidden.WithMessage(fmt.Sprintf("%s header is must be required", v.HeaderAccountID)))
		return
	}
	xUserID := c.GetHeader(v.HeaderUserID)
	if xUserID == "" {
		resp.Error(c, code.ErrForbidden.WithMessage(fmt.Sprintf("%s header is must be required", v.HeaderUserID)))
		return
	}
	// 有token的情况需要进行网关鉴权后账号和用户的验证
	if c.GetHeader(v.HeaderGWToken) != "" {
		gwAccountID := c.GetHeader(v.HeaderGWAccountID)
		if xAccountID != gwAccountID {
			resp.Error(c, code.ErrForbidden.
				WithMessage("非法账号，禁止访问").
				WithResult(fmt.Sprintf("account id expect %s,but got %s", gwAccountID, xAccountID)))
			return
		}
		gwUserID := c.GetHeader(v.HeaderGWUserID)
		if xUserID != gwUserID {
			resp.Error(c, code.ErrForbidden.
				WithMessage("非法用户，禁止访问").
				WithResult(fmt.Sprintf("user id expect %s,but got %s", gwUserID, xUserID)))
			return
		}
	}
	c.Next()
}
