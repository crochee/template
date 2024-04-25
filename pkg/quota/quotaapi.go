package quota

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"template/pkg/resp"
)

func RegisterAPI(router *gin.Engine) {
	router.POST("/used_quotas", SyncUsed)

}

type SyncUsedReq struct {
	Params []*Param
}

func SyncUsed(c *gin.Context) {
	var req SyncUsedReq
	if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
		resp.ErrorParam(c, err)
		return
	}
	finisher, err := Mgr().getFinisher(c.Request.Context(), req.Params)
	if err != nil {
		resp.Error(c, err)
		return
	}

	if err := finisher.sync(c.Request.Context()); err != nil {
		resp.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
