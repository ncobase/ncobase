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

// ginServer creates and initializes the server.
func ginServer(conf *config.Config, em ext.ManagerInterface) (*gin.Engine, error) {
	// Set gin mode
	if conf.RunMode == "" {
		conf.RunMode = gin.ReleaseMode
	}
	// Set mode before creating engine
	gin.SetMode(conf.RunMode)
	// Create gin engine
	engine := gin.New()

	// Initialize middleware
	engine.Use(middleware.Trace)
	engine.Use(middleware.Logger)
	engine.Use(middleware.CORSHandler)
	engine.Use(middleware.OtelTrace)

	// Validate timestamp
	// engine.Use(middleware.Timestamp(conf.Auth.Whitelist))

	// Consume user
	userMiddleware(conf, engine, em)

	// Consume tenant
	tenantMiddleware(conf, engine, em)

	// Casbin middleware
	if err := casbinMiddleware(conf, engine, em); err != nil {
		return nil, err
	}

	// Register REST
	registerRest(engine, conf)

	// Register GraphQL
	// registerGraphql(engine, svc, conf.RunMode)

	// Register extension / plugin routes
	em.RegisterRoutes(engine)

	// Register extension management routes
	if conf.Extension.HotReload {
		// Belong domain group
		g := engine.Group("/sys", middleware.AuthenticatedUser)
		em.ManageRoutes(g)
	}
	// No route
	engine.NoRoute(func(c *gin.Context) {
		resp.Fail(c.Writer, resp.NotFound(ecode.Text(http.StatusNotFound)))
	})
	engine.NoMethod()

	return engine, nil
}

// register user middleware
func userMiddleware(conf *config.Config, engine *gin.Engine, _ ext.ManagerInterface) {
	engine.Use(middleware.ConsumeUser(jwt.NewTokenManager(conf.Auth.JWT.Secret), conf.Auth.Whitelist))
}

// register Tenant middleware
func tenantMiddleware(conf *config.Config, engine *gin.Engine, em ext.ManagerInterface) {
	// get tenant service
	tsExt, err := em.GetService("tenant")
	if err != nil {
		logger.Errorf(context.Background(), "failed to get tenant service: %v", err.Error())
		return
	}
	// get tenant service
	ts, ok := tsExt.(*tenantService.Service)
	if !ok {
		logger.Errorf(context.Background(), "tenant service does not implement")
		return
	}
	if ts == nil {
		return
	}
	engine.Use(middleware.ConsumeTenant(ts, conf.Auth.Whitelist))
}

// EnforcerProvider is an interface for getting the casbin enforcer
type EnforcerProvider interface {
	GetEnforcer() *casbin.Enforcer
}

// register casbin middleware
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
