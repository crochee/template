package proxy

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"go.uber.org/zap"

	"disasterbackup/pkg/client"
	"disasterbackup/pkg/logger"
	"disasterbackup/pkg/selector"
	"disasterbackup/pkg/utils"
)

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}

func NewReverseProxy(selector selector.Selector) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		node := selector.Next()
		targetQuery := node.URL.RawQuery
		req.URL.Scheme = node.URL.Scheme
		req.URL.Host = node.URL.Host
		req.URL.Path, req.URL.RawPath = joinURLPath(&node.URL, req.URL)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "DBR")
		}
	}
	return &httputil.ReverseProxy{
		Director:       director,
		Transport:      client.DefaultTransport,
		ErrorLog:       log.New(logger.SetWriter(false), "[PROXY]", log.LstdFlags),
		BufferPool:     utils.BufPool,
		ModifyResponse: nil,
		ErrorHandler:   errHandler,
	}
}

// StatusClientClosedRequest non-standard HTTP status code for client disconnection.
const StatusClientClosedRequest = 499

// StatusClientClosedRequestText non-standard HTTP status for client disconnection.
const StatusClientClosedRequestText = "Client Closed Request"

func StatusText(statusCode int) string {
	if statusCode == StatusClientClosedRequest {
		return StatusClientClosedRequestText
	}
	return http.StatusText(statusCode)
}

func errHandler(rw http.ResponseWriter, req *http.Request, err error) {
	statusCode := http.StatusInternalServerError
	switch {
	case errors.Is(err, io.EOF):
		statusCode = http.StatusBadGateway
	case errors.Is(err, context.Canceled):
		statusCode = StatusClientClosedRequest
	default:
		var netErr net.Error
		if errors.As(err, &netErr) {
			if netErr.Timeout() {
				statusCode = http.StatusGatewayTimeout
			} else {
				statusCode = http.StatusBadGateway
			}
		}
	}
	text := StatusText(statusCode)
	logger.From(req.Context()).Error("[PROXY] failed",
		zap.Any("url", req.URL), zap.Int("statusCode", statusCode),
		zap.String("text", text), zap.Error(err))
	rw.WriteHeader(statusCode)
	if _, err = rw.Write(utils.Bytes(text)); err != nil {
		logger.From(req.Context()).Error("[PROXY] Error while writing status code", zap.Error(err))
	}
}
