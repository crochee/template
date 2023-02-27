package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"moul.io/http2curl"

	"template/pkg/logger"
	"template/pkg/msg/server"
)

// CurlRoundTripper 使用无埋点信息的客户端
func CurlRoundTripper() http.RoundTripper {
	return &CurlTransporter{
		roundTripper: http.DefaultTransport,
	}
}

// CurlRoundTripperWithFault 使用有埋点信息的客户端
func CurlRoundTripperWithFault() http.RoundTripper {
	return &TransporterWithFault{
		roundTripper: http.DefaultTransport,
	}
}

type CurlTransporter struct {
	roundTripper http.RoundTripper
}

func (t *CurlTransporter) RoundTrip(req *http.Request) (*http.Response, error) {
	// 打印curl语句，便于问题分析和定位
	spanID := uuid.NewV4().String()
	curl, err := http2curl.GetCurlCommand(req)
	if err == nil {
		logger.From(req.Context()).Info("going to do Request",
			zap.String("span_id", spanID),
			zap.String("Request", curl.String()),
		)
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
		zap.Any("headers", resp.Header),
		zap.ByteString("Response", response),
	)
	// Reset resp.Body so it can be use again
	resp.Body = io.NopCloser(bytes.NewBuffer(response))
	return resp, nil
}

type TransporterWithFault struct {
	roundTripper http.RoundTripper
}

func (t *TransporterWithFault) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	// 打印curl语句，便于问题分析和定位
	spanID := uuid.NewV4().String()
	curl, err := http2curl.GetCurlCommand(req)
	if err == nil {
		logger.From(req.Context()).Info("going to do Request",
			zap.String("span_id", spanID),
			zap.String("Request", curl.String()),
		)
	}
	server.Merge(ctx, curl.String())

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
		zap.Any("headers", resp.Header),
		zap.ByteString("Response", response),
	)
	server.Merge(ctx, fmt.Sprintf("request: %s, status: %d, response: %s",
		curl, resp.StatusCode, string(response)))

	// Reset resp.Body so it can be use again
	resp.Body = io.NopCloser(bytes.NewBuffer(response))
	return resp, nil
}
