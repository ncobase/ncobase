package service

import (
	"context"
	"ncobase/cmd/ncobase/middleware"
	"ncobase/common/config"
	"ncobase/common/ecode"
	"ncobase/common/feature"
	"ncobase/common/log"
	"ncobase/common/resp"
	accessService "ncobase/feature/access/service"
	tenantService "ncobase/feature/tenant/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ginServer creates and initializes the server.
func ginServer(conf *config.Config, fm *feature.Manager) (*gin.Engine, error) {
	gin.SetMode(conf.RunMode)
	engine := gin.New()

	// Initialize middleware
	engine.Use(middleware.Timestamp)
	engine.Use(middleware.Trace)
	engine.Use(middleware.Logger)
	engine.Use(middleware.CORSHandler)
	engine.Use(middleware.OtelTrace)

	// Consume user
	userMiddleware(conf, engine, fm)

	// Consume tenant
	tenantMiddleware(conf, engine, fm)

	// Casbin middleware
	casbinMiddleware(conf, engine, fm)

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

	engine.NoRoute(func(c *gin.Context) {
		resp.Fail(c.Writer, resp.NotFound(ecode.Text(http.StatusNotFound)))
	})
	engine.NoMethod()

	return engine, nil
}

// register user middleware
func userMiddleware(conf *config.Config, engine *gin.Engine, _ *feature.Manager) {
	// get user service
	// fu, err := fm.GetService("user")
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
func tenantMiddleware(conf *config.Config, engine *gin.Engine, fm *feature.Manager) {
	// get tenant service
	ft, err := fm.GetService("tenant")
	if err != nil {
		log.Errorf(context.Background(), "failed to get tenant service: %v", err.Error())
		return
	}

	// get tenant service
	ts, ok := ft.(*tenantService.Service)
	if !ok {
		log.Errorf(context.Background(), "tenant service does not implement")
		return
	}
	if ts == nil {
		return
	}
	engine.Use(middleware.ConsumeTenant(ts, conf.Auth.Whitelist))
}

// register casbin middleware
func casbinMiddleware(conf *config.Config, engine *gin.Engine, fm *feature.Manager) {
	// get access service
	fa, err := fm.GetService("access")
	if err != nil {
		log.Errorf(context.Background(), "failed to get access service: %v", err.Error())
		return
	}
	// get access service
	as, ok := fa.(*accessService.Service)
	if !ok {
		log.Errorf(context.Background(), "access service does not implement")
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
