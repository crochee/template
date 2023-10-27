package ptf

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"go.uber.org/multierr"

	"template/pkg/json"
	"template/pkg/timex"
)

const tagString = "string"

type mapIndexValue struct {
	data map[string]interface{}
}

type parse struct {
	tagName string
}

func (p *parse) parseStruct(obj interface{}) ([]*mapIndexValue, error) {
	if obj == nil {
		return []*mapIndexValue{}, nil
	}
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, errors.New("not struct")
	}
	result := &mapIndexValue{
		data: map[string]interface{}{},
	}

	for i := 0; i < t.NumField(); i++ {
		fv := v.Field(i)
		ft := t.Field(i)

		if !fv.IsValid() {
			continue
		}
		if !fv.CanInterface() {
			continue
		}
		if ft.PkgPath != "" { // unexported
			continue
		}
		tag, found := ft.Tag.Lookup(p.tagName)
		if !found {
			continue
		}
		tags := strings.Split(tag, ",")
		var (
			name   string
			option string
		)
		if len(tags) == 1 {
			name = tags[0]
		} else if len(tags) >= 2 {
			name = tags[0]
			option = tags[1]
		}
		if name == "-" {
			continue // ignore "-"
		}
		if name == "" {
			name = ft.Name // use field name
		}

		// map to dynamic tile
		if ft.Type.Kind() == reflect.Map && option == "dynamic_tile" {
			iter := fv.MapRange()
			for iter.Next() {
				// key = name;index
				key := iter.Key().String()
				value := iter.Value().Interface()
				result.data[key] = value
			}
			continue
		}
		if ft.Type.String() == "time.Time" && option == "time" {
			result.data[name] = timex.TimeFormat(fv.Interface().(time.Time))
			continue
		}
		if ft.Anonymous || fv.Kind() == reflect.Slice || fv.Kind() == reflect.Array ||
			fv.Kind() == reflect.Struct || fv.Kind() == reflect.Ptr {
			if fv.IsZero() {
				continue
			}
			embedded, err := p.parse(fv.Interface())
			if err != nil {
				return nil, err
			}
			if (fv.Kind() == reflect.Slice || fv.Kind() == reflect.Array) && option == "fmt" {
				// fmt
				embedded = format(name, embedded)
			}

			for _, embMap := range embedded {
				for embName, embValue := range embMap.data {
					result.data[embName] = embValue
				}
			}
			continue
		}
		if option == tagString {
			var tempString interface{}
			if fv.Kind() == reflect.Uint64 && fv.IsZero() {
				tempString = ""
			} else {
				tempString = value2String(fv)
			}
			if tempString != nil {
				result.data[name] = tempString
				continue
			}
		}
		result.data[name] = fv.Interface()
	}
	return []*mapIndexValue{result}, nil
}

func (p *parse) parse(obj interface{}) ([]*mapIndexValue, error) {
	if obj == nil {
		return []*mapIndexValue{}, nil
	}
	value := reflect.ValueOf(obj)
	switch value.Kind() {
	case reflect.Ptr:
		value = value.Elem()
		if !value.IsValid() {
			return []*mapIndexValue{}, nil
		}
		return p.parse(value.Interface())
	case reflect.Struct:
		return p.parseStruct(obj)
	case reflect.Slice, reflect.Array:
		count := value.Len()
		var errs error
		tempMap := make([]*mapIndexValue, 0, count)
		for i := 0; i < count; i++ {
			if !value.Index(i).CanInterface() {
				errs = multierr.Append(errs, fmt.Errorf("%s can't interface", value.Index(i).String()))
				continue
			}
			if v, err := p.parse(value.Index(i).Interface()); err != nil {
				errs = multierr.Append(errs, err)
			} else {
				tempMap = append(tempMap, v...)
			}
		}
		return tempMap, errs
	default:
		return nil, fmt.Errorf("not support %s", value.Kind().String())
	}
}

func value2String(fv reflect.Value) interface{} {
	kind := fv.Kind()
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(fv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(fv.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(fv.Float(), 'f', 2, 64)
	default:
		data, _ := json.MarshalToString(fv.Interface())
		return data
	}
}

func format(name string, input []*mapIndexValue) []*mapIndexValue {
	result := &mapIndexValue{
		data: map[string]interface{}{},
	}
	for _, embMap := range input {
		for embName, embValue := range embMap.data {
			embName = fmt.Sprintf("%s(%s)", name, embName)
			v, ok := result.data[embName]
			if !ok {
				result.data[embName] = embValue
				continue
			}
			result.data[embName] = fmt.Sprintf("%v,%v", v, embValue)
		}
	}
	return []*mapIndexValue{result}
}
