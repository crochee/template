package client

import (
	"context"
	"net/http"
)

type ModifiableRequest struct {
	ModifyRequest func(*http.Request)
	Req           Requester
}

func (m ModifiableRequest) Build(ctx context.Context, method, url string, body interface{}, headers http.Header) (*http.Request, error) {
	// 设置请求头
	req, err := m.Req.Build(ctx, method, url, body, headers)
	if err != nil {
		return nil, err
	}
	m.ModifyRequest(req)
	return req, nil
}
