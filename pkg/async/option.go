package async

import (
	"context"

	jsoniter "github.com/json-iterator/go"

	"template/pkg/logger/gormx"
	"template/pkg/validator"
)

type option struct {
	manager   ManagerTaskHandler
	marshal   MarshalAPI // mq  assemble request or response
	handler   jsoniter.API
	validator validator.Validator
	autoAck   bool
	uuid      func(ctx context.Context) string
	form      func(ctx context.Context) gormx.Logger
}

type Option func(*option)

func WithManager(manager ManagerTaskHandler) Option {
	return func(o *option) {
		o.manager = manager
	}
}

func WithMarshalAPI(marshal MarshalAPI) Option {
	return func(o *option) {
		o.marshal = marshal
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

func WithAck(auto bool) Option {
	return func(o *option) {
		o.autoAck = auto
	}
}

func WithUuid(uuid func(ctx context.Context) string) Option {
	return func(o *option) {
		o.uuid = uuid
	}
}

func WithLogFrom(from func(context.Context) gormx.Logger) Option {
	return func(o *option) {
		o.form = from
	}
}
