package handler

import (
	"stocms/pkg/resp"

	"github.com/gin-gonic/gin"
)

// ReadUserHandler Read user handler
func (h *Handler) ReadUserHandler(c *gin.Context) {
	result, err := h.svc.ReadUserService(c, c.Param("id"))
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
