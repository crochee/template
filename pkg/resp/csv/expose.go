package csv

import (
	"errors"
	"reflect"
)

type expose struct {
}

// GetStruct means if it has a fieldName value, it returns this value, otherwise it returns itself
func (e expose) GetStruct(data interface{}, fieldNames ...string) (interface{}, error) {
	var err error
	for _, field := range fieldNames {
		if data, err = e.getStructWithName(data, field); err != nil {
			return nil, err
		}
	}
	return data, nil
}

func (expose) getStructWithName(data interface{}, fieldName string) (interface{}, error) {
	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, errors.New("not struct")
	}
	result := v.FieldByName(fieldName)
	if !result.IsValid() {
		return data, nil
	}
	if result.IsNil() {
		return nil, errors.New("it's nil")
	}
	if !result.CanInterface() {
		return nil, errors.New("can't interface")
	}
	return result.Interface(), nil
}
