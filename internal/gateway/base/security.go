package base

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"template/pkg/client"
	"template/pkg/code"
	jsonx "template/pkg/json"
)

type SecurityParser struct {
}

func (SecurityParser) Parse(resp *http.Response, result interface{}, opts ...client.Func) error {
	for _, opt := range opts {
		if err := opt(resp); err != nil {
			return err
		}
	}
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if contentType := resp.Header.Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		return errors.WithStack(code.ErrParseContent.WithResult(
			fmt.Sprintf("can't parse content-type %s", contentType)))
	}
	if resp.StatusCode != http.StatusOK {
		return code.From(resp).WithStatusCode(resp.StatusCode)
	}

	var response struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := jsonx.DecodeUseNumber(resp.Body, &response); err != nil {
		return errors.WithStack(code.ErrParseContent.WithResult(err))
	}
	if response.Code != http.StatusOK {
		return errors.WithStack(code.ErrInternalServerError.WithStatusCode(response.Code).WithMessage(response.Message))
	}
	if result == nil {
		return nil
	}
	if response.Data == nil {
		return errors.WithStack(code.ErrParseContent.WithResult("result is nil"))
	}
	if err := jsonx.UnmarshalNumber(response.Data, result); err != nil {
		return errors.WithStack(code.ErrParseContent.WithResult(err))
	}
	return nil
}
