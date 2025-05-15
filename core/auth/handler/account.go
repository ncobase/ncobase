package handler

import (
	"ncobase/auth/service"
	"ncobase/auth/structs"
	userStructs "ncobase/user/structs"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/net/cookie"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// AccountHandlerInterface is the interface for the handler.
type AccountHandlerInterface interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	Logout(c *gin.Context)
	GetMe(c *gin.Context)
	UpdatePassword(c *gin.Context)
	Tenant(c *gin.Context)
	Tenants(c *gin.Context)
	RefreshToken(c *gin.Context)
	TokenStatus(c *gin.Context)
}

// accountHandler represents the handler.
type accountHandler struct {
	s *service.Service
}

// NewAccountHandler creates a new handler.
func NewAccountHandler(svc *service.Service) AccountHandlerInterface {
	return &accountHandler{
		s: svc,
	}
}

// Register handles user registration.
//
// @Summary Register
// @Description Register a new user.
// @Tags iam
// @Accept json
// @Produce json
// @Param body body structs.RegisterBody true "RegisterBody object"
// @Success 200 {object} map[string]any{id=string,access_token=string} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/register [post]
func (h *accountHandler) Register(c *gin.Context) {
	body := &structs.RegisterBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, _ := h.s.Account.Register(c.Request.Context(), body)

	// _ = cookie.SetTokensFromResult(c.Writer, c.Request, result)
	resp.Success(c.Writer, result)
}

// Login handles user login.
//
// @Summary Login
// @Description Log in a user.
// @Tags iam
// @Accept json
// @Produce json
// @Param body body structs.LoginBody true "LoginBody object"
// @Success 200 {object} map[string]any{id=string,access_token=string} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/login [post]
func (h *accountHandler) Login(c *gin.Context) {
	body := &structs.LoginBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	// Validate captcha
	if body.Captcha != nil && body.Captcha.ID != "" && body.Captcha.Solution != "" {
		if err := h.s.Captcha.ValidateCaptcha(c.Request.Context(), body.Captcha); err != nil {
			resp.Fail(c.Writer, resp.BadRequest(err.Error()))
			return
		}
	}

	result, err := h.s.Account.Login(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	_ = cookie.SetTokensFromResult(c.Writer, c.Request, result)
	resp.Success(c.Writer, result)
}

// GetMe handles reading the current user.
//
// @Summary Get current user
// @Description Retrieve information about the current user.
// @Tags iam
// @Produce json
// @Success 200 {object} structs.AccountMeshes "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/account [get]
// @Security Bearer
func (h *accountHandler) GetMe(c *gin.Context) {
	result, err := h.s.Account.GetMe(c.Request.Context())
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Logout handles user logout.
//
// @Summary Logout
// @Description Logout the current user.
// @Tags iam
// @Produce json
// @Success 200 {object} resp.Exception "success"
// @Router /iam/logout [post]
// @Security Bearer
func (h *accountHandler) Logout(c *gin.Context) {
	cookie.ClearAll(c.Writer)
	resp.Success(c.Writer)
}

// RefreshToken handles token refresh.
//
// @Summary RefreshToken token
// @Description Refresh the current user's access token.
// @Tags iam
// @Accept json
// @Produce json
// @Param body body structs.RefreshTokenBody true "Refresh token"
// @Success 200 {object} map[string]any{id=string,access_token=string,refresh_token=string} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/refresh [post]
func (h *accountHandler) RefreshToken(c *gin.Context) {
	body := &structs.RefreshTokenBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Account.RefreshToken(c.Request.Context(), body.RefreshToken)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	_ = cookie.SetTokensFromResult(c.Writer, c.Request, result)
	resp.Success(c.Writer, result)
}

// TokenStatus checks token status without exposing sensitive information.
//
// @Summary Token status
// @Description Get the current token status.
// @Tags iam
// @Produce json
// @Success 200 {object} map[string]any{is_authenticated=bool} "success"
// @Failure 401 {object} resp.Exception "unauthorized"
// @Router /iam/token-status [get]
func (h *accountHandler) TokenStatus(c *gin.Context) {
	// Get user ID from context (set by authentication middleware)
	userID := ctxutil.GetUserID(c.Request.Context())
	if userID == "" {
		resp.Fail(c.Writer, resp.UnAuthorized("Not authenticated"))
		return
	}

	// Return basic authentication status
	resp.Success(c.Writer, map[string]any{
		"is_authenticated": true,
	})
}

// UpdatePassword handles updating user password.
//
// @Summary Update user password
// @Description Update the password of the current user.
// @Tags iam
// @Accept json
// @Produce json
// @Param body body structs.UserPassword true "UserPassword object"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/account/password [put]
// @Security Bearer
func (h *accountHandler) UpdatePassword(c *gin.Context) {
	body := &userStructs.UserPassword{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	err := h.s.Account.UpdatePassword(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// Tenant handles reading the current user's tenant.
//
// @Summary Get current user tenant
// @Description Retrieve the tenant associated with the current user.
// @Tags iam
// @Produce json
// @Success 200 {object} structs.ReadTenant "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/account/tenant [get]
// @Security Bearer
func (h *accountHandler) Tenant(c *gin.Context) {
	result, err := h.s.Account.Tenant(c.Request.Context())
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Tenants handles reading the current user's tenants.
//
// @Summary Get current user tenants
// @Description Retrieve the tenant associated with the current user.
// @Tags iam
// @Produce json
// @Success 200 {object} structs.ReadTenant "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/account/tenants [get]
// @Security Bearer
func (h *accountHandler) Tenants(c *gin.Context) {
	result, err := h.s.Account.Tenants(c.Request.Context())
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}
