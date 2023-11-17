package middlewares

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"template/pkg/cache"
	"template/pkg/logger/gormx"
	"template/pkg/utils/v"
)

const prefix = "cache:"

var (
	Headers = []string{
		"User-Agent",
		"Accept",
		"Accept-Encoding",
		"Accept-Language",
		"Cookie",
		"User-Agent",
		"X-Account-ID",
		"X-User-ID",
	}

	CacheNoStore = "no-store"
	CacheNoCache = "no-cache"
	CachePublic  = "public "
)

type wrappedWriter struct {
	gin.ResponseWriter
	buffer bytes.Buffer
}

func (rw *wrappedWriter) Write(body []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(body)
	if err == nil {
		rw.buffer.Write(body)
	}
	return n, err
}

// CacheMiddleware 缓存中间件
// 1.只针对GET请求
// 2.跟浏览器缓存保持一致  请求头Cache-Control取no-store和no-cache,区分不需要缓存和有缓存(默认)
// 3.原理：根据rwapath和指定的头生成key,缓存key格式：cache:服务名:key,获取请求头Cache-Control的值
// 走缓存(no-cache)的情况下，查询有缓存，有则返回，否则就走api获取数据并存入cache,不走缓存(no-store)的情况，直接走api，后尝试删除key
func CacheMiddleware(clientFunc func() cache.CacheInterface, serviceName string,
	from func(context.Context) gormx.Logger) gin.HandlerFunc {
	client := clientFunc()
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}
		ctx := c.Request.Context()
		// 生成etag和key
		etag := retrieveEtag(c)
		cacheKey := retrieveCacheKey(etag, serviceName)

		// 不使用缓存
		if c.Request.Header.Get(v.HeaderCacheControl) == CacheNoStore {
			c.Header(v.HeaderCacheControl, CacheNoStore)
			c.Next()
			if err := client.Del(ctx, cacheKey); err != nil {
				from(ctx).Errorf("err:%+v", err)
			}
			return
		}
		// 请求头Cache-Control默认值no-cache
		value, err := client.Get(ctx, cacheKey)
		if err == nil {
			// cache hit
			c.Writer.Header().Add(v.HeaderCacheControl, CacheNoCache)
			c.Writer.Header().Add(v.HeaderCacheControl, CachePublic)
			c.Writer.Header().Add(v.HeaderContentType, value.ContentType)
			var expires time.Duration
			if expires, err = client.TTL(ctx, cacheKey); err == nil {
				c.Writer.Header().Add(v.HeaderCacheControl, fmt.Sprintf("max-age=%d", int64(expires.Seconds())))
			} else {
				from(ctx).Errorf("err:%+v", err)
			}
			c.Writer.WriteHeader(value.Status)
			_, _ = c.Writer.Write(value.Body)
			c.Abort()
			return
		}
		if !errors.Is(err, cache.ErrNil) {
			from(ctx).Errorf("err:%+v", err)
		}
		// cache miss
		c.Writer.Header().Add(v.HeaderCacheControl, CacheNoCache)
		c.Writer.Header().Add(v.HeaderCacheControl, CachePublic)
		c.Writer.Header().Add(v.HeaderCacheControl, fmt.Sprintf("max-age=%d", 300))
		writer := c.Writer
		rw := &wrappedWriter{ResponseWriter: c.Writer}
		c.Writer = rw
		c.Next()
		c.Writer = writer
		if err = client.Set(ctx, cacheKey, &cache.Value{
			Status:      rw.Status(),
			ContentType: c.Writer.Header().Get(textproto.CanonicalMIMEHeaderKey(v.HeaderContentType)),
			Body:        rw.buffer.Bytes(),
		}, 5*time.Minute,
		); err != nil {
			from(ctx).Errorf("err:%+v", err)
		}
	}
}

func retrieveCacheKey(etag, serviceName string) string {
	var buf bytes.Buffer
	buf.WriteString(prefix)
	buf.WriteString(serviceName)
	buf.WriteByte(':')
	buf.WriteString(etag)
	return buf.String()
}

func retrieveEtag(ctx *gin.Context) string {
	var buf bytes.Buffer
	buf.WriteString(ctx.Request.URL.RequestURI()) // path+query
	for _, k := range Headers {                   // header
		k = textproto.CanonicalMIMEHeaderKey(k)
		if h, ok := ctx.Request.Header[k]; ok {
			buf.WriteString(k)
			buf.WriteString(strings.Join(h, ""))
		}
	}
	return md5String(buf.String())
}

func md5String(url string) string {
	h := md5.New()
	_, _ = io.WriteString(h, url)
	return hex.EncodeToString(h.Sum(nil))
}
