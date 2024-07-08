package quota

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"template/pkg/resp"
)

func RegisterAPI(router *gin.Engine) {
	router.POST("v1/quotas/sync", SyncUsed)
}

type SyncUsedReq struct {
	Params []*SyncParam `json:"params"`
}

type SyncParam struct {
	AssociatedID string `json:"associated_id" binding:"required"`
	Name         string `json:"name"`
}

func SyncUsed(c *gin.Context) {
	var req SyncUsedReq
	if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
		resp.ErrorParam(c, err)
		return
	}
	params := make([]*Param, 0, len(req.Params))
	for _, v := range req.Params {
		params = append(params, &Param{
			AssociatedID: v.AssociatedID,
			Name:         v.Name,
		})
	}
	finisher, err := Mgr().getFinisher(c.Request.Context(), params)
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
