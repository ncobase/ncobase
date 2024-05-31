package server

import (
	"net/http"
	"stocms/internal/config"
	"stocms/internal/handler"
	"stocms/internal/helper"
	"stocms/internal/server/middleware"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

func registerRest(e *gin.Engine, h *handler.Handler, conf *config.Config) {
	// root Jump when domain is configured and it is not localhost
	e.GET("/", func(c *gin.Context) {
		if conf.Domain != "localhost" {
			url := helper.GetHost(conf, conf.Domain)
			c.Redirect(http.StatusMovedPermanently, url)
		} else {
			c.String(http.StatusOK, "It's working.")
		}
	})

	// Health
	e.GET("/health", h.HealthHandler)

	// api prefix for v1 version
	v1 := e.Group("/v1")

	// ****** Authentication
	// sign up
	v1.POST("/register", h.RegisterHandler)
	// logout
	v1.POST("/logout", h.LogoutHandler)

	// ****** Authorization
	authorize := v1.Group("/authorize")
	// send authorization code
	authorize.POST("/send", h.SendCodeHandler)
	// verify the verification code to obtain a register token or access token
	authorize.GET("/:code", h.CodeAuthHandler)
	// verify authorization and return current user
	authorize.GET("", middleware.Authorized, h.ReadMeHandler)

	// // ****** Account
	account := v1.Group("/account", middleware.Authorized)
	// current user，with /authorize same
	account.GET("", h.ReadMeHandler)
	// // update current user
	// // account.PUT("", h.UpdateMeHandler)
	// get domains of current user
	// account.GET("/domains", h.AccountDomainHandler)
	// get domain detail of current user
	account.GET("/domains/:id", h.ReadDomainHandler)
	//
	// // ****** User
	user := v1.Group("/users", middleware.Authorized)
	// get detail of user
	user.GET("/:id", h.ReadUserHandler)

	// // get user setting TODO
	// user.GET("/settings", h.Ping)
	// // update user setting TODO
	// user.PUT("/settings", h.Ping)
	// // get user security TODO
	// user.GET("/security", h.Ping)
	// // update user security TODO
	// user.PUT("/security", h.Ping)
	// // update user password and reset TODO
	// user.PUT("/password", h.Ping)
	// get user domain
	// user.GET("/domains", h.ReadDomainHandler)
	// update domain of the user
	// user.PUT("/domains", h.UpdateDomainHandler)

	// // ****** OAuth register callback and redirect
	// oauth := v1.Group("/oauth")
	// oauth.POST("/signup", h.OAuthRegisterHandler)
	// oauth.GET("/profile", h.GetOAuthProfileHandler)
	// oauth.GET("/redirect/:provider", h.OAuthRedirectHandler)
	// oauth.GET("/callback/github", h.OAuthGithubCallbackHandler, h.OAuthCallbackHandler)
	// oauth.GET("/callback/facebook", h.OAuthFacebookCallbackHandler, h.OAuthCallbackHandler)
	//
	// // ****** Domain TODO
	// domain := v1.Group("/domains", middleware.Authorized)
	// // get config params of domain
	// domain.GET("/params", h.Ping)
	// // get dictionary of domain
	// domain.GET("/dictionaries", h.Ping)
	// // get have modules of domain
	// domain.GET("/modules", h.Ping)

	// Swagger
	v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
