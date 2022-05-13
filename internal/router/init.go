package router

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/crochee/devt/internal/util/v"
	"github.com/crochee/devt/pkg/logger"
	"github.com/crochee/devt/pkg/middleware"
)

// New gin router
func New() *gin.Engine {
	// init
	router := gin.New()

	// add middleware
	router.Use(
		middleware.RequestLogger(
			logger.New(
				logger.WithFields(zap.String("service", v.ServiceName)),
				logger.WithLevel(viper.GetString("level")),
				logger.WithWriter(logger.SetWriter(viper.GetString("path"))))),
		middleware.Log,
		middleware.Recovery,
	)

	v1RouterGroup(router)

	return router
}

func v1RouterGroup(router *gin.Engine) {
	v1Router := router.Group("/v1")
	v1Router.GET("/demod")
}
