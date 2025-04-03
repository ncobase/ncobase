package provider

import (
	"context"
	"ncobase/cmd/ncobase/middleware"
	accessService "ncobase/core/access/service"
	tenantService "ncobase/core/tenant/service"
	"ncobase/ncore/config"
	"ncobase/ncore/ecode"
	"ncobase/ncore/extension"
	"ncobase/ncore/logger"
	"ncobase/ncore/resp"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ginServer creates and initializes the server.
func ginServer(conf *config.Config, em *extension.Manager) (*gin.Engine, error) {
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
	casbinMiddleware(conf, engine, em)

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
func userMiddleware(conf *config.Config, engine *gin.Engine, _ *extension.Manager) {
	// get user service
	// fu, err := em.GetService("user")
	// if err != nil {
	// 	log.Errorf(context.Background(), "failed to get user service: %v", err.Error())
	// 	return
	// }
	// // get user service
	// us, ok := fu.(*userService.Service)
	// if !ok {
	// 	log.Errorf(context.Background(), "user service does not implement")
	// 	return
	// }
	// if us == nil {
	// 	return
	// }
	engine.Use(middleware.ConsumeUser(conf.Auth.JWT.Secret, conf.Auth.Whitelist))
}

// register Tenant middleware
func tenantMiddleware(conf *config.Config, engine *gin.Engine, em *extension.Manager) {
	// get tenant service
	ft, err := em.GetService("tenant")
	if err != nil {
		logger.Errorf(context.Background(), "failed to get tenant service: %v", err.Error())
		return
	}

	// get tenant service
	ts, ok := ft.(*tenantService.Service)
	if !ok {
		logger.Errorf(context.Background(), "tenant service does not implement")
		return
	}
	if ts == nil {
		return
	}
	engine.Use(middleware.ConsumeTenant(ts, conf.Auth.Whitelist))
}

// register casbin middleware
func casbinMiddleware(conf *config.Config, engine *gin.Engine, em *extension.Manager) {
	// get access service
	fa, err := em.GetService("access")
	if err != nil {
		logger.Errorf(context.Background(), "failed to get access service: %v", err.Error())
		return
	}
	// get access service
	as, ok := fa.(*accessService.Service)
	if !ok {
		logger.Errorf(context.Background(), "access service does not implement")
		return
	}
	if as == nil {
		return
	}

	enforcer, err := as.CasbinAdapter.InitEnforcer()
	if err != nil {
		panic(err)
	}
	engine.Use(middleware.CasbinAuthorized(enforcer, conf.Auth.Whitelist, as))
}
