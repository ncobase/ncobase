package handler

//
// import (
// 	"ncobase/common/cookie"
// 	"ncobase/feature/auth/service"
// 	"ncobase/feature/auth/structs"
// 	"ncobase/helper"
// 	"net/http"
//
// 	"ncobase/common/oauth"
// 	"ncobase/common/resp"
// 	"ncobase/common/types"
//
// 	"github.com/gin-gonic/gin"
// )
//
// // OAuthHandlerInterface is the interface for the handler.
// type OAuthHandlerInterface interface {
// 	OAuthRegister(c *gin.Context)
// 	OAuthRedirect(c *gin.Context)
// 	OAuthFacebookCallback(c *gin.Context)
// 	OAuthGithubCallback(c *gin.Context)
// 	OAuthCallback(c *gin.Context)
// 	GetOAuthProfile(c *gin.Context)
// }
//
// // oAuthHandler represents the handler.
// type oAuthHandler struct {
// 	OAuthConfig map[string]oauth.ProviderConfig
// 	s           *service.Service
// }
//
// // NewOAuthHandler creates a new handler.
// func NewOAuthHandler(svc *service.Service) OAuthHandlerInterface {
// 	return &oAuthHandler{
// 		s: svc,
// 	}
// }
//
// // OAuthRegister handles OAuth registration.
// //
// // @Summary OAuth register
// // @Description Register a user using OAuth.
// // @Tags oauth
// // @Accept json
// // @Produce json
// // @Param body body structs.OAuthRegisterBody true "OAuthRegisterBody object"
// // @Success 200 {object} resp.Exception "success"
// // @Failure 400 {object} resp.Exception "bad request"
// // @Router /auth/oauth/register [post]
// func (h *oAuthHandler) OAuthRegister(c *gin.Context) {
// 	body := &structs.OAuthRegisterBody{}
// 	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
// 		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
// 		return
// 	} else if len(validationErrors) > 0 {
// 		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
// 		return
// 	}
//
// 	result, _ := h.s.OAuth.OAuthRegister(c.Request.Context(), body)
// 	resp.Success(c.Writer, result)
// }
//
// // OAuthRedirect handles OAuth redirection.
// //
// // @Summary OAuth redirect
// // @Description Redirect to OAuth provider for authentication.
// // @Tags oauth
// // @Param provider path string true "OAuth provider"
// // @Param next query string false "Next URL after authentication"
// // @Success 302 {object} resp.Exception "redirect"
// // @Failure 400 {object} resp.Exception "bad request"
// // @Router /auth/oauth/{provider}/redirect [get]
// func (h *oAuthHandler) OAuthRedirect(c *gin.Context) {
// 	provider := c.Param("provider")
// 	next := c.Query("next")
//
// 	allowedProviders := map[string]bool{
// 		"facebook": true,
// 		"github":   true,
// 	}
//
// 	if !allowedProviders[provider] {
// 		resp.Fail(c.Writer, resp.BadRequest("PROVIDER_INVALID"))
// 		return
// 	}
//
// 	config, exists := h.OAuthConfig[provider]
// 	if !exists {
// 		resp.Fail(c.Writer, resp.BadRequest("PROVIDER_CONFIG_NOT_FOUND"))
// 		return
// 	}
//
// 	redirectUrl := oauth.GenerateOAuthLink(provider, config.ClientID, config.RedirectURL, next)
// 	c.Redirect(http.StatusMovedPermanently, redirectUrl)
// }
//
// // OAuthFacebookCallback handles Facebook OAuth callback.
// //
// // @Summary Facebook OAuth callback
// // @Description Handle callback from Facebook OAuth provider.
// // @Tags oauth
// // @Param code query string true "Authorization code"
// // @Success 200 {object} resp.Exception "success"
// // @Failure 401 {object} resp.Exception "unauthorized"
// // @Router /auth/oauth/facebook/callback [get]
// func (h *oAuthHandler) OAuthFacebookCallback(c *gin.Context) {
// 	result, _ := h.s.OAuth.OAuthCallback(c.Request.Context(), "facebook", c.Query("code"))
// 	if result.Code != http.StatusOK {
// 		resp.UnAuthorized("OAuth Auth Error", nil)
// 		return
// 	}
// 	c.Next()
// }
//
// // OAuthGithubCallback handles GitHub OAuth callback.
// //
// // @Summary GitHub OAuth callback
// // @Description Handle callback from GitHub OAuth provider.
// // @Tags oauth
// // @Param code query string true "Authorization code"
// // @Success 200 {object} resp.Exception "success"
// // @Failure 401 {object} resp.Exception "unauthorized"
// // @Router /auth/oauth/github/callback [get]
// func (h *oAuthHandler) OAuthGithubCallback(c *gin.Context) {
// 	result, _ := h.s.OAuth.OAuthCallback(c.Request.Context(), "github", c.Query("code"))
// 	if result.Code != http.StatusOK {
// 		resp.UnAuthorized("OAuth Auth Error", nil)
// 		return
// 	}
// 	c.Next()
// }
//
// // OAuthCallback handles generic OAuth callback.
// //
// // @Summary OAuth callback
// // @Description Handle callback from OAuth provider.
// // @Tags oauth
// // @Success 302 {object} resp.Exception "redirect"
// // @Failure 400 {object} resp.Exception "bad request"
// // @Router /auth/oauth/callback [get]
// func (h *oAuthHandler) OAuthCallback(c *gin.Context) {
// 	result, _ := h.s.OAuth.OAuthAuthentication(c.Request.Context())
// 	if result.Code == http.StatusMovedPermanently {
// 		c.Redirect(http.StatusMovedPermanently, result.Data.(types.JSON)["redirectUrl"].(string))
// 		return
// 	}
// 	c.JSON(result.Code, result)
// }
//
// // GetOAuthProfile handles getting OAuth profile information.
// //
// // @Summary Get OAuth profile
// // @Description Retrieve profile information from OAuth provider.
// // @Tags oauth
// // @Produce json
// // @Success 200 {object} resp.Exception "success"
// // @Failure 400 {object} resp.Exception "bad request"
// // @Router /auth/oauth/profile [get]
// func (h *oAuthHandler) GetOAuthProfile(c *gin.Context) {
// 	registerToken, err := cookie.Get(c.Request, "register_token")
// 	if err != nil {
// 		resp.Fail(c.Writer, resp.BadRequest("register authorize is empty or invalid"))
// 		return
// 	}
// 	result, _ := h.s.OAuth.GetOAuthProfileInfo(c.Request.Context(), registerToken)
// 	c.JSON(result.Code, result)
// }
