package client

import (
	"context"
	"net/http"
)

type HeaderRequest struct {
	GetHeaders func(ctx context.Context) http.Header
	CoPartner  Requester
}

func (h HeaderRequest) Build(ctx context.Context, method, url string, body interface{}, headers http.Header) (*http.Request, error) {
	// 设置请求头
	req, err := h.CoPartner.Build(ctx, method, url, body, headers)
	if err != nil {
		return nil, err
	}

	headersFromCtx := h.GetHeaders(ctx)
	for key, values := range headersFromCtx {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	req.Header.Set("Accept", "application/json")
	return req, nil
}
