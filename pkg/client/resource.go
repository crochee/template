package client

import "net/http"

type IRequest interface {
	AddEndpoint(endpoint string) IRequest
	AddPath(path string) IRequest
	To() Transport
}

func NewResource() IRequest {
	return &resource{}
}

type resource struct {
	resource string
	endpoint string
}

func (r *resource) AddEndpoint(endpoint string) IRequest {
	return &resource{
		resource: r.resource,
		endpoint: endpoint,
	}
}

func (r *resource) AddPath(path string) IRequest {
	return &resource{
		resource: path,
		endpoint: r.endpoint,
	}
}

func (r *resource) To() Transport {
	return &fixTransport{
		t:        DefaultTransport,
		resource: r.resource,
		endpoint: r.endpoint,
	}
}

type fixTransport struct {
	t        Transport
	resource string
	endpoint string
}

func (t *fixTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return t.t.RoundTrip(request)
}

func (t *fixTransport) WithRequest(requester Requester) Transport {
	return &fixTransport{
		t:        t.t.WithRequest(requester),
		resource: t.resource,
		endpoint: t.endpoint,
	}
}

func (t *fixTransport) WithClient(roundTripper http.RoundTripper) Transport {
	return &fixTransport{
		t:        t.t.WithClient(roundTripper),
		resource: t.resource,
		endpoint: t.endpoint,
	}
}

func (t *fixTransport) WithResponse(response Response) Transport {
	return &fixTransport{
		t:        t.t.WithResponse(response),
		resource: t.resource,
		endpoint: t.endpoint,
	}
}

func (t *fixTransport) Request() Requester {
	return t.t.Request()
}

func (t *fixTransport) Response() Response {
	return t.t.Response()
}

func (t *fixTransport) Client() http.RoundTripper {
	return t.t.Client()
}

func (t *fixTransport) Method(method string) RESTClient {
	return NewRESTClient(t.t, method).Endpoints(t.endpoint).Resource(t.resource)
}
