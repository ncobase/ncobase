package server

import (
	"net/http"
	"stocms/internal/config"
	"stocms/internal/handler"
	"stocms/internal/helper"
	"stocms/internal/server/middleware"
	"stocms/internal/service"
	"stocms/pkg/ecode"
	"stocms/pkg/types"

	"github.com/gin-gonic/gin"
)

// newHTTP creates an HTTP server.
func newHTTP(conf *config.Config, h *handler.Handler, svc *service.Service) (*gin.Engine, error) {

	gin.SetMode(conf.RunMode)
	engine := gin.New()

	// Middleware
	middleware.Init(conf)
	engine.Use(middleware.Logger())
	engine.Use(middleware.CORSHandler())
	engine.Use(middleware.ConsumeUser())

	// Register REST
	registerRest(engine, h, conf)

	// Register GraphQL
	registerGraphql(engine, svc, conf.RunMode)

	// Register config to Context
	engine.Use(func(c *gin.Context) {
		// Set the config to the gin.Context
		helper.SetConfig(c, conf)
		// Proceed to the next middleware or handler
		c.Next()
	})

	engine.NoRoute(notFound)
	engine.NoMethod()

	return engine, nil
}

func notFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, types.JSON{"message": ecode.Text(http.StatusNotFound)})
}
