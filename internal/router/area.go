package router

import (
	"github.com/gin-gonic/gin"

	"template/internal/controllers/area"
	"template/internal/middleware"
	"template/internal/service"
)

func registerArea(router *gin.RouterGroup, srv service.Service) {
	userGroup := router.Group("/areas")
	{
		areaController := area.NewAreaController(srv)
		// public api
		userGroup.Use(middleware.CheckHeaders)
		userGroup.GET("/", areaController.List)
	}
}
