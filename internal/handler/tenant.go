package handler

import (
	"ncobase/internal/data/structs"
	"ncobase/internal/helper"

	"ncobase/common/ecode"
	"ncobase/common/resp"

	"github.com/gin-gonic/gin"
)

// AccountTenantHandler handles reading the current user's tenant.
//
// @Summary Get current user tenant
// @Description Retrieve the tenant associated with the current user.
// @Tags account
// @Produce json
// @Success 200 {object} structs.ReadTenant "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/account/tenant [get]
// @Security Bearer
func (h *Handler) AccountTenantHandler(c *gin.Context) {
	result, err := h.svc.AccountTenantService(c)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// AccountTenantsHandler handles reading the current user's tenants.
//
// @Summary Get current user tenants
// @Description Retrieve the tenant associated with the current user.
// @Tags account
// @Produce json
// @Success 200 {object} structs.ReadTenant "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/account/tenants [get]
// @Security Bearer
func (h *Handler) AccountTenantsHandler(c *gin.Context) {
	result, err := h.svc.AccountTenantsService(c)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// CreateTenantHandler handles creating a tenant.
//
// @Summary Create tenant
// @Description Create a new tenant.
// @Tags tenant
// @Accept json
// @Produce json
// @Param body body structs.CreateTenantBody true "CreateTenantBody object"
// @Success 200 {object} structs.ReadTenant "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/tenants [post]
// @Security Bearer
func (h *Handler) CreateTenantHandler(c *gin.Context) {
	body := &structs.CreateTenantBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.svc.CreateTenantService(c, body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// UserTenantHandler handles reading a user's tenant.
//
// @Summary Get user owned tenant
// @Description Retrieve the tenant associated with the specified user.
// @Tags user
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} structs.ReadTenant "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/users/{username}/tenant [get]
// @Security Bearer
func (h *Handler) UserTenantHandler(c *gin.Context) {
	result, err := h.svc.UserTenantService(c, c.Param("username"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// UpdateTenantHandler handles updating a tenant.
//
// @Summary Update tenant
// @Description Update the tenant information.
// @Tags tenant
// @Accept json
// @Produce json
// @Param slug path string true "Tenant ID"
// @Param body body structs.UpdateTenantBody true "UpdateTenantBody object"
// @Success 200 {object} structs.ReadTenant "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/tenants [put]
// @Security Bearer
func (h *Handler) UpdateTenantHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}
	body := &structs.UpdateTenantBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.svc.UpdateTenantService(c, body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetTenantHandler handles reading tenant information.
//
// @Summary Get tenant
// @Description Retrieve information about a specific tenant.
// @Tags tenant
// @Produce json
// @Param slug path string true "Tenant ID"
// @Success 200 {object} structs.ReadTenant "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/tenants/{slug} [get]
// @Security Bearer
func (h *Handler) GetTenantHandler(c *gin.Context) {
	result, err := h.svc.GetTenantService(c, c.Param("slug"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetTenantMenuHandler handles reading tenant menu.
//
// @Summary Get tenant menu
// @Description Retrieve the menu associated with a specific tenant.
// @Tags tenant
// @Produce json
// @Param slug path string true "Tenant ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/tenants/{slug}/menu [get]
// @Security Bearer
func (h *Handler) GetTenantMenuHandler(c *gin.Context) {
	result, err := h.svc.GetTenantService(c, c.Param("slug"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetTenantSettingHandler handles reading tenant setting.
//
// @Summary Get tenant setting
// @Description Retrieve the settings associated with a specific tenant.
// @Tags tenant
// @Produce json
// @Param slug path string true "Tenant ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/tenants/{slug}/setting [get]
// @Security Bearer
func (h *Handler) GetTenantSettingHandler(c *gin.Context) {
	result, err := h.svc.GetTenantService(c, c.Param("slug"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// DeleteTenantHandler handles deleting a tenant.
//
// @Summary Delete tenant
// @Description Delete a specific tenant.
// @Tags tenant
// @Produce json
// @Param slug path string true "Tenant ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/tenants/{slug} [delete]
// @Security Bearer
func (h *Handler) DeleteTenantHandler(c *gin.Context) {
	result, err := h.svc.DeleteTenantService(c, c.Param("slug"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// ListTenantHandler handles listing tenants.
//
// @Summary List tenants
// @Description Retrieve a list of tenants.
// @Tags tenant
// @Produce json
// @Param params query structs.ListTenantParams true "List tenant parameters"
// @Success 200 {array} structs.ReadTenant"success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/tenants [get]
// @Security Bearer
func (h *Handler) ListTenantHandler(c *gin.Context) {
	params := &structs.ListTenantParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.svc.ListTenantsService(c, params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// ListTenantAssetHandler handles listing tenant assets.
// TODO: implement this
// @Summary List tenant assets
// @Description Retrieve a list of assets associated with a specific tenant.
// @Tags tenant
// @Produce json
// @Param slug path string true "Tenant ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/tenants/{slug}/assets [get]
// @Security Bearer
func (h *Handler) ListTenantAssetHandler(c *gin.Context) {
	// result, err := h.svc.ListTenantAssetsService(c, c.Param("slug"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer, nil)
}

// ListTenantRoleHandler handles listing tenant roles.
// TODO: implement this
// @Summary List tenant roles
// @Description Retrieve a list of roles associated with a specific tenant.
// @Tags tenant
// @Produce json
// @Param slug path string true "Tenant ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/tenants/{slug}/roles [get]
// @Security Bearer
func (h *Handler) ListTenantRoleHandler(c *gin.Context) {
	// result, err := h.svc.ListTenantRolesService(c, c.Param("slug"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer, nil)
}

// ListTenantModuleHandler handles listing tenant modules.
// TODO: implement this
// @Summary List tenant modules
// @Description Retrieve a list of modules associated with a specific tenant.
// @Tags tenant
// @Produce json
// @Param slug path string true "Tenant ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/tenants/{slug}/modules [get]
// @Security Bearer
func (h *Handler) ListTenantModuleHandler(c *gin.Context) {
	// result, err := h.svc.ListTenantModulesService(c, c.Param("slug"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer, nil)
}

// ListTenantSettingHandler handles listing tenant settings.
// TODO: implement this
// @Summary List tenant settings
// @Description Retrieve a list of settings associated with a specific tenant.
// @Tags tenant
// @Produce json
// @Param slug path string true "Tenant ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/tenants/{slug}/settings [get]
// @Security Bearer
func (h *Handler) ListTenantSettingHandler(c *gin.Context) {
	// result, err := h.svc.ListTenantSettingsService(c, c.Param("slug"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer, nil)
}

// ListTenantUserHandler handles listing tenant users.
// TODO: implement this
// @Summary List tenant users
// @Description Retrieve a list of users associated with a specific tenant.
// @Tags tenant
// @Produce json
// @Param slug path string true "Tenant ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/tenants/{slug}/users [get]
// @Security Bearer
func (h *Handler) ListTenantUserHandler(c *gin.Context) {
	// result, err := h.svc.ListTenantUsersService(c, c.Param("slug"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer, nil)
}

// ListTenantGroupHandler handles listing tenant groups.
// TODO: implement this
// @Summary List tenant groups
// @Description Retrieve a list of groups associated with a specific tenant.
// @Tags tenant
// @Produce json
// @Param slug path string true "Tenant ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/tenants/{slug}/groups [get]
// @Security Bearer
func (h *Handler) ListTenantGroupHandler(c *gin.Context) {
	// result, err := h.svc.ListTenantGroupsService(c, c.Param("slug"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer, nil)
}
