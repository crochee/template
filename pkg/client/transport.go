package client

import (
	"net/http"
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

var DefaultTransport Transport = NewTransporter(
	OriginalRequest{},
	CurlRoundTripperWithFault(),
	ResponseHandler{},
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
