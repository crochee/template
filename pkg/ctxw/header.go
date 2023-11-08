package ctxw

import (
	"context"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/trace"

	"template/pkg/code"
	log "template/pkg/logger"
	"template/pkg/set"
	"template/pkg/utils/v"
)

type (
	headerAdminIDKey       struct{}
	headerAdminNameKey     struct{}
	headerAdminPassKey     struct{}
	headerAdminSitesKey    struct{}
	headerAccountIDKey     struct{}
	headerUserIDKey        struct{}
	headerTraceIDKey       struct{}
	headerCacheControlKey  struct{}
	headerAccountNameKey   struct{}
	headerCallFromKey      struct{}
	headerOperatorIDKey    struct{}
	headerOperatorNameKey  struct{}
	headerOperatorTypeKey  struct{}
	headerCallerIDKey      struct{}
	headerCallerCodeKey    struct{}
	headerCallerUserKey    struct{}
	headerCallerExtraKey   struct{}
	headerEventIDKey       struct{}
	headerStaffKey         struct{}
	headerAPIFromKey       struct{}
	HeaderAccessTokenKey   struct{}
	headerAuthorizationKey struct{}
	headerSourceKey        struct{}
	IPKey                  struct{}
)

// Copy 新ctx继承旧ctx的上下文，特性保持为新的模式，目的为保持新的取消能力，但是需要旧的上下文
func Copy(newCtx context.Context, ctx context.Context) context.Context {
	if adminID := GetAdminID(ctx); adminID != "" {
		newCtx = SetAdminID(newCtx, adminID)
	}
	if adminSites := GetAdminSites(ctx); adminSites != nil {
		newCtx = SetAdminSites(newCtx, adminSites...)
	}
	if accountID := GetAccountID(ctx); accountID != "" {
		newCtx = SetAccountID(newCtx, accountID)
	}
	if userID := GetUserID(ctx); userID != "" {
		newCtx = SetUserID(newCtx, userID)
	}
	if traceID := GetTraceID(ctx); traceID != "" {
		newCtx = SetTraceID(newCtx, traceID)
	}
	if cacheControl := GetCacheControl(ctx); cacheControl != "" {
		newCtx = SetCacheControl(newCtx, cacheControl)
	}
	if accountName := GetAccountName(ctx); accountName != "" {
		newCtx = SetAccountName(newCtx, accountName)
	}
	if requestIP := GetIP(ctx); requestIP != "" {
		newCtx = SetIP(newCtx, requestIP)
	}
	if operatorID := GetOperatorID(ctx); operatorID != "" {
		newCtx = SetOperatorID(newCtx, operatorID)
	}
	if operatorName := GetOperatorName(ctx); operatorName != "" {
		newCtx = SetOperatorName(newCtx, operatorName)
	}
	if operatorType := GetOperatorType(ctx); operatorType != "" {
		newCtx = SetOperatorType(newCtx, operatorType)
	}
	if callerID := GetCallerID(ctx); callerID != "" {
		newCtx = SetCallerID(newCtx, callerID)
	}
	if callerCode := GetCallerCode(ctx); callerCode != "" {
		newCtx = SetCallerCode(newCtx, callerCode)
	}
	if callerUser := GetCallerUser(ctx); callerUser != "" {
		newCtx = SetCallerUser(newCtx, callerUser)
	}
	if callerExtra := GetCallerExtra(ctx); callerExtra != "" {
		newCtx = SetCallerExtra(newCtx, callerExtra)
	}
	if source := GetSource(ctx); source != "" {
		newCtx = SetSource(newCtx, source)
	}
	span := trace.SpanFromContext(ctx)
	newCtx = trace.ContextWithSpan(newCtx, span)
	return log.WithContext(newCtx, log.FromContext(ctx))
}

// NewContext 目的：为了将去除context的cancel能力，防止上游context cancel
// 缘由：主要是用于取代api携带的context,因为api的context有cancel能力，
// 而在异步的情况下api执行完毕会自动cancel,同一时间异步任务大部分情况下还没有执行完毕，不应该cancel,所以此处新起background context,在cancel功能上区别与api的，在链路继承其承载的参数
func NewContext(ctx context.Context) context.Context {
	return Copy(context.Background(), ctx)
}

