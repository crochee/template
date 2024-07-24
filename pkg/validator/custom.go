package validator

import (
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	"template/pkg/env"
	"template/pkg/set"
)

var orderCompile = regexp.MustCompile(`^[a-z][a-z_]{0,30}[a-z](\s(asc|ASC|desc|DESC))?(,[a-z][a-z_]{0,30}[a-z](\s(asc|ASC|desc|DESC))?)*$`)

func OrderWithDBSort(f1 validator.FieldLevel) bool {
	valid, ok := f1.Field().Interface().(string)
	if !ok {
		return false
	}
	return orderCompile.MatchString(valid)
}

var nameCompile = regexp.MustCompile("^[0-9a-zA-Z\u4e00-\u9fa5_]{1,32}$")

// ComponentName 只允许输入数字、字母、汉字、下划线，字符长度1-32个字符
func ComponentName(f1 validator.FieldLevel) bool {
	valid, ok := f1.Field().Interface().(string)
	if !ok {
		return false
	}
	return nameCompile.MatchString(valid)
}

func AccountIDsValidate(fl validator.FieldLevel) bool {
	// 非私有云环境下不验证此字段
	if !env.IsPrivate() {
		return true
	}

	value := fl.Field().Interface()
	if value == nil {
		return false
	}
	data, ok := value.([]string)
	if ok {
		if len(data) == 0 {
			return false
		}
	}
	return true
}

// DatetimeValidate datetime tag参数校验函数，需要注册提前注册到main.go文件中
func DatetimeValidate(fl validator.FieldLevel) bool {
	datetime, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	_, err := time.Parse(fl.Param(), datetime)
	if err != nil {
		return false
	}

	return true
}

// CommaListValidate comma_list tag参数校验函数，需要注册提前注册到main.go文件中
func CommaListValidate(fl validator.FieldLevel) bool {
	list, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	targetValues := strings.Split(fl.Param(), " ")
	for _, value := range strings.Split(list, ",") {
		if !set.IsContains(value, targetValues) {
			return false
		}
	}

	return true
}
