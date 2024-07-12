package handler

import (
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/feature/tenant/service"
	"ncobase/feature/tenant/structs"
	"ncobase/helper"

	"github.com/gin-gonic/gin"
)

// TenantHandlerInterface is the interface for the handler.
type TenantHandlerInterface interface {
	AccountTenantHandler(c *gin.Context)
	AccountTenantsHandler(c *gin.Context)
	CreateTenantHandler(c *gin.Context)
	UserTenantHandler(c *gin.Context)
	UpdateTenantHandler(c *gin.Context)
	GetTenantHandler(c *gin.Context)
	GetTenantMenuHandler(c *gin.Context)
	DeleteTenantHandler(c *gin.Context)
	ListTenantHandler(c *gin.Context)
	ListTenantAssetHandler(c *gin.Context)
	ListTenantRoleHandler(c *gin.Context)
	ListTenantModuleHandler(c *gin.Context)
	ListTenantSettingHandler(c *gin.Context)
	ListTenantUserHandler(c *gin.Context)
	ListTenantGroupHandler(c *gin.Context)
}

// TenantHandler represents the handler.
type TenantHandler struct {
	s *service.Service
}

// NewTenantHandler creates a new handler.
func NewTenantHandler(svc *service.Service) TenantHandlerInterface {
	return &TenantHandler{
		s: svc,
	}
}

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
func (h *TenantHandler) AccountTenantHandler(c *gin.Context) {
	result, err := h.s.Tenant.Account(c.Request.Context())
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
func (h *TenantHandler) AccountTenantsHandler(c *gin.Context) {
	result, err := h.s.Tenant.AccountTenants(c.Request.Context())
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
func (h *TenantHandler) CreateTenantHandler(c *gin.Context) {
	body := &structs.CreateTenantBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Tenant.Create(c.Request.Context(), body)
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
func (h *TenantHandler) UserTenantHandler(c *gin.Context) {
	result, err := h.s.Tenant.UserOwn(c.Request.Context(), c.Param("username"))
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
func (h *TenantHandler) UpdateTenantHandler(c *gin.Context) {
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

	result, err := h.s.Tenant.Update(c.Request.Context(), body)
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
func (h *TenantHandler) GetTenantHandler(c *gin.Context) {
	result, err := h.s.Tenant.Get(c.Request.Context(), c.Param("slug"))
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
func (h *TenantHandler) GetTenantMenuHandler(c *gin.Context) {
	result, err := h.s.Tenant.Get(c.Request.Context(), c.Param("slug"))
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
func (h *TenantHandler) GetTenantSettingHandler(c *gin.Context) {
	result, err := h.s.Tenant.Get(c.Request.Context(), c.Param("slug"))
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
func (h *TenantHandler) DeleteTenantHandler(c *gin.Context) {
	if err := h.s.Tenant.Delete(c.Request.Context(), c.Param("slug")); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
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
func (h *TenantHandler) ListTenantHandler(c *gin.Context) {
	params := &structs.ListTenantParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Tenant.List(c.Request.Context(), params)
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
func (h *TenantHandler) ListTenantAssetHandler(c *gin.Context) {
	// result, err := h.svc.ListTenantAssetsService(c.Request.Context(),c.Param("slug"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer)
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
func (h *TenantHandler) ListTenantRoleHandler(c *gin.Context) {
	// result, err := h.svc.ListTenantRolesService(c.Request.Context(),c.Param("slug"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer)
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
func (h *TenantHandler) ListTenantModuleHandler(c *gin.Context) {
	// result, err := h.svc.ListTenantModulesService(c.Request.Context(),c.Param("slug"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer)
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
func (h *TenantHandler) ListTenantSettingHandler(c *gin.Context) {
	// result, err := h.svc.ListTenantSettingsService(c.Request.Context(),c.Param("slug"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer)
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
func (h *TenantHandler) ListTenantUserHandler(c *gin.Context) {
	// result, err := h.svc.ListTenantUsersService(c.Request.Context(),c.Param("slug"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer)
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
func (h *TenantHandler) ListTenantGroupHandler(c *gin.Context) {
	// result, err := h.svc.ListTenantGroupsService(c.Request.Context(),c.Param("slug"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer)
}
