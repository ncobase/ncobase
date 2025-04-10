package handler

import (
	"github.com/ncobase/ncore/pkg/cookie"
	"github.com/ncobase/ncore/pkg/helper"
	"github.com/ncobase/ncore/pkg/resp"
	"ncobase/core/auth/service"
	"ncobase/core/auth/structs"
	userStructs "ncobase/core/user/structs"

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
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
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
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
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

// // Refresh handles user token refresh.
// //
// // @Summary Refresh
// // @Description Refresh the current user's access token.
// // @Tags iam
// // @Produce json
// // @Success 200 {object} map[string]any{id=string,access_token=string} "success"
// // @Failure 400 {object} resp.Exception "bad request"
// // @Router /iam/refresh [post]
// // @Security Bearer
// func (h *Handler) Refresh(c *gin.Context) {
// 	result, err := h.svc.RefreshServicec.Request.Context()
// 	if err != nil {
// 		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
// 		return
// 	}
// 	resp.Success(c.Writer, result)
// }

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
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
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
