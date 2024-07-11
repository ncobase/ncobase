package bootstrap

import (
	"ncobase/common/config"
	"ncobase/common/ecode"
	"ncobase/common/types"
	"ncobase/feature"
	"ncobase/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ginServer creates and initializes the server.
func ginServer(conf *config.Config, fm *feature.Manager) (*gin.Engine, error) { // enforcer *casbin.Enforcer,
	gin.SetMode(conf.RunMode)
	engine := gin.New()

	// Initialize middleware
	middleware.Init(conf)
	engine.Use(middleware.Logger)
	engine.Use(middleware.CORSHandler)
	engine.Use(middleware.ConsumeUser)
	engine.Use(middleware.Trace)

	// // get user tenant service
	// utsi, err := fm.GetService("tenant", "user_tenant")
	// if err != nil {
	// 	return nil, err
	// }
	// userTenantService, _ := utsi.(tenantService.UserTenantServiceInterface)
	// engine.Use(middleware.ConsumeTenant(userTenantService))

	// // get user role service
	// ursi, err := fm.GetService("access", "user_role")
	// if err != nil {
	// 	return nil, err
	// }
	// URSIIMPL, _ := ursi.(accessService.UserRoleServiceInterface)
	//
	// // get role permission service
	// rpsi, err := fm.GetService("access", "role_permission")
	// if err != nil {
	// 	return nil, err
	// }
	// RPSIIMPL, _ := rpsi.(accessService.RolePermissionServiceInterface)

	// // Authorization middleware
	// authzMiddleware := middleware.Authorized(enforcer, conf.Auth.Whitelist, svc)
	// engine.Use(authzMiddleware)

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
