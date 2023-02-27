package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
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
