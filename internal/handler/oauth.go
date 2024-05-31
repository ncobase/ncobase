package handler

import (
	"net/http"
	"stocms/internal/data/structs"
	"stocms/pkg/oauth"
	"stocms/pkg/resp"
	"stocms/pkg/types"

	"github.com/gin-gonic/gin"
)

// OAuthRegisterHandler OAuth register handler
func (h *Handler) OAuthRegisterHandler(c *gin.Context) {
	var body *structs.OAuthRegisterBody
	if err := c.ShouldBind(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, _ := h.svc.OAuthRegisterService(c, body)
	resp.Success(c.Writer, result)
}

// OAuthRedirectHandler OAuth redirect handler
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

// OAuthFacebookCallbackHandler Facebook OAuth callback handler
func (h *Handler) OAuthFacebookCallbackHandler(c *gin.Context) {
	result, _ := h.svc.OAuthCallbackService(c, "facebook", c.Query("code"))
	if result.Code != http.StatusOK {
		resp.UnAuthorized("OAuth Auth Error", nil)
		return
	}
	c.Next()
}

// OAuthGithubCallbackHandler  Github OAuth callback handler
func (h *Handler) OAuthGithubCallbackHandler(c *gin.Context) {
	result, _ := h.svc.OAuthCallbackService(c, "github", c.Query("code"))
	if result.Code != http.StatusOK {
		resp.UnAuthorized("OAuth Auth Error", nil)
		return
	}
	c.Next()
}

// OAuthCallbackHandler  OAuth callback handler
func (h *Handler) OAuthCallbackHandler(c *gin.Context) {
	result, _ := h.svc.OAuthAuthenticationService(c)
	if result.Code == http.StatusMovedPermanently {
		c.Redirect(http.StatusMovedPermanently, result.Data.(types.JSON)["redirectUrl"].(string))
		return
	}
	c.JSON(result.Code, result)
}

// GetOAuthProfileHandler Get OAuth profile handler
func (h *Handler) GetOAuthProfileHandler(c *gin.Context) {
	result, _ := h.svc.GetOAuthProfileInfoService(c)
	c.JSON(result.Code, result)
}
