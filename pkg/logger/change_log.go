package logger

import (
	"net/http"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func RegisterLog(router *gin.RouterGroup) {
	router.GET("/log", getLog)
	router.PUT("/log", updateLog)
}

var debug uint32

type LoggerContent struct {
	Debug bool `json:"debug" binding:"required"`
}

func getLog(c *gin.Context) {
	c.JSON(http.StatusOK, LoggerContent{
		Debug: atomic.LoadUint32(&debug) == 1,
	})
}

func updateLog(c *gin.Context) {
	var req LoggerContent
	if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, map[string]interface{}{
			"code":    "COMMON.4000000002",
			"message": err.Error(),
			"result":  nil,
		})
		return
	}
	if req.Debug {
		atomic.StoreUint32(&debug, 1)
	} else {
		atomic.StoreUint32(&debug, 0)
	}
	c.Status(http.StatusNoContent)
}
