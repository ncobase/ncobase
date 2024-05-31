package handler

import (
	"stocms/internal/data/structs"
	"stocms/pkg/cookie"
	"stocms/pkg/resp"

	"github.com/gin-gonic/gin"
)

// SendCodeHandler Send verify code handler
func (h *Handler) SendCodeHandler(c *gin.Context) {
	var body *structs.SendCodeBody
	if err := c.ShouldBind(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	}

	result, _ := h.svc.SendCodeService(c, body)
	resp.Success(c.Writer, result)
}

// CodeAuthHandler verify code handler
func (h *Handler) CodeAuthHandler(c *gin.Context) {
	result, err := h.svc.CodeAuthService(c, c.Param("code"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// RegisterHandler Register handler
func (h *Handler) RegisterHandler(c *gin.Context) {
	var body *structs.RegisterBody
	if err := c.ShouldBind(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, _ := h.svc.RegisterService(c, body)
	resp.Success(c.Writer, result)
}

// LogoutHandler Logout handler
func (h *Handler) LogoutHandler(c *gin.Context) {
	cookie.ClearAll(c.Writer)
	resp.Success(c.Writer, nil)
}
