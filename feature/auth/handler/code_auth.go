package handler

import (
	"ncobase/common/helper"
	"ncobase/common/resp"
	"ncobase/feature/auth/service"
	"ncobase/feature/auth/structs"

	"github.com/gin-gonic/gin"
)

// CodeAuthHandlerInterface is the interface for the handler.
type CodeAuthHandlerInterface interface {
	SendCode(c *gin.Context)
	CodeAuth(c *gin.Context)
}

// codeAuthHandler represents the handler.
type codeAuthHandler struct {
	s *service.Service
}

// NewCodeAuthHandler creates a new handler.
func NewCodeAuthHandler(svc *service.Service) CodeAuthHandlerInterface {
	return &codeAuthHandler{
		s: svc,
	}
}

// SendCode handles sending a verification code.
//
// @Summary Send verification code
// @Description Send a verification code to the specified destination.
// @Tags authorization
// @Accept json
// @Produce json
// @Param body body structs.SendCodeBody true "SendCodeBody object"
// @Success 200 {object} map[string]any{registered=bool} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/authorize/send [post]
func (h *codeAuthHandler) SendCode(c *gin.Context) {
	body := &structs.SendCodeBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, _ := h.s.CodeAuth.SendCode(c.Request.Context(), body)
	resp.Success(c.Writer, result)
}

// CodeAuth handles verifying a code.
//
// @Summary Verify code
// @Description Verify the provided code.
// @Tags authorization
// @Tags authentication
// @Accept json
// @Produce json
// @Param code path string true "Verification code"
// @Success 200 {object} map[string]any{id=string,access_token=string,email=string,register_token=string}  "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/authorize/{code} [get]
func (h *codeAuthHandler) CodeAuth(c *gin.Context) {
	result, err := h.s.CodeAuth.CodeAuth(c.Request.Context(), c.Param("code"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	// _ = cookie.SetRegisterTokenFromResult(c.Writer, c.Request, result)
	resp.Success(c.Writer, result)
}
