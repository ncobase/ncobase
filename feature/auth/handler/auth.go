package handler

import (
	"ncobase/common/cookie"
	"ncobase/common/resp"
	"ncobase/feature/auth/service"
	"ncobase/feature/auth/structs"
	"ncobase/helper"

	"github.com/gin-gonic/gin"
)

// AuthHandlerInterface is the interface for the handler.
type AuthHandlerInterface interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	Logout(c *gin.Context)
}

// authHandler represents the handler.
type authHandler struct {
	s *service.Service
}

// NewAuthHandler creates a new handler.
func NewAuthHandler(svc *service.Service) AuthHandlerInterface {
	return &authHandler{
		s: svc,
	}
}

// Register handles user registration.
//
// @Summary Register
// @Description Register a new user.
// @Tags authentication
// @Accept json
// @Produce json
// @Param body body structs.RegisterBody true "RegisterBody object"
// @Success 200 {object} map[string]any{id=string,access_token=string} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/register [post]
func (h *authHandler) Register(c *gin.Context) {
	body := &structs.RegisterBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, _ := h.s.Auth.Register(c.Request.Context(), body)
	resp.Success(c.Writer, result)
}

// Login handles user login.
//
// @Summary Login
// @Description Log in a user.
// @Tags authentication
// @Accept json
// @Produce json
// @Param body body structs.LoginBody true "LoginBody object"
// @Success 200 {object} map[string]any{id=string,access_token=string} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/login [post]
func (h *authHandler) Login(c *gin.Context) {
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

	result, err := h.s.Auth.Login(c.Request.Context(), body)
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
// @Tags authentication
// @Produce json
// @Success 200 {object} resp.Exception "success"
// @Router /v1/logout [post]
// @Security Bearer
func (h *authHandler) Logout(c *gin.Context) {
	cookie.ClearAll(c.Writer)
	resp.Success(c.Writer)
}

// // Refresh handles user token refresh.
// //
// // @Summary Refresh
// // @Description Refresh the current user's access token.
// // @Tags authentication
// // @Produce json
// // @Success 200 {object} map[string]any{id=string,access_token=string} "success"
// // @Failure 400 {object} resp.Exception "bad request"
// // @Router /v1/refresh [post]
// // @Security Bearer
// func (h *Handler) Refresh(c *gin.Context) {
// 	result, err := h.svc.RefreshServicec.Request.Context()
// 	if err != nil {
// 		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
// 		return
// 	}
// 	resp.Success(c.Writer, result)
// }
