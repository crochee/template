package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
)

// CrossDomain skip the cross-domain phase
func CrossDomain() gin.HandlerFunc {
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{"*"},
		MaxAge:         12 * 60 * 60,
	})
	return func(ctx *gin.Context) {
		c.HandlerFunc(ctx.Writer, ctx.Request)
		if ctx.Request.Method == http.MethodOptions &&
			ctx.GetHeader("Access-Control-Request-Method") != "" {
			// Abort processing next Gin middlewares.
			ctx.AbortWithStatus(http.StatusNoContent)
		}
	}
}
