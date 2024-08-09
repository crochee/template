package client

import (
	"net"
	"net/http"
	"time"
)

type Transport interface {
	http.RoundTripper
	WithRequest(requester Requester) Transport
	WithClient(roundTripper http.RoundTripper) Transport
	WithResponse(response Response) Transport

	Request() Requester
	Response() Response
	Client() http.RoundTripper

	Method(string) RESTClient
}

var (
	DefaultTransport Transport = NewTransporter(
		OriginalRequest{},
		CurlRoundTripperWithFault(),
		ResponseHandler{},
	)

	StdTransport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2: true,
		// 最好的情况是略大于MaxIdleConnsPerHost
		MaxIdleConns:          3000,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		// 当前客户端因大量并发协程调用导致被调用端端口用尽，出现cannot assign requested address报错，导致服务不可用
		// 从http.DefaultTransport拷贝默认配置，并指定MaxConnsPerHost参数用来限制底层连接池最大数目，复用连接
		MaxConnsPerHost: 5000,
		// NOTE: 默认为2的情况下，面对暴增连接会出现非常多的关闭情况，导致大量TIME_WAIT的情况
		MaxIdleConnsPerHost: 2500,
	}
)

type Transporter struct {
	req          Requester
	roundTripper http.RoundTripper
	resp         Response
}

func NewTransporter(req Requester, roundTripper http.RoundTripper, resp Response) Transport {
	return &Transporter{
		req:          req,
		roundTripper: roundTripper,
		resp:         resp,
	}
}

func (t *Transporter) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.roundTripper.RoundTrip(req)
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

func (t *Transporter) Client() http.RoundTripper {
	return t.roundTripper
}

func (t *Transporter) Method(method string) RESTClient {
	return NewRESTClient(t, method)
}
