package server

import (
	"net/http"
	"ncobase/internal/config"
	"ncobase/internal/handler"
	"ncobase/internal/helper"
	"ncobase/internal/server/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func registerRest(e *gin.Engine, h *handler.Handler, conf *config.Config) {
	// Root Jump when domain is configured and it is not localhost
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

	// API prefix for v1 version
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
	authorize.POST("/send", h.SendCodeHandler)
	authorize.GET("/:code", h.CodeAuthHandler)

	// Account endpoints
	account := v1.Group("/account", middleware.Authorized)
	{
		account.GET("", h.GetMeHandler)
		account.PUT("/password", h.UpdatePasswordHandler)
		account.GET("/domain", h.AccountDomainHandler)
	}

	// User endpoints
	user := v1.Group("/users")
	{
		// user.GET("", h.ListUserHandler)
		// user.POST("", h.CreateUserHandler)
		user.GET("/:username", h.GetUserHandler)
		// user.PUT("/:username", h.UpdateUserHandler)
		// user.DELETE("/:username", h.DeleteUserHandler)
		// user.GET("/:username/roles", h.ListUserRoleHandler)
		// user.GET("/:username/groups", h.ListUserGroupHandler)
		// user.GET("/:username/domain", h.ListUserDomainHandler)
		user.GET("/:username/domain", middleware.Authorized, h.UserDomainHandler)
		// user.GET("/:username/domain/belongs", middleware.Authorized, h.ListUserBelongHandler)
	}

	// OAuth endpoints
	oauth := v1.Group("/oauth")
	{
		oauth.POST("/signup", h.OAuthRegisterHandler)
		oauth.GET("/profile", h.GetOAuthProfileHandler)
		oauth.GET("/redirect/:provider", h.OAuthRedirectHandler)
		oauth.GET("/callback/github", h.OAuthGithubCallbackHandler, h.OAuthCallbackHandler)
		oauth.GET("/callback/facebook", h.OAuthFacebookCallbackHandler, h.OAuthCallbackHandler)
	}

	// Module endpoints
	module := v1.Group("/modules")
	{
		module.GET("", h.ListModuleHandler)
		module.POST("", middleware.Authorized, h.CreateModuleHandler)
		module.GET("/:slug", h.GetModuleHandler)
		module.PUT("/:slug", middleware.Authorized, h.UpdateModuleHandler)
		module.DELETE("/:slug", middleware.Authorized, h.DeleteModuleHandler)
	}

	// Asset endpoints
	asset := v1.Group("/assets")
	{
		asset.GET("", h.ListAssetHandler)
		asset.POST("", middleware.Authorized, h.CreateAssetsHandler)
		asset.GET("/:slug", h.GetAssetHandler)
		asset.PUT("/:slug", middleware.Authorized, h.UpdateAssetHandler)
		asset.DELETE("/:slug", middleware.Authorized, h.DeleteAssetHandler)
	}

	// Domain endpoints
	domain := v1.Group("/domains", middleware.Authorized)
	{
		domain.GET("", h.ListDomainHandler)
		domain.POST("", h.CreateDomainHandler)
		domain.GET("/:slug", h.GetDomainHandler)
		domain.PUT("/:slug", h.UpdateDomainHandler)
		domain.DELETE("/:slug", h.DeleteDomainHandler)
		domain.GET("/:slug/assets", h.ListDomainAssetHandler)
		domain.GET("/:slug/users", h.ListDomainUserHandler)
		domain.GET("/:slug/groups", h.ListDomainGroupHandler)
	}

	// Group endpoints
	// group := v1.Group("/groups", middleware.Authorized)
	// {
	// 	group.GET("", h.ListGroupHandler)
	// 	group.POST("", h.CreateGroupHandler)
	// 	group.GET("/:slug", h.GetGroupHandler)
	// 	group.PUT("/:slug", h.UpdateGroupHandler)
	// 	group.DELETE("/:slug", h.DeleteGroupHandler)
	// 	group.GET("/:slug/roles", h.ListGroupRoleHandler)
	// 	group.GET("/:slug/users", h.ListGroupUserHandler)
	// }

	// Role endpoints
	// role := v1.Group("/roles", middleware.Authorized)
	// {
	// 	role.GET("", h.ListRoleHandler)
	// 	role.POST("", h.CreateRoleHandler)
	// 	role.GET("/:slug", h.GetRoleHandler)
	// 	role.PUT("/:slug", h.UpdateRoleHandler)
	// 	role.DELETE("/:slug", h.DeleteRoleHandler)
	// 	role.GET("/:slug/permissions", h.ListRolePermissionHandler)
	// 	role.GET("/:slug/users", h.ListUserRoleHandler)
	// }

	// Permission endpoints
	// permission := v1.Group("/permissions", middleware.Authorized)
	// {
	// 	permission.GET("", h.ListPermissionHandler)
	// 	permission.POST("", h.CreatePermissionHandler)
	// 	permission.GET("/:slug", h.GetPermissionHandler)
	// 	permission.PUT("/:slug", h.UpdatePermissionHandler)
	// 	permission.DELETE("/:slug", h.DeletePermissionHandler)
	// }

	// Taxonomy endpoints
	taxonomy := v1.Group("/taxa")
	{
		taxonomy.GET("", h.ListTaxonomyHandler)
		taxonomy.POST("", middleware.Authorized, h.CreateTaxonomyHandler)
		taxonomy.GET("/:slug", h.GetTaxonomyHandler)
		taxonomy.PUT("/:slug", middleware.Authorized, h.UpdateTaxonomyHandler)
		taxonomy.DELETE("/:slug", middleware.Authorized, h.DeleteTaxonomyHandler)
	}

	// Topic endpoints
	topic := v1.Group("/topics")
	{
		topic.GET("", h.ListTopicHandler)
		topic.POST("", middleware.Authorized, h.CreateTopicHandler)
		topic.GET("/:slug", h.GetTopicHandler)
		topic.PUT("/:slug", middleware.Authorized, h.UpdateTopicHandler)
		topic.DELETE("/:slug", middleware.Authorized, h.DeleteTopicHandler)
	}

	// Casbin Rule endpoints
	casbin := v1.Group("/pols", middleware.Authorized)
	{
		casbin.GET("", h.ListCasbinRuleHandler)
		casbin.POST("", h.CreateCasbinRuleHandler)
		casbin.GET("/:id", h.GetCasbinRuleHandler)
		casbin.PUT("/:id", h.UpdateCasbinRuleHandler)
		casbin.DELETE("/:id", h.DeleteCasbinRuleHandler)
	}

	// ******************************
	// Admin endpoints
	// ******************************

	// Swagger documentation endpoint
	if conf.RunMode != gin.ReleaseMode {
		e.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}
