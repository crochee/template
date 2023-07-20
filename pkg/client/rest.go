package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.uber.org/multierr"

	"template/pkg/client/query"
	"template/pkg/json"
	"template/pkg/utils"
)

type RESTClient interface {
	Endpoints(endpoint string) RESTClient

	Prefix(segments ...string) RESTClient

	Suffix(segments ...string) RESTClient

	Resource(resource string) RESTClient

	Name(resourceName string) RESTClient

	SubResource(subResources ...string) RESTClient

	Param(paramName string, value ...interface{}) RESTClient

	ParamAny(value interface{}) RESTClient

	SetHeader(key string, values ...string) RESTClient

	Headers(header http.Header) RESTClient

	Body(obj interface{}) RESTClient

	Retry(attempts int, interval time.Duration,
		shouldRetryFunc func(*http.Response, error) bool) RESTClient

	Do(ctx context.Context, result interface{}, opts ...Func) error
	DoNop(ctx context.Context, opts ...Func) error
	DoReader(ctx context.Context) (io.Reader, error)
}

// NameMayNotBe specifies strings that cannot be used as names specified as path segments (like the REST API or etcd store)
var NameMayNotBe = []string{".", ".."}

// NameMayNotContain specifies substrings that cannot be used in names specified as path segments (like the REST API or etcd store)
var NameMayNotContain = []string{"/", "%"}

// IsValidPathSegmentName validates the name can be safely encoded as a path segment
func IsValidPathSegmentName(name string) []string {
	for _, illegalName := range NameMayNotBe {
		if name == illegalName {
			return []string{fmt.Sprintf(`may not be '%s'`, illegalName)}
		}
	}

	var errors []string
	for _, illegalContent := range NameMayNotContain {
		if strings.Contains(name, illegalContent) {
			errors = append(errors, fmt.Sprintf(`may not contain '%s'`, illegalContent))
		}
	}

	return errors
}

type restfulClient struct {
	c Transport

	baseURL *url.URL
	// generic components accessible via method setters
	verb       string
	pathPrefix string
	subPath    string
	params     url.Values
	headers    http.Header
	// retry
	attempts        int
	interval        time.Duration
	shouldRetryFunc func(*http.Response, error) bool
	// structural elements of the request that are part of the Kubernetes API conventions
	resource     string
	resourceName string
	subresource  string

	// output
	err  error
	body io.Reader
}

func (r *restfulClient) addError(err error) RESTClient {
	r.err = multierr.Append(r.err, err)
	return r
}

func (r *restfulClient) Endpoints(endpoint string) RESTClient {
	if endpoint == "" {
		return r
	}
	baseURL, err := url.Parse(endpoint)
	if err != nil {
		return r.addError(err)
	}
	r.baseURL = baseURL
	return r
}

func (r *restfulClient) Prefix(segments ...string) RESTClient {
	r.pathPrefix = path.Join(r.pathPrefix, path.Join(segments...))
	return r
}

func (r *restfulClient) Suffix(segments ...string) RESTClient {
	r.subPath = path.Join(r.subPath, path.Join(segments...))
	return r
}

func (r *restfulClient) Resource(resource string) RESTClient {
	if r.resource != "" {
		return r.addError(fmt.Errorf("resource already set to %q, cannot change to %q", r.resource, resource))
	}
	if reasons := IsValidPathSegmentName(resource); len(reasons) != 0 {
		return r.addError(fmt.Errorf("invalid resource %q: %v", resource, reasons))
	}
	r.resource = resource
	return r
}

func (r *restfulClient) Name(resourceName string) RESTClient {
	if resourceName == "" {
		return r.addError(fmt.Errorf("resource name may not be empty"))
	}
	if r.resourceName != "" {
		r.err = fmt.Errorf("resource name already set to %q, cannot change to %q", r.resourceName, resourceName)
		return r
	}
	if reasons := IsValidPathSegmentName(resourceName); len(reasons) != 0 {
		return r.addError(fmt.Errorf("invalid resource name %q: %v", resourceName, reasons))
	}
	r.resourceName = resourceName
	return r
}

func (r *restfulClient) SubResource(subResources ...string) RESTClient {
	subresource := path.Join(subResources...)
	if r.subresource != "" {
		return r.addError(fmt.Errorf("subresource already set to %q, cannot change to %q", r.subresource, subresource))
	}
	for _, s := range subResources {
		if reasons := IsValidPathSegmentName(s); len(reasons) != 0 {
			return r.addError(fmt.Errorf("invalid subresource %q: %v", s, reasons))
		}
	}
	r.subresource = subresource
	return r
}

func (r *restfulClient) Param(key string, values ...interface{}) RESTClient {
	if key == "" {
		return r
	}
	if r.params == nil {
		r.params = make(url.Values)
	}
	if len(values) == 0 {
		r.params.Del(key)
		return r
	}
	for _, value := range values {
		if value == nil || value == "" {
			continue
		}
		r.params.Add(key, utils.ToString(value))
	}
	return r
}

func (r *restfulClient) ParamAny(value interface{}) RESTClient {
	if value == nil {
		return r
	}
	form, err := query.Values(value)
	if err != nil {
		return r.addError(err)
	}
	if r.params == nil {
		r.params = form
		return r
	}
	r.params = make(url.Values, len(form))
	for key, srcValues := range form {
		dstValues, ok := r.params[key]
		if !ok {
			for _, srcValue := range srcValues {
				r.params.Add(key, srcValue)
			}
			continue
		}
		for _, srcValue := range srcValues {
			found := false
			for _, dstValue := range dstValues {
				if srcValue == dstValue {
					found = true
					break
				}
			}
			if !found {
				r.params.Add(key, srcValue)
			}
		}
	}
	return r
}

