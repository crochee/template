package middlewares

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func Tracing(service string, getAttrs func(context.Context) []attribute.KeyValue) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, span := otel.Tracer(service).Start(c.Request.Context(),
			"Tracing",
			trace.WithAttributes(getAttrs(c.Request.Context())...))
		defer span.End()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
