package handler

import (
	"stocms/internal/data/structs"
	"stocms/internal/helper"
	"stocms/pkg/ecode"
	"stocms/pkg/resp"

	"github.com/gin-gonic/gin"
)

// AccountDomainHandler handles reading the current user's domain.
//
// @Summary Get current user domain
// @Description Retrieve the domain associated with the current user.
// @Tags account
// @Produce json
// @Success 200 {object} structs.ReadDomain "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/account/dom [get]
// @Security Bearer
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
// @Success 200 {object} structs.ReadDomain "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/dom [post]
// @Security Bearer
func (h *Handler) CreateDomainHandler(c *gin.Context) {
	body := &structs.CreateDomainBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
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
// @Summary Get user owned domain
// @Description Retrieve the domain associated with the specified user.
// @Tags user
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} structs.ReadDomain "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/users/{username}/dom [get]
// @Security Bearer
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
// @Param slug path string true "Domain ID"
// @Param body body structs.UpdateDomainBody true "UpdateDomainBody object"
// @Success 200 {object} structs.ReadDomain "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/dom [put]
// @Security Bearer
func (h *Handler) UpdateDomainHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}
	body := &structs.UpdateDomainBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
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
// @Param slug path string true "Domain ID"
// @Success 200 {object} structs.ReadDomain "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/dom/{slug} [get]
// @Security Bearer
func (h *Handler) GetDomainHandler(c *gin.Context) {
	result, err := h.svc.GetDomainService(c, c.Param("slug"))
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
// @Param slug path string true "Domain ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/dom/{slug}/menu [get]
// @Security Bearer
func (h *Handler) GetDomainMenuHandler(c *gin.Context) {
	result, err := h.svc.GetDomainService(c, c.Param("slug"))
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
// @Param slug path string true "Domain ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/dom/{slug}/setting [get]
// @Security Bearer
func (h *Handler) GetDomainSettingHandler(c *gin.Context) {
	result, err := h.svc.GetDomainService(c, c.Param("slug"))
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
// @Param slug path string true "Domain ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/dom/{slug} [delete]
// @Security Bearer
func (h *Handler) DeleteDomainHandler(c *gin.Context) {
	result, err := h.svc.DeleteDomainService(c, c.Param("slug"))
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
// @Param params query structs.ListDomainParams true "List domain parameters"
// @Success 200 {array} structs.ReadDomain"success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/dom [get]
// @Security Bearer
func (h *Handler) ListDomainHandler(c *gin.Context) {
	params := &structs.ListDomainParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.svc.ListDomainsService(c, params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// ListDomainAssetHandler handles listing domain assets.
// TODO: implement this
// @Summary List domain assets
// @Description Retrieve a list of assets associated with a specific domain.
// @Tags domain
// @Produce json
// @Param slug path string true "Domain ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/dom/{slug}/assets [get]
// @Security Bearer
func (h *Handler) ListDomainAssetHandler(c *gin.Context) {
	// result, err := h.svc.ListDomainAssetsService(c, c.Param("slug"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer, nil)
}

// ListDomainUserHandler handles listing domain users.
// TODO: implement this
// @Summary List domain users
// @Description Retrieve a list of users associated with a specific domain.
// @Tags domain
// @Produce json
// @Param slug path string true "Domain ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/dom/{slug}/users [get]
// @Security Bearer
func (h *Handler) ListDomainUserHandler(c *gin.Context) {
	// result, err := h.svc.ListDomainUsersService(c, c.Param("slug"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer, nil)
}

// ListDomainGroupHandler handles listing domain groups.
// TODO: implement this
// @Summary List domain groups
// @Description Retrieve a list of groups associated with a specific domain.
// @Tags domain
// @Produce json
// @Param slug path string true "Domain ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/dom/{slug}/groups [get]
// @Security Bearer
func (h *Handler) ListDomainGroupHandler(c *gin.Context) {
	// result, err := h.svc.ListDomainGroupsService(c, c.Param("slug"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer, nil)
}
