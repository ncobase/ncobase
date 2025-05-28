package provider

import (
	"context"
	"fmt"
	"ncobase/cmd/ncobase/middleware"
	tenantService "ncobase/tenant/service"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/ecode"
	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/security/jwt"

	"github.com/gin-gonic/gin"
)

// ginServer creates and initializes the server with properly ordered middleware
func ginServer(conf *config.Config, em ext.ManagerInterface) (*gin.Engine, error) {
	// Set gin mode
	if conf.RunMode == "" {
		conf.RunMode = gin.ReleaseMode
	}
	gin.SetMode(conf.RunMode)

	// Create gin engine
	engine := gin.New()

	// Initialize middleware in correct order

	// 1. Basic infrastructure middleware
	engine.Use(middleware.CORSHandler)

	// 2. Context and tracing setup
	engine.Use(middleware.Trace)

	// 3. Logging
	engine.Use(middleware.Logger)

	// 4. OpenTelemetry tracing
	engine.Use(middleware.OtelTrace)

	// 5. Optional: Timestamp validation
	// engine.Use(middleware.Timestamp(conf.Auth.Whitelist))

	// 6. User authentication and context setup
	userMiddleware(conf, engine, em)

	// 7. Tenant context setup
	tenantMiddleware(conf, engine, em)

	// 8. Authorization (Casbin)
	if err := casbinMiddleware(conf, engine, em); err != nil {
		return nil, err
	}

	// Register routes
	registerRest(engine, conf)

	// Register GraphQL (if needed)
	// registerGraphql(engine, svc, conf.RunMode)

	// Register extension/plugin routes
	em.RegisterRoutes(engine)

	// Register extension management routes
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

// userMiddleware registers user authentication middleware
func userMiddleware(conf *config.Config, engine *gin.Engine, _ ext.ManagerInterface) {
	engine.Use(middleware.ConsumeUser(jwt.NewTokenManager(conf.Auth.JWT.Secret), conf.Auth.Whitelist))
}

// tenantMiddleware registers tenant context middleware
func tenantMiddleware(conf *config.Config, engine *gin.Engine, em ext.ManagerInterface) {
	// Get tenant service
	tsExt, err := em.GetService("tenant")
	if err != nil {
		logger.Errorf(context.Background(), "failed to get tenant service: %v", err.Error())
		return
	}

	// Cast to tenant service
	ts, ok := tsExt.(*tenantService.Service)
	if !ok {
		logger.Errorf(context.Background(), "tenant service does not implement expected interface")
		return
	}
	if ts == nil {
		return
	}

	engine.Use(middleware.ConsumeTenant(ts, conf.Auth.Whitelist))
}

// EnforcerProvider interface for getting casbin enforcer
type EnforcerProvider interface {
	GetEnforcer() *casbin.Enforcer
}

// casbinMiddleware registers casbin authorization middleware
func casbinMiddleware(conf *config.Config, engine *gin.Engine, em ext.ManagerInterface) error {
	asExt, err := em.GetExtension("access")
	if err != nil {
		return fmt.Errorf("failed to get access module: %v", err)
	}

	if ep, ok := asExt.(EnforcerProvider); ok {
		if enforcer := ep.GetEnforcer(); enforcer != nil {
			engine.Use(middleware.CasbinAuthorized(em, enforcer, conf.Auth.Whitelist))
			return nil
		}
		logger.Warnf(context.Background(), "casbin enforcer is nil, skipping casbin middleware")
		return nil
	}

	return fmt.Errorf("access module does not provide enforcer")
}
