package bootstrap

import (
	"ncobase/core/handler"
	"ncobase/helper"
	"ncobase/middleware"
	"net/http"

	"ncobase/common/config"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// registerRest registers the REST routes.
func registerRest(e *gin.Engine, h *handler.Handler, conf *config.Config) {
	// Root endpoint, redirect when domain is configured and not localhost
	e.GET("/", func(c *gin.Context) {
		if domain := conf.Domain; domain != "localhost" {
			url := helper.GetHost(conf, domain)
			c.Redirect(http.StatusMovedPermanently, url)
			return
		}
		c.String(http.StatusOK, "It's working.")
	})

	// Health check endpoint
	e.GET("/health", h.HealthHandler)

	// API v1 endpoints
	v1 := e.Group("/v1")

	// Authentication endpoints
	v1.POST("/login", h.LoginHandler)
	v1.POST("/register", h.RegisterHandler)
	v1.POST("/logout", h.LogoutHandler)

	// Captcha endpoints
	captcha := v1.Group("/captcha")
	{
		captcha.GET("/generate", h.GenerateCaptchaHandler)
		captcha.GET("/:captcha", h.CaptchaStreamHandler)
		captcha.POST("/validate", h.ValidateCaptchaHandler)
	}

	// Authorization endpoints
	authorize := v1.Group("/authorize")
	{
		authorize.POST("/send", h.SendCodeHandler)
		authorize.GET("/:code", h.CodeAuthHandler)
	}

	// Account endpoints
	account := v1.Group("/account", middleware.Authenticated)
	{
		account.GET("", h.GetMeHandler)
		account.PUT("/password", h.UpdatePasswordHandler)
		account.GET("/tenant", h.AccountTenantHandler)
		account.GET("/tenants", h.AccountTenantsHandler)
	}

	// User endpoints
	users := v1.Group("/users")
	{
		// users.GET("", h.ListUserHandler)
		// users.POST("", h.CreateUserHandler)
		users.GET("/:username", h.GetUserHandler)
		// users.PUT("/:username", h.UpdateUserHandler)
		// users.DELETE("/:username", h.DeleteUserHandler)
		// users.GET("/:username/roles", h.ListUserRoleHandler)
		// users.GET("/:username/groups", h.ListUserGroupHandler)
		// users.GET("/:username/tenants", h.UserTenantHandler)
		// users.GET("/:username/tenants/:slug", h.UserTenantHandler)
		// users.GET("/:username/tenant/belongs", middleware.Authenticated, h.ListUserBelongHandler)
	}

	// Module endpoints
	modules := v1.Group("/modules", middleware.Authenticated)
	{
		modules.GET("", h.ListModuleHandler)
		modules.POST("", h.CreateModuleHandler)
		modules.GET("/:slug", h.GetModuleHandler)
		modules.PUT("/:slug", h.UpdateModuleHandler)
		modules.DELETE("/:slug", h.DeleteModuleHandler)
	}

	// Tenant endpoints
	tenants := v1.Group("/tenants", middleware.Authenticated)
	{
		tenants.GET("", h.ListTenantHandler)
		tenants.POST("", h.CreateTenantHandler)
		tenants.GET("/:slug", h.GetTenantHandler)
		tenants.PUT("/:slug", h.UpdateTenantHandler)
		tenants.DELETE("/:slug", h.DeleteTenantHandler)

		// // Tenant asset endpoints
		// tenants.GET("/:tenant/assets", h.ListTenantAssetHandler)
		// tenants.POST("/:tenant/assets", h.CreateTenantAssetsHandler)
		// tenants.GET("/:tenant/assets/:asset", h.GetTenantAssetHandler)
		// tenants.PUT("/:tenant/assets/:asset", h.UpdateTenantAssetHandler)
		// tenants.DELETE("/:tenant/assets/:asset", h.DeleteTenantAssetHandler)
		//
		// // // Tenant role endpoints
		// // tenants.GET("/:tenant/roles", h.ListTenantRoleHandler)
		// // tenants.POST("/:tenant/roles", h.CreateTenantRoleHandler)
		// // tenants.GET("/:tenant/roles/:role", h.GetTenantRoleHandler)
		// // tenants.PUT("/:tenant/roles/:role", h.UpdateTenantRoleHandler)
		// // tenants.DELETE("/:tenant/roles/:role", h.DeleteTenantRoleHandler)
		// // tenants.GET("/:tenant/roles/:roleSlug/permissions", h.ListTenantRolePermissionHandler)
		// // tenants.GET("/:tenant/roles/:roleSlug/users", h.ListTenantRoleUserHandler)
		// //
		// // // Tenant permission endpoints
		// // tenants.GET("/:tenant/permissions", h.ListTenantPermissionHandler)
		// // tenants.POST("/:tenant/permissions", h.CreateTenantPermissionHandler)
		// // tenants.GET("/:tenant/permissions/:permission", h.GetTenantPermissionHandler)
		// // tenants.PUT("/:tenant/permissions/:permission", h.UpdateTenantPermissionHandler)
		// // tenants.DELETE("/:tenant/permissions/:permission", h.DeleteTenantPermissionHandler)
		// //
		// // // Tenant module endpoints
		// // tenants.GET("/:tenant/modules", h.ListTenantModuleHandler)
		// // tenants.POST("/:tenant/modules", h.CreateTenantModuleHandler)
		// // tenants.GET("/:tenant/modules/:module", h.GetTenantModuleHandler)
		// // tenants.PUT("/:tenant/modules/:module", h.UpdateTenantModuleHandler)
		// // tenants.DELETE("/:tenant/modules/:module", h.DeleteTenantModuleHandler)
		// //
		// // Tenant menu endpoints
		// tenants.GET("/:tenant/menus", h.ListTenantMenusHandler)
		// tenants.POST("/:tenant/menus", h.CreateTenantMenuHandler)
		// tenants.GET("/:tenant/menus/:menu", h.GetTenantMenuHandler)
		// tenants.PUT("/:tenant/menus/:menu", h.UpdateTenantMenuHandler)
		// tenants.DELETE("/:tenant/menus/:menu", h.DeleteTenantMenuHandler)
		// //
		// // // Tenant policy endpoints
		// // tenants.GET("/:tenant/policies", h.ListTenantPolicyHandler)
		// // tenants.POST("/:tenant/policies", h.CreateTenantPolicyHandler)
		// // tenants.GET("/:tenant/policies/:policyId", h.GetTenantPolicyHandler)
		// // tenants.PUT("/:tenant/policies/:policyId", h.UpdateTenantPolicyHandler)
		// // tenants.DELETE("/:tenant/policies/:policyId", h.DeleteTenantPolicyHandler)
		// //
		// // // Tenant taxonomy endpoints
		// // tenants.GET("/:tenant/taxonomies", h.ListTenantTaxonomyHandler)
		// // tenants.POST("/:tenant/taxonomies", h.CreateTenantTaxonomyHandler)
		// // tenants.GET("/:tenant/taxonomies/:taxonomy", h.GetTenantTaxonomyHandler)
		// // tenants.PUT("/:tenant/taxonomies/:taxonomy", h.UpdateTenantTaxonomyHandler)
		// // tenants.DELETE("/:tenant/taxonomies/:taxonomy", h.DeleteTenantTaxonomyHandler)
		// //
		// // // Tenant topic endpoints
		// // tenants.GET("/:tenant/topics", h.ListTenantTopicHandler)
		// // tenants.POST("/:tenant/topics", h.CreateTenantTopicHandler)
		// // tenants.GET("/:tenant/topics/:topic", h.GetTenantTopicHandler)
		// // tenants.PUT("/:tenant/topics/:topic", h.UpdateTenantTopicHandler)
		// // tenants.DELETE("/:tenant/topics/:topic", h.DeleteTenantTopicHandler)
	}

	// Group endpoints
	// groups := v1.Group("/groups", middleware.Authenticated)
	// {
	// 	groups.GET("", h.ListGroupHandler)
	// 	groups.POST("", h.CreateGroupHandler)
	// 	groups.GET("/:slug", h.GetGroupHandler)
	// 	groups.PUT("/:slug", h.UpdateGroupHandler)
	// 	groups.DELETE("/:slug", h.DeleteGroupHandler)
	// 	groups.GET("/:slug/roles", h.ListGroupRoleHandler)
	// 	groups.GET("/:slug/users", h.ListGroupUserHandler)
	// }

	// Role endpoints
	roles := v1.Group("/roles", middleware.Authenticated)
	{
		roles.GET("", h.ListRoleHandler)
		roles.POST("", h.CreateRoleHandler)
		roles.GET("/:slug", h.GetRoleHandler)
		roles.PUT("/:slug", h.UpdateRoleHandler)
		roles.DELETE("/:slug", h.DeleteRoleHandler)
		roles.GET("/:slug/permissions", h.ListRolePermissionHandler)
		roles.GET("/:slug/users", h.ListUserRoleHandler)
	}

	// Permission endpoints
	permissions := v1.Group("/permissions", middleware.Authenticated)
	{
		permissions.GET("", h.ListPermissionHandler)
		permissions.POST("", h.CreatePermissionHandler)
		permissions.GET("/:slug", h.GetPermissionHandler)
		permissions.PUT("/:slug", h.UpdatePermissionHandler)
		permissions.DELETE("/:slug", h.DeletePermissionHandler)
	}

	// Casbin Rule endpoints
	policies := v1.Group("/policies", middleware.Authenticated)
	{
		policies.GET("", h.ListCasbinRuleHandler)
		policies.POST("", h.CreateCasbinRuleHandler)
		policies.GET("/:id", h.GetCasbinRuleHandler)
		policies.PUT("/:id", h.UpdateCasbinRuleHandler)
		policies.DELETE("/:id", h.DeleteCasbinRuleHandler)
	}

	// Swagger documentation endpoint
	if conf.RunMode != gin.ReleaseMode {
		e.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}
