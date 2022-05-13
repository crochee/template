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
