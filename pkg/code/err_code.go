package code

import (
	"fmt"
	"net/http"
	"strconv"

	jsoniter "github.com/json-iterator/go"
)

// From parse the ErrorCode from the http.Response
func From(response *http.Response) ErrorCode {
	decoder := jsoniter.ConfigCompatibleWithStandardLibrary.NewDecoder(response.Body)
	decoder.UseNumber()
	var result errCode
	if err := decoder.Decode(&result); err != nil {
		return ErrParseContent.WithResult(err)
	}
	return &result
}

// Froze defines ErrorCode
func Froze(code, message string) ErrorCode {
	return &errCode{
		ErrCode:    code,
		ErrMessage: message,
	}
}

// ErrorCode length
const (
	codeLength       = 10
	statusCodeLength = 3
)

type errCode struct {
	// 3(http)+3(service)+4(error)
	ErrCode    string      `json:"code" binding:"required,len=10"`
	ErrMessage string      `json:"message"`
	ErrResult  interface{} `json:"result"`
}

func (e *errCode) Error() string {
	return fmt.Sprintf("code:%s,message:%s,result:%v",
		e.ErrCode, e.ErrMessage, e.ErrResult)
}

func (e *errCode) MarshalJSON() ([]byte, error) {
	return jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(e)
}

func (e *errCode) UnmarshalJSON(bytes []byte) error {
	return jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(bytes, e)
}

func (e *errCode) StatusCode() int {
	statusCode, _ := strconv.Atoi(e.ErrCode[:statusCodeLength])
	return statusCode
}

func (e *errCode) Code() string {
	return e.ErrCode
}

func (e *errCode) Message() string {
	return e.ErrMessage
}

func (e *errCode) Result() interface{} {
	return e.ErrResult
}

func (e *errCode) WithStatusCode(statusCode int) ErrorCode {
	ec := *e
	ec.ErrCode = fmt.Sprintf("%3d%s", statusCode, ec.ErrCode[statusCodeLength:])
	return &ec
}

func (e *errCode) WithCode(code string) ErrorCode {
	ec := *e
	ec.ErrCode = code
	return &ec
}

func (e *errCode) WithMessage(msg string) ErrorCode {
	ec := *e
	ec.ErrMessage = msg
	return &ec
}

func (e *errCode) WithResult(result interface{}) ErrorCode {
	ec := *e
	ec.ErrResult = result
	return &ec
}

func (e *errCode) Equal(v error) bool {
	for v != nil {
		u, ok := v.((interface {
			Unwrap() error
		}))
		if !ok {
			break
		}
		v = u.Unwrap()
	}
	err, ok := v.(ErrorCode)
	if !ok {
		return false
	}
	return err.Code() == e.Code() && err.Message() == e.Message()
}
