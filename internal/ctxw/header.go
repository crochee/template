package ctxw

import (
	"context"

	"go_template/pkg/logger"
)

type (
	headerAccountIDKey    struct{}
	headerUserIDKey       struct{}
	headerTraceIDKey      struct{}
	headerCacheControlKey struct{}
	headerCallFromKey     struct{}
)

func NewContext(ctx context.Context) context.Context {
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
	if cacheControl := GetCacheControl(ctx); cacheControl != "" {
		newCtx = SetCacheControl(newCtx, cacheControl)
	}
	if callFrom := GetCallFrom(ctx); callFrom != "" {
		newCtx = SetCallFrom(newCtx, callFrom)
	}
	return logger.With(newCtx, logger.From(ctx))
}

func SetAccountAndUserID(ctx context.Context, accountID, userID string) context.Context {
	ctx = SetUserID(SetAccountID(ctx, accountID), userID)
	return logger.With(ctx, logger.From(ctx))
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
