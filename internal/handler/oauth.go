package handler

import (
	"ncobase/internal/data/structs"
	"ncobase/internal/helper"
	"net/http"

	"github.com/ncobase/common/oauth"
	"github.com/ncobase/common/resp"
	"github.com/ncobase/common/types"

	"github.com/gin-gonic/gin"
)

// OAuthRegisterHandler handles OAuth registration.
//
// @Summary OAuth register
// @Description Register a user using OAuth.
// @Tags oauth
// @Accept json
// @Produce json
// @Param body body structs.OAuthRegisterBody true "OAuthRegisterBody object"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/oauth/register [post]
func (h *Handler) OAuthRegisterHandler(c *gin.Context) {
	body := &structs.OAuthRegisterBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, _ := h.svc.OAuthRegisterService(c, body)
	resp.Success(c.Writer, result)
}

// OAuthRedirectHandler handles OAuth redirection.
//
// @Summary OAuth redirect
// @Description Redirect to OAuth provider for authentication.
// @Tags oauth
// @Param provider path string true "OAuth provider"
// @Param next query string false "Next URL after authentication"
// @Success 302 {object} resp.Exception "redirect"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/oauth/{provider}/redirect [get]
func (h *Handler) OAuthRedirectHandler(c *gin.Context) {
	provider := c.Param("provider")
	next := c.Query("next")

	allowedProviders := map[string]bool{
		"facebook": true,
		"github":   true,
	}

	if !allowedProviders[provider] {
		resp.Fail(c.Writer, resp.BadRequest("PROVIDER_INVALID"))
		return
	}

	config, exists := h.OAuthConfig[provider]
	if !exists {
		resp.Fail(c.Writer, resp.BadRequest("PROVIDER_CONFIG_NOT_FOUND"))
		return
	}

	redirectUrl := oauth.GenerateOAuthLink(provider, config.ClientID, config.RedirectURL, next)
	c.Redirect(http.StatusMovedPermanently, redirectUrl)
}

// OAuthFacebookCallbackHandler handles Facebook OAuth callback.
//
// @Summary Facebook OAuth callback
// @Description Handle callback from Facebook OAuth provider.
// @Tags oauth
// @Param code query string true "Authorization code"
// @Success 200 {object} resp.Exception "success"
// @Failure 401 {object} resp.Exception "unauthorized"
// @Router /v1/oauth/facebook/callback [get]
func (h *Handler) OAuthFacebookCallbackHandler(c *gin.Context) {
	result, _ := h.svc.OAuthCallbackService(c, "facebook", c.Query("code"))
	if result.Code != http.StatusOK {
		resp.UnAuthorized("OAuth Auth Error", nil)
		return
	}
	c.Next()
}

// OAuthGithubCallbackHandler handles GitHub OAuth callback.
//
// @Summary GitHub OAuth callback
// @Description Handle callback from GitHub OAuth provider.
// @Tags oauth
// @Param code query string true "Authorization code"
// @Success 200 {object} resp.Exception "success"
// @Failure 401 {object} resp.Exception "unauthorized"
// @Router /v1/oauth/github/callback [get]
func (h *Handler) OAuthGithubCallbackHandler(c *gin.Context) {
	result, _ := h.svc.OAuthCallbackService(c, "github", c.Query("code"))
	if result.Code != http.StatusOK {
		resp.UnAuthorized("OAuth Auth Error", nil)
		return
	}
	c.Next()
}

// OAuthCallbackHandler handles generic OAuth callback.
//
// @Summary OAuth callback
// @Description Handle callback from OAuth provider.
// @Tags oauth
// @Success 302 {object} resp.Exception "redirect"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/oauth/callback [get]
func (h *Handler) OAuthCallbackHandler(c *gin.Context) {
	result, _ := h.svc.OAuthAuthenticationService(c)
	if result.Code == http.StatusMovedPermanently {
		c.Redirect(http.StatusMovedPermanently, result.Data.(types.JSON)["redirectUrl"].(string))
		return
	}
	c.JSON(result.Code, result)
}

// GetOAuthProfileHandler handles getting OAuth profile information.
//
// @Summary Get OAuth profile
// @Description Retrieve profile information from OAuth provider.
// @Tags oauth
// @Produce json
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/oauth/profile [get]
func (h *Handler) GetOAuthProfileHandler(c *gin.Context) {
	result, _ := h.svc.GetOAuthProfileInfoService(c)
	c.JSON(result.Code, result)
}
