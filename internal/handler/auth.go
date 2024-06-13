package handler

import (
	"stocms/internal/data/structs"
	"stocms/internal/helper"
	"stocms/pkg/cookie"
	"stocms/pkg/resp"

	"github.com/gin-gonic/gin"
)

// SendCodeHandler handles sending a verification code.
//
// @Summary Send verification code
// @Description Send a verification code to the specified destination.
// @Tags authorization
// @Accept json
// @Produce json
// @Param body body structs.SendCodeBody true "SendCodeBody object"
// @Success 200 {object} types.JSON{registered=bool} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/authorize/send [post]
func (h *Handler) SendCodeHandler(c *gin.Context) {
	body := &structs.SendCodeBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, _ := h.svc.SendCodeService(c, body)
	resp.Success(c.Writer, result)
}

// CodeAuthHandler handles verifying a code.
//
// @Summary Verify code
// @Description Verify the provided code.
// @Tags authorization
// @Tags authentication
// @Accept json
// @Produce json
// @Param code path string true "Verification code"
// @Success 200 {object} types.JSON{id=string,access_token=string,email=string,register_token=string}  "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/authorize/{code} [get]
func (h *Handler) CodeAuthHandler(c *gin.Context) {
	result, err := h.svc.CodeAuthService(c, c.Param("code"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// RegisterHandler handles user registration.
//
// @Summary Register
// @Description Register a new user.
// @Tags authentication
// @Accept json
// @Produce json
// @Param body body structs.RegisterBody true "RegisterBody object"
// @Success 200 {object} types.JSON{id=string,access_token=string} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/register [post]
func (h *Handler) RegisterHandler(c *gin.Context) {
	body := &structs.RegisterBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, _ := h.svc.RegisterService(c, body)
	resp.Success(c.Writer, result)
}

// LogoutHandler handles user logout.
//
// @Summary Logout
// @Description Logout the current user.
// @Tags authentication
// @Produce json
// @Success 200 {object} resp.Exception "success"
// @Router /v1/logout [post]
// @Security Bearer
func (h *Handler) LogoutHandler(c *gin.Context) {
	cookie.ClearAll(c.Writer)
	resp.Success(c.Writer, nil)
}

// LoginHandler handles user login.
//
// @Summary Login
// @Description Log in a user.
// @Tags authentication
// @Accept json
// @Produce json
// @Param body body structs.LoginBody true "LoginBody object"
// @Success 200 {object} types.JSON{id=string,access_token=string} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/login [post]
func (h *Handler) LoginHandler(c *gin.Context) {
	body := &structs.LoginBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}
	result, err := h.svc.LoginService(c, body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// // RefreshHandler handles user token refresh.
// //
// // @Summary Refresh
// // @Description Refresh the current user's access token.
// // @Tags authentication
// // @Produce json
// // @Success 200 {object} types.JSON{id=string,access_token=string} "success"
// // @Failure 400 {object} resp.Exception "bad request"
// // @Router /v1/refresh [post]
// // @Security Bearer
// func (h *Handler) RefreshHandler(c *gin.Context) {
// 	result, err := h.svc.RefreshService(c)
// 	if err != nil {
// 		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
// 		return
// 	}
// 	resp.Success(c.Writer, result)
// }
