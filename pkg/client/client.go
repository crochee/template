package client

import (
	"context"
	"net/http"
	"net/url"
)

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewClient(prefix string, opts ...RespOpt) *Clients {
	return &Clients{
		url:    NewURLHandler(prefix),
		req:    originalRequest{},
		client: NewStandardClient(),
		resp:   NewDecoder(opts...),
	}
}

type Clients struct {
	url    URLHandler
	req    Request
	client Client
	resp   Response
}

func (c *Clients) Send(ctx context.Context, method, path string, params url.Values,
	body []byte, headers http.Header, value interface{}, opts ...RespOpt) error {
	req, err := c.req.Build(ctx, method, c.url.URLWithQuery(ctx, path, params),
		body, c.url.HeaderFrom(ctx, headers))
	if err != nil {
		return err
	}
	var resp *http.Response
	if resp, err = c.client.Do(req); err != nil {
		return err
	}
	defer resp.Body.Close()
	return c.resp.Unmarshal(resp, value, opts...)
}

func (c *Clients) Get(ctx context.Context, path string, params url.Values, headers http.Header, value interface{}, opts ...RespOpt) error {
	return c.Send(ctx, http.MethodGet, path, params, nil, headers, value, opts...)
}

func (c *Clients) Post(ctx context.Context, path string, params url.Values,
	body []byte, headers http.Header, value interface{}, opts ...RespOpt) error {
	return c.Send(ctx, http.MethodPost, path, params, body, headers, value, opts...)
}

func (c *Clients) Put(ctx context.Context, path string, params url.Values,
	body []byte, headers http.Header, value interface{}, opts ...RespOpt) error {
	return c.Send(ctx, http.MethodPut, path, params, body, headers, value, opts...)
}

func (c *Clients) Delete(ctx context.Context, path string, params url.Values,
	headers http.Header, value interface{}, opts ...RespOpt) error {
	return c.Send(ctx, http.MethodDelete, path, params, nil, headers, value, opts...)
}

func (c *Clients) Patch(ctx context.Context, path string, params url.Values,
	body []byte, headers http.Header, value interface{}, opts ...RespOpt) error {
	return c.Send(ctx, http.MethodPatch, path, params, body, headers, value, opts...)
}