// NewSimpleContext 生成一个简单的context供调用外部服务使用。
// 原因：DCS内部的context现在含有太多的请求头和多余的请求信息，
// 而这些信息都是外部服务用不到的，所以需要生成一个纯净的context。
func NewSimpleContext(ctx context.Context) context.Context {
	newCtx := context.Background()
	if accountID := GetAccountID(ctx); accountID != "" {
		newCtx = SetAccountID(newCtx, accountID)
	}
	if userID := GetUserID(ctx); userID != "" {
		newCtx = SetUserID(newCtx, userID)
	}
	if traceID := GetTraceID(ctx); traceID != "" {
		newCtx = SetTraceID(newCtx, traceID)
	}
	if requestIP := GetIP(ctx); requestIP != "" {
		newCtx = SetIP(newCtx, requestIP)
	}
	if operatorID := GetOperatorID(ctx); operatorID != "" {
		newCtx = SetOperatorID(newCtx, operatorID)
	}
	if operatorName := GetOperatorName(ctx); operatorName != "" {
		newCtx = SetOperatorName(newCtx, operatorName)
	}
	if operatorType := GetOperatorType(ctx); operatorType != "" {
		newCtx = SetOperatorType(newCtx, operatorType)
	}
	return log.WithContext(newCtx, log.FromContext(ctx))
}

func NewTraceIDContext(ctx context.Context) context.Context {
	newTraceID := uuid.NewV4().String()
	newCtx := SetTraceID(ctx, newTraceID)
	l := log.FromContext(ctx).Logger.With().Str("new_trace_id", newTraceID).Logger()
	return log.WithContext(newCtx, &log.Logger{
		Logger: l,
	})
}

func SetAccountAndUserID(ctx context.Context, accountID, userID string) context.Context {
	ctx = SetUserID(SetAccountID(ctx, accountID), userID)
	return log.WithContext(ctx, &log.Logger{Logger: log.FromContext(ctx).With().Fields(map[string]interface{}{
		"account_id": GetAccountID(ctx),
		"user_id":    GetUserID(ctx),
	}).Logger()})
}

func SetTraceIDLogger(ctx context.Context, traceId string) context.Context {
	ctx = SetTraceID(ctx, traceId)
	return log.WithContext(ctx, &log.Logger{Logger: log.FromContext(ctx).With().Str("trace_id", traceId).Logger()})
}

// SetAdminID Add adminID to context.Context.
func SetAdminID(ctx context.Context, adminID string) context.Context {
	return context.WithValue(ctx, headerAdminIDKey{}, adminID)
}

// GetAdminID Get the adminID from context.Context.
func GetAdminID(ctx context.Context) string {
	adminID, ok := ctx.Value(headerAdminIDKey{}).(string)
	if !ok {
		return ""
	}
	return adminID
}

// SetAdminName Add admin name to context.Context.
func SetAdminName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, headerAdminNameKey{}, name)
}

// GetAdminName get the admin user's name from context.Context
func GetAdminName(ctx context.Context) string {
	name, ok := ctx.Value(headerAdminNameKey{}).(string)
	if !ok {
		return ""
	}
	return name
}

// SetAdminPass Add admin password to context.Context.
func SetAdminPass(ctx context.Context, password string) context.Context {
	return context.WithValue(ctx, headerAdminPassKey{}, password)
}

// GetAdminPass get the admin user's password from context.Context
func GetAdminPass(ctx context.Context) string {
	password, ok := ctx.Value(headerAdminPassKey{}).(string)
	if !ok {
		return ""
	}
	return password
}

// EmptyAdminSites will empty the site list of admin to context.Context.
func EmptyAdminSites(ctx context.Context) context.Context {
	return context.WithValue(ctx, headerAdminSitesKey{}, []string{})
}

// SetAdminSites will set the site list of admin to context.Context.
func SetAdminSites(ctx context.Context, sites ...string) context.Context {
	return context.WithValue(ctx, headerAdminSitesKey{}, sites)
}

// GetAdminSites return the site list of admin from context.Context.
func GetAdminSites(ctx context.Context) []string {
	sites, ok := ctx.Value(headerAdminSitesKey{}).([]string)
	if !ok {
		return nil
	}
	return sites
}

// GetAdminSiteSet return the set list of admin sites from context.Context.
func GetAdminSiteSet(ctx context.Context) *set.Set {
	sites, ok := ctx.Value(headerAdminSitesKey{}).([]string)
	if !ok {
		return nil
	}
	s := set.NewSet()
	for _, site := range sites {
		s.Add(site)
	}
	return s
}

// SetAccountID Add accountID to context.Context.
func SetAccountID(ctx context.Context, accountID string) context.Context {
	return context.WithValue(ctx, headerAccountIDKey{}, accountID)
}

// GetAccountID Get the accountID from context.Context.
func GetAccountID(ctx context.Context) string {
	accountID, ok := ctx.Value(headerAccountIDKey{}).(string)
	if !ok {
		return ""
	}
	return accountID
}

// SetUserID Add userID to context.Context.
func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, headerUserIDKey{}, userID)
}

