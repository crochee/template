package base

import (
	"context"
	"net/http"

	"github.com/spf13/viper"

	"template/internal/ctxw"
	"template/internal/util/v"
)

type IfpRequest struct{}

func (i IfpRequest) Build(ctx context.Context, method, url string, body interface{}, headers http.Header) (*http.Request, error) {
	// 设置请求头
	co := NewCoPartner(viper.GetString("ifp.ak"), viper.GetString("ifp.sk"))
	req, err := co.Build(ctx, method, url, body, headers)
	if err != nil {
		return nil, err
	}

	req.Header.Set(v.HeaderAccountID, ctxw.GetAccountID(ctx))
	req.Header.Set(v.HeaderUserID, ctxw.GetUserID(ctx))
	req.Header.Set(v.HeaderTraceID, ctxw.GetTraceID(ctx))
	req.Header.Set("Accept", "application/json")
	return req, nil
}
