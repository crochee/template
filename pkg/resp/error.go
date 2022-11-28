package resp

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/crochee/devt/pkg/code"
	"github.com/crochee/devt/pkg/logger"
)

type response struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

// Error gin Response with error
func Error(c *gin.Context, err error) {
	logger.From(c.Request.Context()).Error("response failed", zap.Error(err))
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
			Code:    fmt.Sprintf("%s.%3d%s", code.ErrCodeUnknown.ServiceName(), code.ErrCodeUnknown.StatusCode(), code.ErrCodeUnknown.Code()),
			Message: code.ErrCodeUnknown.Message(),
			Result:  fmt.Sprintf("%v %e", code.ErrCodeUnknown.Result(), err),
		})
		return
	}
	c.AbortWithStatusJSON(e.StatusCode(), &response{
		Code:    fmt.Sprintf("%s.%3d%s", e.ServiceName(), e.StatusCode(), e.Code()),
		Message: e.Message(),
		Result:  e.Result(),
	})
}

// ErrorParam gin response with invalid parameter tip
func ErrorParam(c *gin.Context, err error) {
	Error(c,
		code.ErrInvalidParam.WithResult(err.Error()))
}
