package code

import (
	"fmt"
)

var (
	// 00~99为服务级别错误码

	ErrInternalServerError  = Froze("5000000001", "服务器内部错误")
	ErrInvalidParam         = Froze("4000000002", "请求参数不正确")
	ErrNotFound             = Froze("4040000003", "资源不存在")
	ErrNotAllowMethod       = Froze("4050000004", "不允许此方法")
	ErrParseContent         = Froze("5000000005", "解析内容失败")
	ErrForbidden            = Froze("4030000006", "不允许访问")
	ErrUnauthorized         = Froze("4010000007", "用户未认证")
	ErrCodeUnknown          = Froze("5000000008", "未知错误")
	ErrCodeRedisCacheOption = Froze("5000000009", "Redis缓存操作失败")

	// woslo 错误
	ErrCodeInvalidParam        = Froze("400-1000000", "请求参数不正确")
	ErrCodeNotFound            = Froze("404-1000001", "资源不存在")
	ErrCodeInternalServerError = Froze("500-1000002", "服务器内部错误")
	ErrCodeInvalidStatus       = Froze("500-1000003", "资源状态不满足操作要求")
	ErrCodeEUnknown            = Froze("500-1000004", "未知错误")
	ErrCodeForbidden           = Froze("403-1000005", "禁止请求")
	ErrCodeInvalidBody         = Froze("500-1000006", "响应体无法被正常解析")
	ErrCodeERedisCacheOption   = Froze("500-1100129", "Redis缓存操作失败")
)

// AddCode business code to codeMessageBox
func AddCode(m map[ErrorCode]struct{}) error {
	temp := make(map[string]struct{})
	for errorCode := range map[ErrorCode]struct{}{
		ErrInternalServerError:  {},
		ErrInvalidParam:         {},
		ErrNotFound:             {},
		ErrNotAllowMethod:       {},
		ErrParseContent:         {},
		ErrForbidden:            {},
		ErrUnauthorized:         {},
		ErrCodeUnknown:          {},
		ErrCodeRedisCacheOption: {},

		ErrCodeInvalidParam:        {},
		ErrCodeNotFound:            {},
		ErrCodeInternalServerError: {},
		ErrCodeInvalidStatus:       {},
		ErrCodeEUnknown:            {},
		ErrCodeForbidden:           {},
		ErrCodeInvalidBody:         {},
		ErrCodeERedisCacheOption:   {},
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
	return nil
}

func MapErr(err error, opts ...func(error) error) error {
	for _, f := range opts {
		err = f(err)
	}
	return err
}

func OkOrDefault(err error) error {
	if v, ok := err.(ErrorCode); ok {
		return v
	}
	return ErrCodeInternalServerError.WithResult(err.Error())
}
