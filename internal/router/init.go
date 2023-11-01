package router

import (
	"github.com/gin-gonic/gin"

	"template/internal/controllers"
	"template/internal/gateway"
	"template/internal/service"
	"template/internal/store"
	"template/pkg/logger/gormx"
	"template/pkg/middlewares"
)

// New gin router
func New(store store.Store, client gateway.Client) *gin.Engine {
	router := gin.New()
	router.GET("/health", controllers.Health)
	// add middlewares
	router.Use(
		middlewares.AccessLog(gormx.NewZapGormWriterFrom),
		middlewares.Recovery(gormx.NewZapGormWriterFrom),
	)
	srv := service.NewService(store, client)
	v1RouterGroup(router, srv)

	return router
}

func v1RouterGroup(router *gin.Engine, srv service.Service) {
	v1Router := router.Group("/v1")
	{
		registerArea(v1Router, srv)
	}
}
