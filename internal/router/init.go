package router

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"template/internal/controllers"
	"template/internal/gateway"
	"template/internal/service"
	"template/internal/store"
	"template/pkg/logger"
	"template/pkg/middlewares"
)

// New gin router
func New(store store.Store, client gateway.Client) *gin.Engine {
	// init
	gin.DefaultWriter = logger.SetWriter(viper.GetBool("log.console"), "")

	router := gin.New()
	router.GET("/health", controllers.Health)
	// add middlewares
	router.Use(
		middlewares.Log,
		middlewares.Recovery,
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
