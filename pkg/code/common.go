package code

import (
	"fmt"
)

var (
	// 00~99为服务级别错误码

	ErrInternalServerError = Froze("COMMON.5000000001", "服务器内部错误")
	ErrInvalidParam        = Froze("COMMON.4000000002", "请求参数不正确")
	ErrNotFound            = Froze("COMMON.4040000003", "资源不存在")
	ErrNotAllowMethod      = Froze("COMMON.4050000004", "不允许此方法")
	ErrParseContent        = Froze("COMMON.5000000005", "解析内容失败")
	ErrForbidden           = Froze("COMMON.4030000006", "不允许访问")
	ErrUnauthorized        = Froze("COMMON.4010000007", "用户未认证")
	ErrCodeUnknown         = Froze("COMMON.5000000008", "未知错误")
	ErrDatabaseException   = Froze("COMMON.4000000009", "数据异常")
	ErrNoRowsAffected      = Froze("COMMON.4001000010", "数据无更新")
)

// AddCode business code to codeMessageBox
func AddCode(m map[ErrorCode]struct{}) error {
	temp := make(map[string]struct{})
	for errorCode := range map[ErrorCode]struct{}{
		ErrInternalServerError: {},
		ErrInvalidParam:        {},
		ErrNotFound:            {},
		ErrNotAllowMethod:      {},
		ErrParseContent:        {},
		ErrForbidden:           {},
		ErrUnauthorized:        {},
		ErrCodeUnknown:         {},
	} {
		if err := check(errorCode); err != nil {
			return err
		}
		code := errorCode.Code()
		_, ok := temp[code]
		if ok {
			return fmt.Errorf("error code %s(%s) already exists,result:%v",
				code, errorCode.Message(), errorCode.Result())
		}
		temp[code] = struct{}{}
	}
	for errorCode := range m {
		if err := check(errorCode); err != nil {
			return err
		}
		code := errorCode.Code()
		_, ok := temp[code]
		if ok {
			return fmt.Errorf("error code %s(%s) already exists,result:%v",
				code, errorCode.Message(), errorCode.Result())
		}
		temp[code] = struct{}{}
	}
	return nil
}

// check validate ErrorCode's code must be 3(service)+4(error)
func check(err ErrorCode) error {
	code := err.Code()
	statusCode := err.StatusCode()
	if statusCode < 100 || statusCode >= 600 {
		return fmt.Errorf("error code %s has invalid status code %d", code, statusCode)
	}
	if l := len(code); l != 7 {
		return fmt.Errorf("error code %s is %d,but it must be 7", code, l)
	}
	return nil
}
