package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"time"

	"moul.io/http2curl"

	"template/pkg/logger/gormx"
	"template/pkg/msg/server"
)

// CurlRoundTripper 使用无埋点信息的客户端
func CurlRoundTripper() *CustomTransporter {
	return &CustomTransporter{
		Merge: func(context.Context, string) {
		},
		From:         gormx.NewZapGormWriterFrom,
		RoundTripper: http.DefaultTransport,
	}
}

// CurlRoundTripperWithFault 使用有埋点信息的客户端
func CurlRoundTripperWithFault() *CustomTransporter {
	return &CustomTransporter{
		Merge:        server.Merge,
		From:         gormx.NewZapGormWriterFrom,
		RoundTripper: http.DefaultTransport,
	}
}

// CurlRoundTripperDisableCurl 使用不打印请求信息的客户端
func CurlRoundTripperDisableCurl() http.RoundTripper {
	return &CustomTransporter{
		Merge:        server.Merge,
		From:         gormx.Nop,
		RoundTripper: http.DefaultTransport,
	}
}

// CurlRoundTripperWithTls 使用无埋点信息的https客户端
func CurlRoundTripperWithTls(tls *tls.Config) *CustomTransporter {
	return &CustomTransporter{
		Merge: func(context.Context, string) {
		},
		From: gormx.NewZapGormWriterFrom,
		RoundTripper: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       tls,
		},
	}
}

// CurlRoundTripperWithTlsFault 使用有埋点信息的https客户端
func CurlRoundTripperWithTlsFault(tls *tls.Config) *CustomTransporter {
	return &CustomTransporter{
		Merge: server.Merge,
		From:  gormx.NewZapGormWriterFrom,
		RoundTripper: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       tls,
		},
	}
}

type CustomTransporter struct {
	Merge        func(context.Context, string)
	From         func(context.Context) gormx.Logger
	RoundTripper http.RoundTripper
}

func (c *CustomTransporter) RoundTrip(req *http.Request) (*http.Response, error) {
	var start, end time.Time
	ctx := req.Context()
	// 打印curl语句，便于问题分析和定位
	curl, err := http2curl.GetCurlCommand(req)
	if err != nil {
		return nil, err
	}
	content := &TransportContent{
		Request: curl.String(),
	}
	defer func() {
		contentStr := FormatContent(content)
		c.From(ctx).Infof("total cost:%dms, call Request end content %s", end.Sub(start).Milliseconds(), contentStr)
		c.Merge(ctx, contentStr)
	}()
	var resp *http.Response
	start = time.Now()
	if resp, err = c.RoundTripper.RoundTrip(req); err != nil {
		return nil, err
	}
	end = time.Now()
	defer resp.Body.Close()
	content.Status = resp.Status
	if resp.StatusCode == http.StatusNoContent {
		return resp, nil
	}
	var response []byte
	if response, err = io.ReadAll(resp.Body); err != nil {
		return nil, err
	}
	content.Response = string(response)
	// Reset resp.Body so it can be use again
	resp.Body = io.NopCloser(bytes.NewBuffer(response))
	return resp, nil
}
