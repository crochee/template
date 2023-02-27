package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"template/pkg/resp"
)

func RegisterAPI(router *gin.Engine) {
	router.GET("/v1/produce/config", getProduceConfig)
	router.PATCH("/v1/produce/config", updateProduceConfig)
}

//swagger:route GET /v1/produce/config 故障定位服务-内部运维使用 SwagNullRequest
//
// 查询生产者配置.
//
// This will get producer config info.
//
//     Responses:
//       200: SGetProduceConfigRes
//       default: ResponseCode
func getProduceConfig(c *gin.Context) {
	ctx := c.Request.Context()
	result, err := GetProduce(ctx)
	if err != nil {
		Error(ctx, err)
		resp.ErrorParam(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

//swagger:route PATCH /v1/produce/config 故障定位服务-内部运维使用 SUpdateProduceConfigRequest
//
// 更新生产者配置.
//
// This will update producer config info.
//
//     Responses:
//       200: ResponseCode
//       default: ResponseCode
func updateProduceConfig(c *gin.Context) {
	ctx := c.Request.Context()
	param := ProduceConfig{}
	if err := c.BindJSON(&param); err != nil {
		Error(ctx, err)
		resp.ErrorParam(c, err)
		return
	}
	if err := UpdateProduce(ctx, &param); err != nil {
		Error(ctx, err)
		resp.ErrorParam(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
