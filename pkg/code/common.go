package code

import "fmt"

var (
	// 00~99为服务级别错误码

	ErrInternalServerError = Froze("5000000000", "服务器内部错误")
	ErrInvalidParam        = Froze("4000000001", "请求参数不正确")
	ErrNotFound            = Froze("4040000002", "资源不存在")
	ErrNotAllowMethod      = Froze("4050000003", "不允许此方法")
	ErrParseContent        = Froze("5000000004", "解析内容失败")
	ErrCodeUnknown         = Froze("5000000005", "未知错误")
)

// AddCode business code to codeMessageBox
func AddCode(m map[ErrorCode]struct{}) error {
	temp := make(map[string]string)
	for errorCode := range map[ErrorCode]struct{}{
		ErrInternalServerError: {},
		ErrInvalidParam:        {},
		ErrNotFound:            {},
		ErrNotAllowMethod:      {},
		ErrParseContent:        {},
		ErrCodeUnknown:         {},
	} {
		if err := check(errorCode); err != nil {
			return err
		}
		code := errorCode.Code()
		value, ok := temp[code]
		if ok {
			return fmt.Errorf("error code %s(%s) already exists", code, value)
		}
		temp[code] = errorCode.Message()
	}
	for errorCode := range m {
		if err := check(errorCode); err != nil {
			return err
		}
		code := errorCode.Code()
		value, ok := temp[code]
		if ok {
			return fmt.Errorf("error code %s(%s) already exists", code, value)
		}
		temp[code] = errorCode.Message()
	}
	return nil
}

// check validate ErrorCode's code must be 3(http)+3(service)+4(error)
func check(err ErrorCode) error {
	code := err.Code()
	statusCode := err.StatusCode()
	if statusCode < 100 || statusCode >= 600 {
		return fmt.Errorf("error code %s has invalid status code %d", code, statusCode)
	}
	if l := len(code); l != codeLength {
		return fmt.Errorf("error code %s is %d,but it must be 10", code, l)
	}
	return nil
}
