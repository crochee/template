package client

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"template/pkg/code"
	"template/pkg/json"
)

type Requester interface {
	Build(ctx context.Context, method, url string, body interface{}, headers http.Header) (*http.Request, error)
}

type OriginalRequest struct{}

func (o OriginalRequest) Build(ctx context.Context, method, url string, body interface{}, headers http.Header) (*http.Request, error) {
	var (
		reader     io.Reader
		objectData bool
	)
	if body != nil {
		switch data := body.(type) {
		case string:
			reader = strings.NewReader(data)
		case []byte:
			reader = bytes.NewReader(data)
		case io.Reader:
			reader = data
		default:
			content, err := json.Marshal(data)
			if err != nil {
				return nil, errors.WithStack(code.ErrInternalServerError.WithResult(err.Error()))
			}
			objectData = true
			reader = bytes.NewReader(content)
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, errors.WithStack(code.ErrInternalServerError.WithResult(err.Error()))
	}
	for key, values := range headers {
		for _, value := range values {
			req.Header.Set(key, value)
		}
	}
	if objectData {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}
	return req, nil
}
