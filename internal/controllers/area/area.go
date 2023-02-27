package area

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"template/internal/request"
	"template/internal/service"
	"template/pkg/resp"
)

type AreaController struct {
	srv service.Service
}

func NewAreaController(srv service.Service) *AreaController {
	return &AreaController{
		srv: srv,
	}
}

// List 获取区域列表
func (a *AreaController) List(c *gin.Context) {
	ctx := c.Request.Context()
	var req request.QueryAreaListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.ErrorParam(c, err)
		return
	}

	result, err := a.srv.Area().List(ctx, &req)
	if err != nil {
		resp.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
