package server

import (
	"net/http"
	"stocms/internal/config"
	"stocms/internal/handler"
	"stocms/internal/helper"
	"stocms/internal/server/middleware"

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

	// Authorization endpoints
	authorize := v1.Group("/authorize")
	authorize.POST("/send", h.SendCodeHandler)
	authorize.GET("/:code", h.CodeAuthHandler)

	// Account endpoints
	account := v1.Group("/account", middleware.Authorized)
	account.GET("", h.ReadMeHandler)
	account.POST("/password", h.UpdatePasswordHandler)
	account.GET("/domain", h.AccountDomainHandler)

	// User endpoints
	user := v1.Group("/users", middleware.Authorized)
	user.GET("/:username", h.ReadUserHandler)
	user.GET("/:username/domain", h.UserDomainHandler)

	// OAuth endpoints
	oauth := v1.Group("/oauth")
	oauth.POST("/signup", h.OAuthRegisterHandler)
	oauth.GET("/profile", h.GetOAuthProfileHandler)
	oauth.GET("/redirect/:provider", h.OAuthRedirectHandler)
	oauth.GET("/callback/github", h.OAuthGithubCallbackHandler, h.OAuthCallbackHandler)
	oauth.GET("/callback/facebook", h.OAuthFacebookCallbackHandler, h.OAuthCallbackHandler)

	// Resource endpoints
	resource := v1.Group("/resources")
	resource.GET("", h.ListResourceHandler)
	resource.POST("", h.CreateResourcesHandler)
	resource.GET("/:slug", h.GetResourceHandler)
	resource.DELETE("/:slug", h.DeleteResourceHandler)

	// module endpoints
	module := v1.Group("/modules")
	module.GET("", h.ListModuleHandler)
	module.POST("", h.CreateModuleHandler)
	module.GET("/:slug", h.GetModuleHandler)
	module.PUT("/:slug", h.UpdateModuleHandler)
	module.DELETE("/:slug", h.DeleteModuleHandler)

	// Taxonomy endpoints
	taxonomy := v1.Group("/taxonomies", middleware.Authorized)
	taxonomy.GET("", h.ListTaxonomyHandler)
	taxonomy.POST("", h.CreateTaxonomyHandler)
	taxonomy.GET("/:slug", h.GetTaxonomyHandler)
	taxonomy.PUT("/:slug", h.UpdateTaxonomyHandler)
	taxonomy.DELETE("/:slug", h.DeleteTaxonomyHandler)

	// Topic endpoints
	topic := v1.Group("/topics", middleware.Authorized)
	topic.GET("", h.ListTopicHandler)
	topic.POST("", h.CreateTopicHandler)
	topic.GET("/:slug", h.GetTopicHandler)
	topic.PUT("/:slug", h.UpdateTopicHandler)
	topic.DELETE("/:slug", h.DeleteTopicHandler)

	// Swagger documentation endpoint
	if conf.RunMode != gin.ReleaseMode {
		e.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}
