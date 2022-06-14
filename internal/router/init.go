package router

import (
	"github.com/gin-gonic/gin"

	"go_template/pkg/middleware"
)

// New gin router
func New() *gin.Engine {
	// init
	router := gin.New()

	// add middleware
	router.Use(
		middleware.RequestLogger,
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
