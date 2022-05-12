package validator

import (
	"errors"
	"reflect"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	translations "github.com/go-playground/validator/v10/translations/zh"
	"go.uber.org/multierr"
)

type Validator interface {
	ValidateStruct(obj interface{}) error
	Engine() interface{}
}

// New validator
func New() (*defaultValidator, error) {
	v := &defaultValidator{Validate: validator.New()}
	v.Validate.SetTagName("binding")
	v.translator, _ = ut.New(zh.New()).GetTranslator("zh")
	if err := translations.RegisterDefaultTranslations(v.Validate, v.translator); err != nil {
		return nil, err
	}
	return v, nil
}

func NewValidator() Validator {
	v := &defaultValidator{Validate: validator.New()}
	v.Validate.SetTagName("binding")
	return v
}

type defaultValidator struct {
	Validate   *validator.Validate
	translator ut.Translator
}

// ValidateStruct receives any kind of type, but only performed struct or pointer to struct type.
func (v *defaultValidator) ValidateStruct(obj interface{}) error {
	err := v.defaultValidateStruct(obj)
	if err == nil {
		return nil
	}
	return v.Translate(err)
}

// Translate receives struct type
func (v *defaultValidator) Translate(err error) error {
	var vErrs validator.ValidationErrors
	if !errors.As(err, &vErrs) {
		return err
	}
	var errs error
	for _, s := range vErrs.Translate(v.translator) {
		errs = multierr.Append(errs, errors.New(s))
	}
	return errs
}

// validateStruct receives struct type
func (v *defaultValidator) validateStruct(obj interface{}) error {
	return v.Validate.Struct(obj)
}

// Engine returns the underlying validator engine which powers the default
// Validator instance. This is useful if you want to register custom validations
// or struct level validations. See validator GoDoc for more info -
// https://godoc.org/gopkg.in/go-playground/validator.v8
func (v *defaultValidator) Engine() interface{} {
	return v.Validate
}

func (v *defaultValidator) defaultValidateStruct(obj interface{}) error {
	if obj == nil {
		return nil
	}
	value := reflect.ValueOf(obj)
	switch value.Kind() { // nolint:exhaustive
	case reflect.Ptr:
		return v.ValidateStruct(value.Elem().Interface())
	case reflect.Struct:
		return v.validateStruct(obj)
	case reflect.Slice, reflect.Array:
		count := value.Len()
		var errs error
		for i := 0; i < count; i++ {
			errs = multierr.Append(errs, v.ValidateStruct(value.Index(i).Interface()))
		}
		return errs
	default:
		return nil
	}
}

func RegisterValidation(v Validator, tag string, fn validator.Func, callValidationEvenIfNull ...bool) error {
	validate, ok := v.Engine().(*validator.Validate)
	if !ok {
		return nil
	}
	return validate.RegisterValidation(tag, fn, callValidationEvenIfNull...)
}

func Var(v Validator, field interface{}, tag string) error {
	validate, ok := v.Engine().(*validator.Validate)
	if !ok {
		return nil
	}
	return validate.Var(field, tag)
}
