package code

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"go_template/pkg/json"
)

// From parse the ErrorCode from the http.Response
func From(response *http.Response) ErrorCode {
	var result struct {
		Code    string      `json:"code"`
		Message string      `json:"message"`
		Result  interface{} `json:"result"`
	}
	if err := json.DecodeUseNumber(response.Body, &result); err != nil {
		return ErrParseContent.WithResult(err.Error())
	}
	return Froze(result.Code, result.Message).WithResult(result.Result)
}

// Froze defines ErrorCode
func Froze(code, message string) ErrorCode {
	// 默认 ErrInternalServerError
	e := &errCode{
		serviceName:    "undefined",
		httpStatusCode: http.StatusInternalServerError,
		code:           "0000001",
		message:        message,
		result:         nil,
	}
	multiErrCode := code
	index := strings.Index(multiErrCode, ".")
	if index > 0 {
		e.serviceName = multiErrCode[:index]
		if index >= len(multiErrCode)-1 {
			return e.WithResult(code + ";" + message)
		}
		multiErrCode = multiErrCode[index+1:]
	}
	if len(multiErrCode) <= 3 {
		return e.WithResult(code + ";" + message)
	}
	httpStatusCode, err := strconv.Atoi(multiErrCode[:3])
	if err != nil {
		return e.WithResult(
			fmt.Sprintf("code:%s,message:%s;%e",
				code, message, err))
	}
	if httpStatusCode < 100 || httpStatusCode > 599 {
		return e.WithResult(code + ";" + message)
	}
	e.httpStatusCode = httpStatusCode
	e.code = multiErrCode[3:]
	return e
}

type errCode struct {
	serviceName    string
	httpStatusCode int
	// 3(service)+4(error)
	code    string
	message string
	result  interface{}
}

func (e *errCode) Error() string {
	return fmt.Sprintf("service_name:%s,http_status_code:%d,code:%s,message:%s,result:%v",
		e.serviceName, e.httpStatusCode, e.code, e.message, e.result)
}

func (e *errCode) ServiceName() string {
	return e.serviceName
}

func (e *errCode) StatusCode() int {
	return e.httpStatusCode
}

func (e *errCode) Code() string {
	return e.code
}

func (e *errCode) Message() string {
	return e.message
}

func (e *errCode) Result() interface{} {
	return e.result
}

func (e *errCode) WithStatusCode(statusCode int) ErrorCode {
	ec := *e
	ec.httpStatusCode = statusCode
	return &ec
}

func (e *errCode) WithCode(code string) ErrorCode {
	ec := *e
	ec.code = code
	return &ec
}

func (e *errCode) WithMessage(msg string) ErrorCode {
	ec := *e
	ec.message = msg
	return &ec
}

func (e *errCode) WithResult(result interface{}) ErrorCode {
	ec := *e
	ec.result = result
	return &ec
}

func (e *errCode) Is(v error) bool {
	err, ok := v.(ErrorCode)
	if !ok {
		return false
	}
	return err.Code() == e.Code()
}