func (r *restfulClient) SetHeader(key string, values ...string) RESTClient {
	if key == "" {
		return r
	}
	if r.headers == nil {
		r.headers = http.Header{}
	}
	if len(values) == 0 {
		r.headers.Del(key)
		return r
	}
	for _, value := range values {
		if value == "" {
			continue
		}
		r.headers.Add(key, value)
	}
	return r
}

func (r *restfulClient) Headers(header http.Header) RESTClient {
	if r.headers == nil {
		r.headers = header
		return r
	}
	r.headers = make(http.Header, len(header))
	for key, srcValues := range header {
		dstValues, ok := r.headers[key]
		if !ok {
			for _, srcValue := range srcValues {
				r.headers.Add(key, srcValue)
			}
			continue
		}
		for _, srcValue := range srcValues {
			found := false
			for _, dstValue := range dstValues {
				if srcValue == dstValue {
					found = true
					break
				}
			}
			if !found {
				r.headers.Add(key, srcValue)
			}
		}
	}
	return r
}

func (r *restfulClient) Retry(attempts int, interval time.Duration,
	shouldRetryFunc func(*http.Response, error) bool) RESTClient {
	r.attempts = attempts
	r.interval = interval
	r.shouldRetryFunc = shouldRetryFunc
	return r
}

func (r *restfulClient) newBackOff() backoff.BackOff {
	if r.attempts < 2 || r.interval <= 0 {
		return &backoff.ZeroBackOff{}
	}

	b := backoff.NewExponentialBackOff()
	b.InitialInterval = r.interval
	b.Multiplier = math.Pow(2, 1/float64(r.attempts-1))
	b.Reset()
	return b
}

func (r *restfulClient) roundTrip(req *http.Request, operate func(*http.Request) (*http.Response, error)) (*http.Response, error) {
	// 需要重试的条件
	if r.attempts > 0 {
		body := req.Body
		defer body.Close()
		req.Body = io.NopCloser(body)
	}
	var attempts int
	backOff := r.newBackOff() // 退避算法 保证时间间隔为指数级增长
	currentInterval := 0 * time.Millisecond
	t := time.NewTimer(currentInterval)
	for {
		select {
		case <-t.C:
			shouldRetry := attempts < r.attempts
			resp, err := operate(req)
			if !shouldRetry || (r.shouldRetryFunc != nil && !r.shouldRetryFunc(resp, err)) {
				t.Stop()
				return resp, err
			}
			// 计算下一次
			currentInterval = backOff.NextBackOff()
			attempts++
			// 定时器重置
			t.Reset(currentInterval)
		case <-req.Context().Done():
			t.Stop()
			return nil, req.Context().Err()
		}
	}
}

func (r *restfulClient) Body(body interface{}) RESTClient {
	if body == nil {
		return r
	}
	switch data := body.(type) {
	case string:
		r.body = strings.NewReader(data)
	case []byte:
		r.body = bytes.NewReader(data)
	case io.Reader:
		r.body = data
	default:
		content, err := json.Marshal(data)
		if err != nil {
			return r.addError(err)
		}
		r.SetHeader("Content-Type", "application/json; charset=utf-8")
		r.body = bytes.NewReader(content)
	}
	return r
}

func (r *restfulClient) Do(ctx context.Context, result interface{}, opts ...Func) error {
	if r.err != nil {
		return r.err
	}
	uri := r.URL().String()
	req, err := r.c.Request().Build(ctx, r.verb, uri, r.body, r.headers)
	if err != nil {
		return err
	}
	var resp *http.Response
	if resp, err = r.roundTrip(req, r.c.RoundTrip); err != nil {
		return err
	}
	defer resp.Body.Close()
	return r.c.Response().Parse(resp, result, opts...)
}

func (r *restfulClient) DoNop(ctx context.Context, opts ...Func) error {
	return r.Do(ctx, nil, opts...)
}

func (r *restfulClient) DoReader(ctx context.Context) (io.Reader, error) {
	if r.err != nil {
		return nil, r.err
	}
	uri := r.URL().String()
	req, err := r.c.Request().Build(ctx, r.verb, uri, r.body, r.headers)
	if err != nil {
		return nil, err
	}
	var resp *http.Response
	if resp, err = r.roundTrip(req, r.c.RoundTrip); err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var buffer bytes.Buffer
	if _, err = io.Copy(&buffer, resp.Body); err != nil {
		return nil, err
	}
	return &buffer, nil
}

func (r *restfulClient) URL() *url.URL {
	p := r.pathPrefix
	if r.resource != "" {
		p = path.Join(p, r.resource)
	}
	// Join trims trailing slashes, so preserve r.pathPrefix's trailing slash for backwards compatibility if nothing was changed
	if r.resourceName != "" || r.subPath != "" || r.subresource != "" {
		p = path.Join(p, r.resourceName, r.subresource, r.subPath)
	}
	finalURL := &url.URL{}
	if r.baseURL != nil {
		*finalURL = *r.baseURL
	}
	finalURL.Path = path.Join(finalURL.Path, p)
	queryValues := url.Values{}
	for key, values := range r.params {
		for _, value := range values {
			queryValues.Add(key, value)
		}
	}
	finalURL.RawQuery = queryValues.Encode()
	return finalURL
}
