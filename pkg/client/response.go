package client

import (
	"fmt"
	"mime"
	"net/http"

	"github.com/pkg/errors"

	"template/pkg/code"
	"template/pkg/json"
)

type Response interface {
	Parse(resp *http.Response, result interface{}, opts ...func(*http.Response) error) error
}

type ResponseHandler struct {
}

func (o ResponseHandler) Parse(resp *http.Response, result interface{}, opts ...func(*http.Response) error) error {
	for _, opt := range opts {
		if err := opt(resp); err != nil {
			return err
		}
	}
	if resp.StatusCode == http.StatusNoContent || result == nil {
		return nil
	}

	if err := o.checkContentType(resp.Header.Get("Content-Type")); err != nil {
		return err
	}
	if err := json.DecodeUseNumber(resp.Body, result); err != nil {
		return code.ErrParseContent.WithResult(err)
	}
	return nil
}

func (o ResponseHandler) checkContentType(contentType string) error {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return errors.WithStack(code.ErrInternalServerError.WithResult(err.Error()))
	}
	if mediaType != "application/json" {
		return errors.WithStack(code.ErrInternalServerError.WithMessage(
			fmt.Sprintf("can't parse content-type %s", contentType)))
	}
	return nil
}
