package handler

import (
	"stocms/internal/data/structs"
	"stocms/pkg/resp"

	"github.com/gin-gonic/gin"
)

// ReadUserHandler Read user handler
func (h *Handler) ReadUserHandler(c *gin.Context) {
	result, err := h.svc.ReadUserService(c, c.Param("username"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// ReadMeHandler Read current user handler
func (h *Handler) ReadMeHandler(c *gin.Context) {
	result, err := h.svc.ReadMeService(c)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// UpdatePasswordHandler Update user password handler
func (h *Handler) UpdatePasswordHandler(c *gin.Context) {
	var body *structs.UserRequestBody
	if err := c.ShouldBind(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.svc.UpdatePasswordService(c, body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}
