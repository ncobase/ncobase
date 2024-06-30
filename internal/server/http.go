package server

import (
	"ncobase/common/config"
	"ncobase/common/ecode"
	"ncobase/common/types"
	"ncobase/internal/handler"
	"ncobase/internal/server/middleware"
	"ncobase/internal/service"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

// newHTTP creates an HTTP server.
func newHTTP(conf *config.Config, h *handler.Handler, svc *service.Service, enforcer *casbin.Enforcer, pm *PluginManager) (*gin.Engine, error) {
	gin.SetMode(conf.RunMode)
	engine := gin.New()

	// Initialize middleware
	initializeMiddleware(engine, conf, svc, enforcer)

	// Register REST
	registerRest(engine, h, conf)

	// Register GraphQL
	registerGraphql(engine, svc, conf.RunMode)

	// Register plugin routes
	pm.RegisterPluginRoutes(engine)

	// Register plugin management routes
	if conf.Plugin.HotReload {
		pm.AddPluginRoutes(engine)
	}

	engine.NoRoute(notFound)
	engine.NoMethod()

	return engine, nil
}

// notFound returns a 404 JSON response.
func notFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, types.JSON{"message": ecode.Text(http.StatusNotFound)})
}

// initializeMiddleware initializes the middleware.
func initializeMiddleware(engine *gin.Engine, conf *config.Config, svc *service.Service, enforcer *casbin.Enforcer) {
	middleware.Init(conf)
	engine.Use(middleware.Logger)
	engine.Use(middleware.CORSHandler)
	engine.Use(middleware.ConsumeUser)
	engine.Use(middleware.BindConfig)
	engine.Use(middleware.BindGinContext)
	engine.Use(middleware.Trace)
	engine.Use(middleware.ConsumeTenant(svc))

	// Authorization middleware
	authzMiddleware := middleware.Authorized(enforcer, conf.Auth.Whitelist, svc)
	engine.Use(authzMiddleware)
}
