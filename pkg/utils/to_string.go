package utils

import (
	"fmt"
	"reflect"
	"strconv"

	"template/pkg/json"
	"template/pkg/utils/v"
)

func ToUint64(data string) (uint64, error) {
	if data == "" {
		return 0, nil
	}
	return strconv.ParseUint(data, v.Decimal, 64)
}

func ToString(param interface{}) string {
	value := reflect.ValueOf(param)
	kind := value.Kind()
	if kind == reflect.Ptr {
		return ToString(value.Elem().Interface())
	}
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(value.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(value.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(value.Float(), 'g', -1, 64)
	case reflect.String:
		return value.String()
	case reflect.Bool:
		return fmt.Sprintf("%t", value.Bool())
	default:
		data, _ := json.Marshal(value.Interface())
		return String(data)
	}
}
