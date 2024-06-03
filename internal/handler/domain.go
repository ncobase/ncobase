package handler

import (
	"stocms/internal/data/structs"
	"stocms/pkg/resp"

	"github.com/gin-gonic/gin"
)

// AccountDomainHandler Read current user domain handler
func (h *Handler) AccountDomainHandler(c *gin.Context) {
	result, err := h.svc.AccountDomainService(c)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// UserDomainHandler Read user domain handler
func (h *Handler) UserDomainHandler(c *gin.Context) {
	result, err := h.svc.UserDomainService(c, c.Param("username"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// UpdateDomainHandler Update domain handler
func (h *Handler) UpdateDomainHandler(c *gin.Context) {
	var body *structs.UpdateDomainBody
	if err := c.ShouldBind(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.svc.UpdateDomainService(c, body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// ReadDomainHandler Read domain handler
func (h *Handler) ReadDomainHandler(c *gin.Context) {
	result, err := h.svc.ReadDomainService(c, c.Param("id"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// ReadDomainMenuHandler Read domain menu handler
func (h *Handler) ReadDomainMenuHandler(c *gin.Context) {
	result, err := h.svc.ReadDomainService(c, c.Param("id"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// ReadDomainSettingHandler Read domain setting handler
func (h *Handler) ReadDomainSettingHandler(c *gin.Context) {
	result, err := h.svc.ReadDomainService(c, c.Param("id"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}
