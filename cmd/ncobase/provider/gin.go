package provider

import (
	"context"
	"ncobase/cmd/ncobase/middleware"
	"net/http"
	"time"

	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/ecode"
	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"

	"github.com/gin-gonic/gin"
)

// ginServer creates and initializes server
func ginServer(conf *config.Config, em ext.ManagerInterface) (*gin.Engine, error) {
	// Set gin mode
	if conf.RunMode == "" {
		conf.RunMode = gin.ReleaseMode
	}
	gin.SetMode(conf.RunMode)

	// Create gin engine
	engine := gin.New()

	// Initialize middleware in correct order

	// 1. Basic infrastructure
	engine.Use(middleware.CORSHandler)
	engine.Use(middleware.Trace)
	engine.Use(middleware.ClientInfo)
	engine.Use(middleware.Logger)
	engine.Use(middleware.OtelTrace)

	// 2. Authentication
	engine.Use(middleware.ConsumeUser(em, conf.Auth.Whitelist))

	// 3. Session management
	if err := sessionMiddleware(conf, engine, em); err != nil {
		logger.Warnf(context.Background(), "Failed to setup session middleware: %v", err)
	}

	// 4. Tenant context
	engine.Use(middleware.ConsumeTenant(em, conf.Auth.Whitelist))

	// 5. Authorization
	engine.Use(middleware.CasbinAuthorized(em, conf.Auth.Whitelist))

	// Register routes
	registerRest(engine, conf)
	em.RegisterRoutes(engine)

	// Extension management routes
	if conf.Extension.HotReload {
		g := engine.Group("/sys", middleware.AuthenticatedUser)
		em.ManageRoutes(g)
	}

	// Handle not found routes
	engine.NoRoute(func(c *gin.Context) {
		resp.Fail(c.Writer, resp.NotFound(ecode.Text(http.StatusNotFound)))
	})
	engine.NoMethod()

	return engine, nil
}

// sessionMiddleware sets up session management
func sessionMiddleware(conf *config.Config, engine *gin.Engine, em ext.ManagerInterface) error {
	// Session tracking and validation
	engine.Use(middleware.SessionMiddleware(em))
	engine.Use(middleware.ValidateSessionMiddleware(em))

	// Optional session limits
	if conf.Auth.MaxSessions > 0 {
		engine.Use(middleware.SessionLimitMiddleware(em, conf.Auth.MaxSessions))
	}

	// Start background cleanup task
	cleanupInterval := 1 * time.Hour
	if conf.Auth.SessionCleanupInterval > 0 {
		cleanupInterval = time.Duration(conf.Auth.SessionCleanupInterval) * time.Minute
	}

	go middleware.SessionCleanupTask(context.Background(), em, cleanupInterval)

	logger.Infof(context.Background(), "Session middleware initialized")
	return nil
}