// GetUserID Get the userID from context.Context.
func GetUserID(ctx context.Context) string {
	userID, ok := ctx.Value(headerUserIDKey{}).(string)
	if !ok {
		return ""
	}
	return userID
}

// SetTraceID Add traceID to context.Context.
func SetTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, headerTraceIDKey{}, traceID)
}

// GetTraceID Get the traceID from context.Context.
func GetTraceID(ctx context.Context) string {
	traceID, ok := ctx.Value(headerTraceIDKey{}).(string)
	if !ok {
		return ""
	}
	return traceID
}

// SetCacheControl Add Cache-Control header to context.Context.
func SetCacheControl(ctx context.Context, cacheControl string) context.Context {
	return context.WithValue(ctx, headerCacheControlKey{}, cacheControl)
}

// GetCacheControl Get the Cache-Control header from context.Context.
func GetCacheControl(ctx context.Context) string {
	cacheControl, ok := ctx.Value(headerCacheControlKey{}).(string)
	if !ok {
		return ""
	}
	return cacheControl
}

// SetAccountName Add X-Account-Name header to context.Context.
func SetAccountName(ctx context.Context, accountName string) context.Context {
	return context.WithValue(ctx, headerAccountNameKey{}, accountName)
}

// GetAccountName Get the X-Account-Name header from context.Context.
func GetAccountName(ctx context.Context) string {
	cacheControl, ok := ctx.Value(headerAccountNameKey{}).(string)
	if !ok {
		return ""
	}
	return cacheControl
}

// SetCallFrom Add X-Call-From header to context.Context.
func SetCallFrom(ctx context.Context, callFrom string) context.Context {
	return context.WithValue(ctx, headerCallFromKey{}, callFrom)
}

// GetCallFrom Get the X-Call-From header from context.Context.
func GetCallFrom(ctx context.Context) string {
	callFrom, ok := ctx.Value(headerCallFromKey{}).(string)
	if !ok {
		return ""
	}
	return callFrom
}

// SetOperatorID Add `operatorId` header to context.Context.
func SetOperatorID(ctx context.Context, operatorID string) context.Context {
	return context.WithValue(ctx, headerOperatorIDKey{}, operatorID)
}

// GetOperatorID Get the `operatorUserType` header from context.Context.
func GetOperatorID(ctx context.Context) string {
	operatorID, ok := ctx.Value(headerOperatorIDKey{}).(string)
	if !ok {
		return ""
	}
	return operatorID
}

// SetOperatorName Add `innerLoginUserName` header to context.Context.
func SetOperatorName(ctx context.Context, operatorName string) context.Context {
	return context.WithValue(ctx, headerOperatorNameKey{}, operatorName)
}

// GetOperatorName Get the `innerLoginUserName` header from context.Context.
func GetOperatorName(ctx context.Context) string {
	operatorName, ok := ctx.Value(headerOperatorNameKey{}).(string)
	if !ok {
		return ""
	}
	return operatorName
}

// SetOperatorType Add X-Operator-Type header to context.Context.
func SetOperatorType(ctx context.Context, operatorType string) context.Context {
	return context.WithValue(ctx, headerOperatorTypeKey{}, operatorType)
}

// GetOperatorType Get the X-Operator-Type header from context.Context.
func GetOperatorType(ctx context.Context) string {
	operatorType, ok := ctx.Value(headerOperatorTypeKey{}).(string)
	if !ok {
		return ""
	}
	return operatorType
}

// SetCallerID add `X-Caller-ID` header to context.Context
func SetCallerID(ctx context.Context, callID string) context.Context {
	return context.WithValue(ctx, headerCallerIDKey{}, callID)
}

// GetCallerID get the `X-Caller-ID` header from context.Context
func GetCallerID(ctx context.Context) string {
	if callerID, ok := ctx.Value(headerCallerIDKey{}).(string); ok {
		return callerID
	}
	return ""
}

// SetCallerCode add `X-Caller-Code` header to context.Context
func SetCallerCode(ctx context.Context, callName string) context.Context {
	return context.WithValue(ctx, headerCallerCodeKey{}, callName)
}

// GetCallerCode get the `X-Caller-Code` header from context.Context
func GetCallerCode(ctx context.Context) string {
	if callerCode, ok := ctx.Value(headerCallerCodeKey{}).(string); ok {
		return callerCode
	}
	return ""
}

// SetCallerUser add `X-Caller-User` header to context.Context
func SetCallerUser(ctx context.Context, callName string) context.Context {
	return context.WithValue(ctx, headerCallerUserKey{}, callName)
}

