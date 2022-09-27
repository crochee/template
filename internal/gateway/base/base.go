package base

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"go_template/internal/ctxw"
	"go_template/internal/util/v"
	"go_template/pkg/client"
	"go_template/pkg/code"
	"go_template/pkg/json"
	"go_template/pkg/utils"
	pv "go_template/pkg/utils/v"
)

type BaseResponse struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Result  *json.RawMessage `json:"result,omitempty"`
}

type Parser struct {
}

func (p Parser) Parse(resp *http.Response, result interface{}, opts ...client.OptionFunc) error {
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	var response BaseResponse
	if err := json.DecodeUseNumber(resp.Body, &response); err != nil {
		return errors.WithStack(code.ErrParseContent.WithResult(err))
	}
	if response.Code != http.StatusOK {
		err := code.ErrInternalServerError.WithMessage(response.Message)
		if response.Result != nil {
			err = err.WithResult(string(*response.Result))
		}
		return errors.WithStack(err)
	}
	rv := reflect.ValueOf(result)
	if rv.IsNil() {
		return nil
	}
	if rv.Kind() != reflect.Ptr {
		return code.ErrInternalServerError.WithResult("result is not a pointer")
	}
	if response.Result == nil {
		return errors.WithStack(code.ErrParseContent.WithResult("result is nil"))
	}
	if err := json.UnmarshalNumber(*response.Result, result); err != nil {
		return errors.WithStack(code.ErrParseContent.WithResult(err))
	}
	return nil
}

type DCSParser struct {
}

func (p DCSParser) Parse(resp *http.Response, result interface{}, opts ...client.OptionFunc) error {
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if contentType := resp.Header.Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		return errors.WithStack(code.ErrParseContent.WithResult(
			fmt.Sprintf("can't parse content-type %s", contentType)))
	}
	if resp.StatusCode != http.StatusOK {
		return code.From(resp).WithStatusCode(resp.StatusCode)
	}
	rv := reflect.ValueOf(result)
	if rv.IsNil() {
		return nil
	}
	if rv.Kind() != reflect.Ptr {
		return code.ErrInternalServerError.WithResult("result is not a pointer")
	}
	var response BaseResponse
	if err := json.DecodeUseNumber(resp.Body, &response); err != nil {
		return errors.WithStack(code.ErrParseContent.WithResult(err))
	}
	if response.Result == nil {
		return errors.WithStack(code.ErrParseContent.WithResult("result is nil"))
	}
	if err := json.UnmarshalNumber(*response.Result, result); err != nil {
		return errors.WithStack(code.ErrParseContent.WithResult(err))
	}
	return nil
}

type DCSRequest struct {
	client.OriginalRequest
}

func (d DCSRequest) Build(ctx context.Context, method, url string, body interface{}, headers http.Header) (*http.Request, error) {
	// 设置请求头
	req, err := d.OriginalRequest.Build(ctx, method, url, body, headers)
	if err != nil {
		return nil, err
	}
	req.Header.Set(v.HeaderAccountID, viper.GetString("account_id"))
	req.Header.Set(v.HeaderUserID, viper.GetString("user_id"))
	req.Header.Set(v.HeaderRealAccountID, ctxw.GetAccountID(ctx))
	req.Header.Set(v.HeaderRealUserID, ctxw.GetUserID(ctx))
	req.Header.Set(v.HeaderTraceID, ctxw.GetTraceID(ctx))
	req.Header.Set("Accept", "application/json")
	return req, nil
}

const (
	HeaderKeySignedHeader = "signedHeader"
)

// NewCoPartner 合作伙伴API请求构造
func NewCoPartner(ak, sk string) client.Requester {
	return &coPartner{
		ak:              ak,
		sk:              sk,
		OriginalRequest: client.OriginalRequest{},
	}
}

type coPartner struct {
	ak string
	sk string
	client.OriginalRequest
}

func (c coPartner) Build(ctx context.Context, method, url string, body interface{}, headers http.Header) (*http.Request, error) {
	req, err := c.OriginalRequest.Build(ctx, method, url, body, headers)
	if err != nil {
		return nil, err
	}

	if c.ak == "" {
		return nil, errors.WithStack(code.ErrInternalServerError.WithResult("miss AK"))
	}
	if c.sk == "" {
		return nil, errors.WithStack(code.ErrInternalServerError.WithResult("miss SK"))
	}

	var keyMap map[string]string
	if keyMap, err = convertBody(body); err != nil {
		return nil, err
	}
	// Release all request headers
	req.Header.Set(HeaderKeySignedHeader, "1")
	if keyMap, err = c.parseHeader(req, keyMap); err != nil {
		return nil, err
	}
	// 生成签名头
	signature := generateSign(keyMap, c.ak, c.sk)
	// Set request headers
	req.Header.Set("algorithm", "HmacSHA256")
	req.Header.Set("accessKey", c.ak)
	req.Header.Set("sign", signature)
	req.Header.Set("requestTime", strconv.FormatInt(time.Now().UnixMilli(), pv.Decimal))

	return req, nil
}

func (c coPartner) parseHeader(req *http.Request, keyMap map[string]string) (map[string]string, error) {
	signedHeaders := req.Header.Values(HeaderKeySignedHeader)
	if len(signedHeaders) == 0 {
		return nil, errors.WithStack(code.ErrInternalServerError.WithResult("miss signedHeader"))
	}
	if signedHeaders[0] == "1" { // Release all request headers
		return keyMap, nil
	}
	for _, k := range signedHeaders {
		keyMap[k] = req.Header.Get(k)
	}
	return keyMap, nil
}

// convertBody 将body解析为第一层的map结构
func convertBody(body interface{}) (map[string]string, error) {
	if body == nil {
		return map[string]string{}, nil
	}
	var reader io.Reader
	switch data := body.(type) {
	case string:
		reader = strings.NewReader(data)
	case []byte:
		reader = bytes.NewReader(data)
	case io.Reader:
		reader = data
	default:
		content, err := json.Marshal(data)
		if err != nil {
			return nil, errors.WithStack(code.ErrInternalServerError.WithResult(err.Error()))
		}
		reader = bytes.NewReader(content)
	}
	data := map[string]json.RawMessage{}
	if err := json.DecodeUseNumber(reader, &data); err != nil {
		return nil, errors.WithStack(code.ErrParseContent.WithResult(err))
	}
	result := make(map[string]string, len(data))
	for key, value := range data {
		result[key] = utils.String(value)
	}
	return result, nil
}

// generateSign 生成签名头
func generateSign(params map[string]string, accessKey, secretKey string) string {
	params["APP_KEY"] = accessKey

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf strings.Builder
	for _, key := range keys {
		val := params[key]
		if val == "" {
			continue
		}
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(key)
		buf.WriteByte('=')
		buf.WriteString(val)
	}

	paramToSign := buf.String()
	h := hmac.New(sha256.New, utils.Bytes(secretKey))
	h.Write(utils.Bytes(paramToSign))
	return hex.EncodeToString(h.Sum(nil))
}
