package resp

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"template/pkg/code"
	"template/pkg/logger"
	"template/pkg/msg"
)

type Response struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

func (r *Response) Convert(statusCode int) code.ErrorCode {
	return code.Froze(fmt.Sprintf("%d-%s", statusCode, r.Code),
		r.Message).WithResult(r.Result)
}

// Error gin Response with error
func Error(c *gin.Context, err error) {
	ctx := c.Request.Context()
	logger.From(ctx).Error("response failed", zap.Error(err))

	span := trace.SpanFromContext(ctx)
	span.AddEvent(semconv.ExceptionEventName, trace.WithAttributes(
		msg.LocateKey.String(msg.CallerFunc(0)),
		semconv.ExceptionMessage(fmt.Sprintf("%+v", err)),
	))
	for err != nil {
		u, ok := err.(interface {
			Unwrap() error
		})
		if !ok {
			break
		}
		err = u.Unwrap()
	}
	if err == nil {
		c.AbortWithStatusJSON(code.ErrInternalServerError.StatusCode(), code.ErrInternalServerError)
		return
	}
	sc, scok := err.(interface {
		StatusCode() int
		Code() string
		Message() string
		Result() interface{}
	})
	if !scok {
		c.AbortWithStatusJSON(code.ErrCodeUnknown.StatusCode(), &Response{
			Code:    fmt.Sprintf("%s.%3d%s", code.ErrCodeUnknown.ServiceName(), code.ErrCodeUnknown.StatusCode(), code.ErrCodeUnknown.Code()),
			Message: code.ErrCodeUnknown.Message(),
			Result:  fmt.Sprintf("%v %e", code.ErrCodeUnknown.Result(), err),
		})
		return
	}
	jm, jmok := err.(json.Marshaler)
	if jmok {
		c.AbortWithStatusJSON(sc.StatusCode(), jm)
		return
	}
	sn, snok := err.(interface {
		ServiceName() string
	})
	if !snok {
		c.AbortWithStatusJSON(sc.StatusCode(), &Response{
			Code:    sc.Code(),
			Message: sc.Message(),
			Result:  sc.Result(),
		})
		return
	}
	c.AbortWithStatusJSON(sc.StatusCode(), &Response{
		Code:    fmt.Sprintf("%s.%3d%s", sn.ServiceName(), sc.StatusCode(), sc.Code()),
		Message: sc.Message(),
		Result:  sc.Result(),
	})
}

// ErrorParam gin response with invalid parameter tip
func ErrorParam(c *gin.Context, err error) {
	Error(c,
		code.ErrInvalidParam.WithResult(err.Error()))
}
