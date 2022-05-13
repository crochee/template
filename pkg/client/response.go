package client

import (
	"net/http"

	jsoniter "github.com/json-iterator/go"

	"go-template/pkg/code"
)

type RespOpt func(*responseOption)

type responseOption struct {
	expectStatusCode int
	escape           func(response *http.Response) code.ErrorCode
}

type Response interface {
	Unmarshal(resp *http.Response, result interface{}, opts ...RespOpt) error
}

type originalResponse struct {
	opts *responseOption
}

func (o *originalResponse) Unmarshal(resp *http.Response, result interface{}, opts ...RespOpt) error {
	options := *o.opts
	for _, f := range opts {
		f(&options)
	}
	if resp.StatusCode != options.expectStatusCode {
		return options.escape(resp)
	}

	if result == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	decoder := jsoniter.ConfigCompatibleWithStandardLibrary.NewDecoder(resp.Body)
	decoder.UseNumber()
	if err := decoder.Decode(result); err != nil {
		return code.ErrParseContent.WithResult(err)
	}
	return nil
}

// NewDecoder create a new Response
func NewDecoder(opts ...RespOpt) Response {
	o := &originalResponse{
		opts: &responseOption{
			expectStatusCode: http.StatusOK,
			escape:           code.From,
		},
	}
	for _, f := range opts {
		f(o.opts)
	}
	return o
}
