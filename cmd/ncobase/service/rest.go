package service

import (
	"ncobase/common/helper"
	"net/http"

	"ncobase/common/config"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// registerRest registers the REST routes.
func registerRest(e *gin.Engine, conf *config.Config) {
	// Root endpoint, redirect when domain is configured and not localhost
	e.GET("/", func(c *gin.Context) {
		if domain := conf.Domain; domain != "localhost" {
			url := helper.GetHost(conf, domain)
			c.Redirect(http.StatusMovedPermanently, url)
			return
		}
		c.String(http.StatusOK, "It's working.")
	})

	// Swagger documentation endpoint
	if conf.RunMode != gin.ReleaseMode {
		e.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}
