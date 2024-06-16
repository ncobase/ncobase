package server

import (
	"ncobase/internal/config"
	"ncobase/internal/handler"
	"ncobase/internal/server/middleware"
	"ncobase/internal/service"
	"ncobase/pkg/ecode"
	"ncobase/pkg/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

// newHTTP creates an HTTP server.
func newHTTP(conf *config.Config, h *handler.Handler, svc *service.Service) (*gin.Engine, error) {

	gin.SetMode(conf.RunMode)
	engine := gin.New()

	// Middleware
	middleware.Init(conf)
	engine.Use(middleware.Logger)
	engine.Use(middleware.CORSHandler)
	engine.Use(middleware.ConsumeUser)
	engine.Use(middleware.BindConfig)
	engine.Use(middleware.BindGinContext)
	engine.Use(middleware.Trace)
	engine.Use(middleware.ConsumeTenant(svc))

	// Register REST
	registerRest(engine, h, conf)

	// Register GraphQL
	registerGraphql(engine, svc, conf.RunMode)

	engine.NoRoute(notFound)
	engine.NoMethod()

	return engine, nil
}

func notFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, types.JSON{"message": ecode.Text(http.StatusNotFound)})
}
