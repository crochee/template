package client

import (
	"net/http"

	"github.com/pkg/errors"
	"go_template/pkg/code"

	"go_template/pkg/json"
)

type Func func(*http.Response) error

// DefaultFunc provides default error handling function implementation.
// If the implementation does not meet your needs, you can change it yourself
var DefaultFunc = []Func{ErrorFunc(http.StatusOK)}

func ErrorFunc(expectStatusCode int) func(*http.Response) error {
	return func(resp *http.Response) error {
		if resp.StatusCode != expectStatusCode {
			var result struct {
				Code    string      `json:"code"`
				Message string      `json:"message"`
				Result  interface{} `json:"result"`
			}
			if err := json.DecodeUseNumber(resp.Body, &result); err != nil {
				return errors.WithStack(code.ErrParseContent.WithResult(err.Error()))
			}
			return code.Froze(result.Code, result.Message).WithResult(result.Result)
		}
		return nil
	}
}
