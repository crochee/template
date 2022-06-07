package message

import (
	"context"

	jsoniter "github.com/json-iterator/go"

	"go_template/pkg/validator"
)

type option struct {
	handlerFunc func(context.Context, []byte) error
	marshal     MarshalAPI // mq  assemble request or response
	handler     jsoniter.API
	validator   validator.Validator
}

type Option func(*option)

func WithFunc(marshal MarshalAPI) Option {
	return func(o *option) {
		o.marshal = marshal
	}
}

func WithMarshalAPI(f func(context.Context, []byte) error) Option {
	return func(o *option) {
		o.handlerFunc = f
	}
}

func WithJSON(handler jsoniter.API) Option {
	return func(o *option) {
		o.handler = handler
	}
}

func WithValidator(validator validator.Validator) Option {
	return func(o *option) {
		o.validator = validator
	}
}
