package middlewares

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"template/pkg/resp"
)

// SetHeaderContext set some public parameters to header and context
func SetHeaderContext(setCtx func(context.Context,
	http.Header) (context.Context, error)) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx, err := setCtx(c.Request.Context(), c.Request.Header)
		if err != nil {
			resp.Error(c, err)
			return
		}
		// reset
		c.Request = c.Request.WithContext(ctx)

		// before request
		c.Next()
		// after request
	}
}
