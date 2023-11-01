package middlewares

import (
	"github.com/gin-gonic/gin"

	"template/pkg/code"
	"template/pkg/resp"
	"template/pkg/utils/v"
)

type option struct {
	sources   []string
	isPrivate bool
}

type Option func(*option)

// WithSource set sources to check header X-Source
func WithSource(sources []string) Option {
	return func(o *option) {
		o.sources = sources
	}
}

// WithIsPrivate set isPrivate to check whether request is need skip
func WithIsPrivate(isPrivate bool) Option {
	return func(o *option) {
		o.isPrivate = isPrivate
	}
}

// 目前的运维接口主要是提供给云警使用， 要求调用时请求头需要带X-Source: msp
// 控制台调用时， 要求请求头需要带X-Source: console
// CSK团队调用时， 要求请求头需要带X-Source: csk
// DSK团队调用时， 要求请求头需要带X-Source: dsk
// NAS团队调用时， 要求请求头需要带X-Source: nas
// 云边高速团队调用时， 要求请求头需要带X-Source: ceen
// paas团队调用时， 要求请求头需要带X-Source: paas
var defaultSources = []string{
	v.HeaderXSourceValueConsole,
	v.HeaderXSourceValueMsp,
	v.HeaderXSourceValueCsk,
	v.HeaderXSourceValueDsk,
	v.HeaderXSourceValueNas,
	v.HeaderXSourceValueCeen,
	v.HeaderXSourceValuePaas,
}

func CheckSource(opts ...Option) gin.HandlerFunc {
	o := &option{}
	for _, opt := range opts {
		opt(o)
	}
	//去重和去空
	newSources := make([]string, 0, len(o.sources))
	for _, source := range o.sources {
		isContain := false
		for _, s := range defaultSources {
			if s == source {
				isContain = true
				break
			}
		}
		// 若没有配置allowed_x_source_headers, configSources空字符串，所以需要跳过空字符串的判断
		if !isContain && source != "" {
			newSources = append(newSources, source)
		}
	}
	newSources = append(newSources, defaultSources...)

	o.sources = newSources

	return func(c *gin.Context) {
		// 私有云环境不需要验证
		if !o.isPrivate {
			return
		}
		allowed := false
		for _, source := range o.sources {
			if c.GetHeader(v.HeaderSource) == source {
				allowed = true
			}
		}
		if !allowed {
			tipMsg := "not allow to access this api, please check your header in the request"
			resp.Error(c, code.ErrForbidden.WithResult(tipMsg))
			return
		}
	}
}
