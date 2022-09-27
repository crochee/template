package client

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"go_template/pkg/code"
	"go_template/pkg/json"
)

type Response interface {
	Parse(resp *http.Response, result interface{}, opts ...OptionFunc) error
}

type OptionFunc func(*Option)

type Option struct {
	ExpectStatusCode int
	Escape           func(response *http.Response) code.ErrorCode
}

func ExpectOK(status int) OptionFunc {
	return func(o *Option) { o.ExpectStatusCode = status }
}

func ErrorParser(f func(*http.Response) code.ErrorCode) OptionFunc {
	return func(o *Option) { o.Escape = f }
}

type originalResponse struct {
	opts *Option
}

func (o *originalResponse) Parse(resp *http.Response, result interface{}, opts ...OptionFunc) error {
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	for _, f := range opts {
		f(o.opts)
	}
	if resp.StatusCode != o.opts.ExpectStatusCode {
		if contentType := resp.Header.Get("Content-Type"); !strings.Contains(contentType, "application/json") {
			return errors.WithStack(code.ErrParseContent.WithResult(
				fmt.Sprintf("can't parse content-type %s", contentType)))
		}
		return o.opts.Escape(resp).WithStatusCode(resp.StatusCode)
	}
	rv := reflect.ValueOf(result)
	if rv.IsNil() {
		return nil
	}
	if rv.Kind() != reflect.Ptr {
		return code.ErrInternalServerError.WithResult("result is not a pointer")
	}
	if contentType := resp.Header.Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		return errors.WithStack(code.ErrParseContent.WithResult(
			fmt.Sprintf("can't parse content-type %s", contentType)))
	}
	if err := json.DecodeUseNumber(resp.Body, result); err != nil {
		return code.ErrParseContent.WithResult(err)
	}
	return nil
}

// NewDecoder create a new Response
func NewDecoder(opts ...OptionFunc) Response {
	o := &originalResponse{
		opts: &Option{
			ExpectStatusCode: http.StatusOK,
			Escape:           code.From,
		},
	}
	for _, f := range opts {
		f(o.opts)
	}
	return o
}
