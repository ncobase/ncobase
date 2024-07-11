package bootstrap

import (
	"ncobase/cmd/bootstrap/middleware"
	"ncobase/common/config"
	"ncobase/common/ecode"
	"ncobase/common/types"
	"ncobase/feature"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ginServer creates and initializes the server.
func ginServer(conf *config.Config, fm *feature.Manager) (*gin.Engine, error) {
	gin.SetMode(conf.RunMode)
	engine := gin.New()

	// Initialize middleware
	engine.Use(middleware.Logger)
	engine.Use(middleware.CORSHandler)
	engine.Use(middleware.Trace)

	// Register REST
	registerRest(engine, conf)

	// Register GraphQL
	// registerGraphql(engine, svc, conf.RunMode)

	// Register feature / plugin routes
	fm.RegisterRoutes(engine)

	// Register feature management routes
	if conf.Feature.HotReload {
		fm.ManageRoutes(engine)
	}

	// Register WebSocket route
	registerWebSocketRoutes(engine)

	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, types.JSON{"message": ecode.Text(http.StatusNotFound)})
	})
	engine.NoMethod()

	return engine, nil
}
