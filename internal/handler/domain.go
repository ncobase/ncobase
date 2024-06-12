package handler

import (
	"stocms/internal/data/structs"
	"stocms/pkg/resp"

	"github.com/gin-gonic/gin"
)

// AccountDomainHandler handles reading the current user's domain.
//
// @Summary Get current user domain
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

// CreateDomainHandler handles creating a domain.
//
// @Summary Create domain
// @Description Create a new domain.
// @Tags domain
// @Accept json
// @Produce json
// @Param body body structs.CreateDomainBody true "CreateDomainBody object"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /domain [post]
func (h *Handler) CreateDomainHandler(c *gin.Context) {
	var body *structs.CreateDomainBody
	if err := c.ShouldBind(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.svc.CreateDomainService(c, body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// UserDomainHandler handles reading a user's domain.
//
// @Summary Get user domain
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

// GetDomainHandler handles reading domain information.
//
// @Summary Get domain
// @Description Retrieve information about a specific domain.
// @Tags domain
// @Produce json
// @Param id path string true "Domain ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /domain/{id} [get]
func (h *Handler) GetDomainHandler(c *gin.Context) {
	result, err := h.svc.GetDomainService(c, c.Param("id"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetDomainMenuHandler handles reading domain menu.
//
// @Summary Get domain menu
// @Description Retrieve the menu associated with a specific domain.
// @Tags domain
// @Produce json
// @Param id path string true "Domain ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /domain/{id}/menu [get]
func (h *Handler) GetDomainMenuHandler(c *gin.Context) {
	result, err := h.svc.GetDomainService(c, c.Param("id"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetDomainSettingHandler handles reading domain setting.
//
// @Summary Get domain setting
// @Description Retrieve the settings associated with a specific domain.
// @Tags domain
// @Produce json
// @Param id path string true "Domain ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /domain/{id}/setting [get]
func (h *Handler) GetDomainSettingHandler(c *gin.Context) {
	result, err := h.svc.GetDomainService(c, c.Param("id"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// DeleteDomainHandler handles deleting a domain.
//
// @Summary Delete domain
// @Description Delete a specific domain.
// @Tags domain
// @Produce json
// @Param id path string true "Domain ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /domain/{id} [delete]
func (h *Handler) DeleteDomainHandler(c *gin.Context) {
	result, err := h.svc.DeleteDomainService(c, c.Param("id"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// ListDomainHandler handles listing domains.
//
// @Summary List domains
// @Description Retrieve a list of domains.
// @Tags domain
// @Produce json
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /domain [get]
func (h *Handler) ListDomainHandler(c *gin.Context) {
	params := &structs.ListDomainParams{}
	if err := c.ShouldBindQuery(params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	if err := params.Validate(); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.svc.ListDomainsService(c, params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}
