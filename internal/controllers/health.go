package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health 健康检查,用于k8s pod的心跳检查
func Health(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
