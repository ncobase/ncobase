package bootstrap

import (
	"ncobase/common/config"
	"ncobase/common/ecode"
	"ncobase/common/types"
	"ncobase/core/handler"
	"ncobase/core/service"
	"ncobase/middleware"
	"ncobase/plugin"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

// ginServer creates and initializes the server.
func ginServer(conf *config.Config, h *handler.Handler, svc *service.Service, enforcer *casbin.Enforcer, pm *plugin.Manager) (*gin.Engine, error) {
	gin.SetMode(conf.RunMode)
	engine := gin.New()

	// Initialize middleware
	middleware.Init(conf)
	engine.Use(middleware.Logger)
	engine.Use(middleware.CORSHandler)
	engine.Use(middleware.ConsumeUser)
	// engine.Use(middleware.BindConfig)
	// engine.Use(middleware.BindGinContext)
	engine.Use(middleware.Trace)
	engine.Use(middleware.ConsumeTenant(svc))

	// Authorization middleware
	authzMiddleware := middleware.Authorized(enforcer, conf.Auth.Whitelist, svc)
	engine.Use(authzMiddleware)

	// Register WebSocket route
	registerWebSocketRoutes(engine)

	// Register REST
	registerRest(engine, h, conf)

	// Register GraphQL
	registerGraphql(engine, svc, conf.RunMode)

	// Register plugin routes
	pm.RegisterPluginRoutes(engine)

	// Register plugin management routes
	if conf.Plugin.HotReload {
		pm.ManageRoutes(engine)
	}

	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, types.JSON{"message": ecode.Text(http.StatusNotFound)})
	})
	engine.NoMethod()

	return engine, nil
}
