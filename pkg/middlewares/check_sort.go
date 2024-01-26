package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	e "template/pkg/code"
	log "template/pkg/logger"
	"template/pkg/resp"
	"template/pkg/validator"
)

// CheckQueryFieldSort check field from c.query is valid or not
func CheckQueryFieldSort(c *gin.Context) {
	ctx := c.Request.Context()
	sort := c.Query("sort")
	if sort != "" {
		if err := validator.Var(binding.Validator, sort, "order"); err != nil {
			log.FromContext(ctx).Err(err).Msg("fail to validator sort field")
			ec := e.ErrCodeInvalidParam.WithResult(err.Error())
			resp.Error(c, ec)
			return
		}
	}
	// before request
	c.Next()
	// after request
}
