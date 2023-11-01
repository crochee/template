package middlewares

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"template/pkg/msg"
	"template/pkg/utils/v"
)

func Tracing(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var attrs []attribute.KeyValue
		if accountID := c.GetHeader(v.HeaderAccountID); accountID != "" {
			attrs = append(attrs, msg.AccountIDKey.String(accountID))
		}
		if userID := c.GetHeader(v.HeaderUserID); userID != "" {
			attrs = append(attrs, msg.UserIDKey.String(userID))
		}
		var (
			span trace.Span
			ctx  context.Context
		)
		if len(attrs) > 0 {
			ctx, span = otel.Tracer(service).Start(c.Request.Context(),
				"Tracing", trace.WithAttributes(attrs...))
		} else {
			ctx, span = otel.Tracer(service).Start(c.Request.Context(), "Tracing")
		}
		defer span.End()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
