package client

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

type Option func(*option)

type option struct {
	t *http.Transport
}

// TLSConfig 修改tls
func TLSConfig(cfg *tls.Config) Option {
	return func(o *option) { o.t.TLSClientConfig = cfg }
}

// Timeout 修改超时时间
func Timeout(t time.Duration) Option {
	return func(o *option) { o.t.ResponseHeaderTimeout = t }
}

// NewStandardClient 标准库client
func NewStandardClient(opts ...Option) *standardClient {
	o := &option{
		t: &http.Transport{
			MaxIdleConnsPerHost: 100,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 300 * time.Second,
			ForceAttemptHTTP2:     true,
		},
	}
	for _, opt := range opts {
		opt(o)
	}
	return &standardClient{client: &http.Client{Transport: o.t}}
}

type standardClient struct {
	client *http.Client
}

func (s *standardClient) Do(req *http.Request) (*http.Response, error) {
	return s.client.Do(req)
}
