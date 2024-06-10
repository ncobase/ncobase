package handler

import (
	"stocms/internal/data/structs"
	"stocms/pkg/resp"

	"github.com/gin-gonic/gin"
)

// AccountDomainHandler handles reading the current user's domain.
//
// @Summary Read current user domain
// @Description Retrieve the domain associated with the current user.
// @Tags account
// @Produce json
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /account/domain [get]
func (h *Handler) AccountDomainHandler(c *gin.Context) {
	result, err := h.svc.AccountDomainService(c)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// UserDomainHandler handles reading a user's domain.
//
// @Summary Read user domain
// @Description Retrieve the domain associated with the specified user.
// @Tags user
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /user/{username}/domain [get]
func (h *Handler) UserDomainHandler(c *gin.Context) {
	result, err := h.svc.UserDomainService(c, c.Param("username"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// UpdateDomainHandler handles updating a domain.
//
// @Summary Update domain
// @Description Update the domain information.
// @Tags domain
// @Accept json
// @Produce json
// @Param body body structs.UpdateDomainBody true "UpdateDomainBody object"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /domain [put]
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

// ReadDomainHandler handles reading domain information.
//
// @Summary Read domain
// @Description Retrieve information about a specific domain.
// @Tags domain
// @Produce json
// @Param id path string true "Domain ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /domain/{id} [get]
func (h *Handler) ReadDomainHandler(c *gin.Context) {
	result, err := h.svc.ReadDomainService(c, c.Param("id"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// ReadDomainMenuHandler handles reading domain menu.
//
// @Summary Read domain menu
// @Description Retrieve the menu associated with a specific domain.
// @Tags domain
// @Produce json
// @Param id path string true "Domain ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /domain/{id}/menu [get]
func (h *Handler) ReadDomainMenuHandler(c *gin.Context) {
	result, err := h.svc.ReadDomainService(c, c.Param("id"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// ReadDomainSettingHandler handles reading domain setting.
//
// @Summary Read domain setting
// @Description Retrieve the settings associated with a specific domain.
// @Tags domain
// @Produce json
// @Param id path string true "Domain ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /domain/{id}/setting [get]
func (h *Handler) ReadDomainSettingHandler(c *gin.Context) {
	result, err := h.svc.ReadDomainService(c, c.Param("id"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}
