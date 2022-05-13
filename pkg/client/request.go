package client

import (
	"bytes"
	"context"
	"net/http"

	"moul.io/http2curl"

	"github.com/crochee/devt/pkg/logger"
)

type Request interface {
	Build(ctx context.Context, method, url string, body []byte, headers http.Header) (*http.Request, error)
}

type originalRequest struct{}

func (o originalRequest) Build(ctx context.Context, method, url string, body []byte, headers http.Header) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	for key, values := range headers {
		for _, v := range values {
			req.Header.Set(key, v)
		}
	}
	return req, nil
}

type curlRequest struct {
	originalRequest
}

func (c curlRequest) Build(ctx context.Context, method, url string, body []byte, headers http.Header) (*http.Request, error) {
	req, err := c.originalRequest.Build(ctx, method, url, body, headers)
	if err != nil {
		return nil, err
	}
	// 打印curl语句，便于问题分析和定位
	var curl *http2curl.CurlCommand
	if curl, err = http2curl.GetCurlCommand(req); err == nil {
		logger.From(ctx).Debug(curl.String())
	} else {
		logger.From(ctx).Error(err.Error())
	}
	return req, nil
}