// GetCallerUser get the `X-Caller-User` header from context.Context
func GetCallerUser(ctx context.Context) string {
	if callerUser, ok := ctx.Value(headerCallerUserKey{}).(string); ok {
		return callerUser
	}
	return ""
}

// SetCallerExtra add `X-Caller-Extra` header to context.Context
func SetCallerExtra(ctx context.Context, callExtra string) context.Context {
	return context.WithValue(ctx, headerCallerExtraKey{}, callExtra)
}

// GetCallerExtra get the `X-Caller-Extra` header from context.Context
func GetCallerExtra(ctx context.Context) string {
	if callerExtra, ok := ctx.Value(headerCallerExtraKey{}).(string); ok {
		return callerExtra
	}
	return ""
}

const Environment = "environment"

// ValidateAuth 横向越权校验
func ValidateAuth(ctx context.Context, account string, site ...string) error {
	// 私有云环境，不做横向越权校验
	env := viper.GetString(Environment)
	if env == "private" {
		return nil
	}

	var siteID string
	if len(site) > 0 {
		siteID = site[0]
	}
	if adminID := GetAdminID(ctx); adminID != "" {
		// 云警管理员
		sites := GetAdminSiteSet(ctx)
		if siteID == "" {
			// 跨站点分布式资源，只有全局管理员能操作
			// if !sites.Contains(v.QueryValAll) {
			// 	return e.ErrCodeNotFound
			// }

			// NOTE(huangt): 2023-3-2 与产品会议结果：DCS 将全局资源权限放开，具体能否
			// 操作该资源，由云警控制台自行控制。待后续云警支持“无站点资源”权限后，此处
			// 再做具体处理。
			return nil
		}
		// 站点资源, 全局管理员和站点管理员可以操作
		if !sites.Contains(v.QueryValAll) && !sites.Contains(siteID) {
			return errors.WithStack(code.ErrNotFound)
		}
	} else {
		// 普通用户
		if account != GetAccountID(ctx) {
			return errors.WithStack(code.ErrNotFound)
		}
	}
	return nil
}

// SetEventID Add event id to context.Context.
func SetEventID(ctx context.Context, eventID uint64) context.Context {
	return context.WithValue(ctx, headerEventIDKey{}, eventID)
}

// GetEventID Get the event id from context.Context.
func GetEventID(ctx context.Context) uint64 {
	eventID, ok := ctx.Value(headerEventIDKey{}).(uint64)
	if !ok {
		return eventID
	}
	return eventID
}

// SetStaff Add the staff to context.Context.
func SetStaff(ctx context.Context, staff string) context.Context {
	return context.WithValue(ctx, headerStaffKey{}, staff)
}

// GetStaff Get the staff from context.Context.
func GetStaff(ctx context.Context) string {
	staff, ok := ctx.Value(headerStaffKey{}).(string)
	if !ok {
		return ""
	}
	return staff
}

// SetAccessToken Add the ACCESS_TOKEN to context.Context.
func SetAccessToken(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, HeaderAccessTokenKey{}, userID)
}

// GetAccessToken Get the ACCESS_TOKEN from context.Context.
func GetAccessToken(ctx context.Context) string {
	token, ok := ctx.Value(HeaderAccessTokenKey{}).(string)
	if !ok {
		return ""
	}
	return token
}

// SetAPIFrom Add `X-API-From` to context.Context.
func SetAPIFrom(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, headerAPIFromKey{}, value)
}

// GetAPIFrom get the `X-API-From` header from context.Context.
func GetAPIFrom(ctx context.Context) string {
	value, ok := ctx.Value(headerAPIFromKey{}).(string)
	if !ok {
		return ""
	}
	return value
}

// SetAuthorization Add the Authorization to context.Context.
func SetAuthorization(ctx context.Context, authorization string) context.Context {
	return context.WithValue(ctx, headerAuthorizationKey{}, authorization)
}

// GetAuthorization Get the Authorization from context.Context.
func GetAuthorization(ctx context.Context) string {
	authorization, ok := ctx.Value(headerAuthorizationKey{}).(string)
	if !ok {
		return ""
	}
	return authorization
}

func SetSource(ctx context.Context, source string) context.Context {
	return context.WithValue(ctx, headerSourceKey{}, source)
}

func GetSource(ctx context.Context) string {
	source, ok := ctx.Value(headerSourceKey{}).(string)
	if !ok {
		return ""
	}
	return source
}

// SetIP Add IP to context.Context.
func SetIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, IPKey{}, ip)
}

// GetIP Get the IP from context.Context.
func GetIP(ctx context.Context) string {
	IP, ok := ctx.Value(IPKey{}).(string)
	if !ok {
		return ""
	}
	return IP
}
