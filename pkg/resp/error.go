package resp

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"go_template/pkg/code"
)

type response struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

// Error gin Response with error
func Error(c *gin.Context, err error) {
	for err != nil {
		u, ok := err.(interface {
			Unwrap() error
		})
		if !ok {
			break
		}
		err = u.Unwrap()
	}
	e, ok := err.(code.ErrorCode)
	if !ok {
		c.AbortWithStatusJSON(code.ErrCodeUnknown.StatusCode(), &response{
			Code:    fmt.Sprintf("%3d%s", code.ErrCodeUnknown.StatusCode(), code.ErrCodeUnknown.Code()),
			Message: code.ErrCodeUnknown.Message(),
			Result:  fmt.Sprintf("%v %v", code.ErrCodeUnknown.Result(), err),
		})
		return
	}
	c.AbortWithStatusJSON(e.StatusCode(), &response{
		Code:    fmt.Sprintf("%3d%s", e.StatusCode(), e.Code()),
		Message: e.Message(),
		Result:  e.Result(),
	})
}

// ErrorParam gin response with invalid parameter tip
func ErrorParam(ctx *gin.Context, err error) {
	Error(ctx,
		code.ErrInvalidParam.WithResult(err.Error()))
}
