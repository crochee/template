package resp

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"go-template/pkg/code"
)

type response struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

// Error gin Response with error
func Error(ctx *gin.Context, err error) {
	for err != nil {
		u, ok := err.((interface {
			Unwrap() error
		}))
		if !ok {
			break
		}
		err = u.Unwrap()
	}
	e, ok := err.(code.ErrorCode)
	if !ok {
		ctx.AbortWithStatusJSON(code.ErrCodeUnknown.StatusCode(), &response{
			Code:    code.ErrCodeUnknown.Code(),
			Message: code.ErrCodeUnknown.Message(),
			Result:  fmt.Sprintf("%v %v", code.ErrCodeUnknown.Result(), err),
		})
		return
	}
	ctx.AbortWithStatusJSON(e.StatusCode(), &response{
		Code:    e.Code(),
		Message: e.Message(),
		Result:  e.Result(),
	})
}

// ErrorParam gin response with invalid parameter tip
func ErrorParam(ctx *gin.Context, err error) {
	Error(ctx,
		code.ErrInvalidParam.WithResult(err))
}
