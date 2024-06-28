package server

import (
	"ncobase/internal/handler"
	"ncobase/internal/helper"
	"ncobase/internal/server/middleware"
	"net/http"

	"ncobase/common/config"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

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

	// Asset endpoints
	assets := v1.Group("/assets", middleware.Authenticated)
	{
		assets.GET("", h.ListAssetHandler)
		assets.POST("", h.CreateAssetsHandler)
		assets.GET("/:slug", h.GetAssetHandler)
		assets.PUT("/:slug", h.UpdateAssetHandler)
		assets.DELETE("/:slug", h.DeleteAssetHandler)
	}

	// Tenant endpoints
	tenants := v1.Group("/tenants", middleware.Authenticated)
	{
		tenants.GET("", h.ListTenantHandler)
		tenants.POST("", h.CreateTenantHandler)
		tenants.GET("/:slug", h.GetTenantHandler)
		tenants.PUT("/:slug", h.UpdateTenantHandler)
		tenants.DELETE("/:slug", h.DeleteTenantHandler)
		tenants.GET("/:slug/assets", h.ListTenantAssetHandler)
		tenants.GET("/:slug/roles", h.ListTenantRoleHandler)
		tenants.GET("/:slug/modules", h.ListTenantModuleHandler)
		tenants.GET("/:slug/settings", h.ListTenantSettingHandler)
		tenants.GET("/:slug/users", h.ListTenantUserHandler)
		tenants.GET("/:slug/groups", h.ListTenantGroupHandler)
	}

	// Menu endpoints
	menus := v1.Group("/menus", middleware.Authenticated)
	{
		menus.GET("", h.ListMenusHandler)
		menus.POST("", h.CreateMenuHandler)
		menus.GET("/:slug", h.GetMenuHandler)
		menus.PUT("/:slug", h.UpdateMenuHandler)
		menus.DELETE("/:slug", h.DeleteMenuHandler)
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

	// Taxonomy endpoints
	taxonomies := v1.Group("/taxonomies", middleware.Authenticated)
	{
		taxonomies.GET("", h.ListTaxonomyHandler)
		taxonomies.POST("", h.CreateTaxonomyHandler)
		taxonomies.GET("/:slug", h.GetTaxonomyHandler)
		taxonomies.PUT("/:slug", h.UpdateTaxonomyHandler)
		taxonomies.DELETE("/:slug", h.DeleteTaxonomyHandler)
	}

	// Topic endpoints
	topics := v1.Group("/topics", middleware.Authenticated)
	{
		topics.GET("", h.ListTopicHandler)
		topics.POST("", h.CreateTopicHandler)
		topics.GET("/:slug", h.GetTopicHandler)
		topics.PUT("/:slug", h.UpdateTopicHandler)
		topics.DELETE("/:slug", h.DeleteTopicHandler)
	}

	// Swagger documentation endpoint
	if conf.RunMode != gin.ReleaseMode {
		e.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}
