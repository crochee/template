package client

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"time"

	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"moul.io/http2curl"

	"go_template/pkg/logger"
)

type Transport interface {
	http.RoundTripper
	WithRequest(requester Requester) Transport
	WithClient(roundTripper http.RoundTripper) Transport
	WithResponse(response Response) Transport

	Request() Requester
	Response() Response

	Method(string) RESTClient
}

// DefaultTransport 默认配置的传输层实现
var DefaultTransport Transport = &Transporter{
	req: OriginalRequest{},
	roundTripper: &http.Transport{
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
	},
	resp: NewDecoder(),
}

type Transporter struct {
	req          Requester
	roundTripper http.RoundTripper
	resp         Response
}

func (t *Transporter) RoundTrip(req *http.Request) (*http.Response, error) {
	// 打印curl语句，便于问题分析和定位
	spanID := uuid.NewV4().String()
	curl, err := http2curl.GetCurlCommand(req)
	if err == nil {
		logger.From(req.Context()).Info("going to do Request",
			zap.String("span_id", spanID),
			zap.String("Request", curl.String()),
		)
	} else {
		logger.From(req.Context()).Error(err.Error())
	}

	var resp *http.Response
	if resp, err = t.roundTripper.RoundTrip(req); err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		logger.From(req.Context()).Info("get response body",
			zap.String("span_id", spanID),
			zap.Int("status", resp.StatusCode),
		)
	}
	var response []byte
	if response, err = io.ReadAll(resp.Body); err != nil {
		return nil, err
	}
	logger.From(req.Context()).Info("get response body",
		zap.String("span_id", spanID),
		zap.ByteString("Response", response),
	)
	// Reset resp.Body so it can be use again
	resp.Body = io.NopCloser(bytes.NewBuffer(response))
	return resp, nil
}

func (t *Transporter) WithRequest(requester Requester) Transport {
	return &Transporter{
		req:          requester,
		roundTripper: t.roundTripper,
		resp:         t.resp,
	}
}

func (t *Transporter) WithClient(roundTripper http.RoundTripper) Transport {
	return &Transporter{
		req:          t.req,
		roundTripper: roundTripper,
		resp:         t.resp,
	}
}

func (t *Transporter) WithResponse(response Response) Transport {
	return &Transporter{
		req:          t.req,
		roundTripper: t.roundTripper,
		resp:         response,
	}
}

func (t *Transporter) Request() Requester {
	return t.req
}

func (t *Transporter) Response() Response {
	return t.resp
}

func (t *Transporter) Method(method string) RESTClient {
	return &restfulClient{c: t, verb: method}
}
