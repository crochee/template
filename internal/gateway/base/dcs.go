package base

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"template/internal/ctxw"
	"template/internal/util/v"
	"template/pkg/client"
	"template/pkg/code"
	jsonx "template/pkg/json"
)

type DCSRequest struct {
	client.OriginalRequest
}

func (d DCSRequest) Build(ctx context.Context, method, url string, body interface{}, headers http.Header) (*http.Request, error) {
	// 设置请求头
	req, err := d.OriginalRequest.Build(ctx, method, url, body, headers)
	if err != nil {
		return nil, err
	}

	req.Header.Set(v.HeaderAccountID, ctxw.GetAccountID(ctx))
	req.Header.Set(v.HeaderUserID, ctxw.GetUserID(ctx))
	req.Header.Set(v.HeaderTraceID, ctxw.GetTraceID(ctx))
	req.Header.Set("Accept", "application/json")
	return req, nil
}

type DCSParser struct {
}

func (p DCSParser) Parse(resp *http.Response, result interface{}, opts ...client.Func) error {
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
	if result == nil {
		return nil
	}

	var response Response
	if err := jsonx.DecodeUseNumber(resp.Body, &response); err != nil {
		return errors.WithStack(code.ErrParseContent.WithResult(err))
	}
	if response.Result == nil {
		return errors.WithStack(code.ErrParseContent.WithResult("result is nil"))
	}
	if err := jsonx.UnmarshalNumber(*response.Result, result); err != nil {
		return errors.WithStack(code.ErrParseContent.WithResult(err))
	}
	return nil
}
