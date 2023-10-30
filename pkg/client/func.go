package client

import (
	"net/http"

	"github.com/pkg/errors"

	"template/pkg/code"
	"template/pkg/json"
)

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
