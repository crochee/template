package code

import "encoding/json"

type ErrorCode interface {
	error
	json.Marshaler
	json.Unmarshaler
	StatusCode() int
	Code() string
	Message() string
	Result() interface{}
	WithStatusCode(int) ErrorCode
	WithCode(string) ErrorCode
	WithMessage(string) ErrorCode
	WithResult(interface{}) ErrorCode
	Equal(v error) bool
}
