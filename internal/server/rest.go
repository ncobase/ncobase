package server

import (
	"stocms/internal/handler"
	"stocms/internal/server/middleware"

	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

func registerRestRouter(e *gin.Engine, h *handler.Handler) {
	// Handler prefix
	v1 := e.Group("/")
	// Authorize
	v1.GET("/authorize", middleware.Authorized, nil)
	// Hello
	v1.GET("/sample/hello", h.Hello)
	// Swagger
	v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
