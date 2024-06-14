package base

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/pkg/errors"

	"template/pkg/client"
	"template/pkg/code"
	jsonx "template/pkg/json"
)

type Response struct {
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Result  *json.RawMessage `json:"result,omitempty"`
}

type Parser struct {
}

func (p Parser) Parse(resp *http.Response, result interface{}, opts ...client.Func) error {
	for _, opt := range opts {
		if err := opt(resp); err != nil {
			return err
		}
	}
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	var response Response
	if err := jsonx.DecodeUseNumber(resp.Body, &response); err != nil {
		return errors.WithStack(code.ErrParseContent.WithResult(err))
	}
	if response.Code != strconv.Itoa(http.StatusOK) {
		err := code.Froze(response.Code, response.Message)
		if response.Result != nil {
			err = err.WithResult(string(*response.Result))
		}
		return errors.WithStack(err)
	}
	if result == nil {
		return nil
	}
	if response.Result == nil {
		return errors.WithStack(code.ErrParseContent.WithResult("result is nil"))
	}
	if err := jsonx.UnmarshalNumber(*response.Result, result); err != nil {
		return errors.WithStack(code.ErrParseContent.WithResult(err))
	}
	return nil
}
