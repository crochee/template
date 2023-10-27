package req

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"template/pkg/client"
	"template/pkg/code"
	jsonx "template/pkg/json"
	"template/pkg/utils"
	pv "template/pkg/utils/v"
)

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
	if body != nil {
		r, ok := body.(io.Reader)
		if ok {
			buf, err := io.ReadAll(r)
			if err != nil {
				return nil, errors.WithStack(code.ErrInternalServerError.WithResult(err.Error()))
			}
			body = buf
		}
	}
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
	// Release all request headers
	req.Header.Set(HeaderKeySignedHeader, "0")
	var signature string
	if signature, err = c.sign(method, req.URL.Query(), req.Header, body); err != nil {
		return nil, err
	}
	// Set request headers
	req.Header.Set("algorithm", "HmacSHA256")
	req.Header.Set("accessKey", c.ak)
	req.Header.Set("sign", signature)
	req.Header.Set("requestTime", strconv.FormatInt(time.Now().UnixMilli(), pv.Decimal))

	return req, nil
}

func (c coPartner) parseHeader(header http.Header, keyMap map[string][]string) (map[string][]string, error) {
	signedHeaders := header.Values(HeaderKeySignedHeader)
	if len(signedHeaders) == 0 {
		return nil, errors.WithStack(code.ErrInternalServerError.WithResult("miss signedHeader"))
	}
	if signedHeaders[0] == "0" { // Release all request headers
		return keyMap, nil
	}
	for _, k := range signedHeaders {
		keyMap[k] = header.Values(k)
	}
	return keyMap, nil
}

func (c coPartner) sign(method string, resultMap map[string][]string, header http.Header, body interface{}) (string, error) {
	switch method {
	case http.MethodGet:
		body = nil
	case http.MethodPost:
		resultMap = map[string][]string{}
	}
	bodyMap, err := c.convertBody(body)
	if err != nil {
		return "", err
	}
	for key, value := range bodyMap {
		_, ok := resultMap[key]
		if !ok {
			resultMap[key] = value
			continue
		}
		resultMap[key] = append(resultMap[key], value...)
	}
	if resultMap, err = c.parseHeader(header, resultMap); err != nil {
		return "", err
	}
	// 生成签名头
	return c.generateSign(resultMap, c.ak, c.sk), nil
}

// convertBody 将body解析为第一层的map结构
func (c coPartner) convertBody(body interface{}) (map[string][]string, error) {
	if body == nil {
		return map[string][]string{}, nil
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
	if err := jsonx.DecodeUseNumber(reader, &data); err != nil {
		return nil, errors.WithStack(code.ErrParseContent.WithResult(err))
	}
	result := make(map[string][]string, len(data))
	for key, value := range data {
		var temp interface{}
		if value != nil {
			if err := jsonx.UnmarshalNumber(value, &temp); err != nil {
				return nil, errors.WithStack(code.ErrParseContent.WithResult(err.Error()))
			}
		}
		valueStr, err := c.convertToString(temp, value)
		if err != nil {
			return nil, err
		}
		_, ok := result[key]
		if !ok {
			result[key] = []string{valueStr}
			continue
		}
		result[key] = append(result[key], valueStr)
	}
	return result, nil
}

// convertToString 将其他类型转化成字符串
func (c coPartner) convertToString(param interface{}, data json.RawMessage) (string, error) {
	value := reflect.ValueOf(param)
	kind := value.Kind()
	if kind == reflect.Ptr {
		return c.convertToString(value.Elem().Interface(), data)
	}
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(value.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(value.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(value.Float(), 'g', -1, 64), nil
	case reflect.String:
		return value.String(), nil
	case reflect.Bool:
		return fmt.Sprintf("%t", value.Bool()), nil
	default:
		result, err := data.MarshalJSON()
		if err != nil {
			return "", errors.WithStack(code.ErrInternalServerError.WithResult(err.Error()))
		}
		return utils.String(result), nil
	}
}

// generateSign 生成签名头
func (c coPartner) generateSign(params map[string][]string, accessKey, secretKey string) string {
	params["APP_KEY"] = []string{accessKey}

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf strings.Builder
	for _, key := range keys {
		val := params[key]
		l := len(val)
		if l == 0 {
			continue
		} else if l > 1 {
			sort.Strings(val)
		}
		for _, tempValue := range val {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(key)
			buf.WriteByte('=')
			buf.WriteString(tempValue)
		}
	}

	paramToSign := buf.String()
	h := hmac.New(sha256.New, utils.Bytes(secretKey))
	h.Write(utils.Bytes(paramToSign))
	return hex.EncodeToString(h.Sum(nil))
}
