package client

import (
	"context"
	"net/http"
	"net/url"
)

type URLHandler interface {
	URL(ctx context.Context, path string) string
	URLWithQuery(ctx context.Context, path string, value url.Values) string
	Header(ctx context.Context) http.Header
	HeaderFrom(ctx context.Context, header http.Header) http.Header
}

func NewURLHandler(prefix string) URLHandler {
	return DefaultIP{
		prefix: prefix,
	}
}

type DefaultIP struct {
	prefix string
}

func (d DefaultIP) Header(ctx context.Context) http.Header {
	return d.HeaderFrom(ctx, nil)
}

func (d DefaultIP) HeaderFrom(ctx context.Context, srcHeader http.Header) http.Header {
	nv := 0
	for _, vv := range srcHeader {
		nv += len(vv)
	}
	sv := make([]string, nv) // shared backing array for headers' values
	header := make(http.Header, len(srcHeader)+4)
	for k, vv := range srcHeader {
		n := copy(sv, vv)
		header[k] = sv[:n:n]
		sv = sv[n:]
	}
	return header
}

func (d DefaultIP) URLWithQuery(ctx context.Context, path string, value url.Values) string {
	if len(value) == 0 {
		return d.URL(ctx, path)
	}
	return d.URL(ctx, path) + "?" + value.Encode()
}

func (d DefaultIP) URL(ctx context.Context, path string) string {
	host := d.prefix
	if host == "" {
		host = "http://127.0.0.1:8080"
	}
	return host + path
}
