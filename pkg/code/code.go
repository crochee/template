package code

type ErrorCode interface {
	error
	ServiceName() string
	StatusCode() int
	Code() string
	Message() string
	Result() interface{}
	WithStatusCode(int) ErrorCode
	WithCode(string) ErrorCode
	WithMessage(string) ErrorCode
	WithResult(interface{}) ErrorCode
	Is(error) bool
}
