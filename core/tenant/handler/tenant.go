package handler

import (
	"ncobase/core/tenant/service"
	"ncobase/core/tenant/structs"

	"github.com/ncobase/ncore/pkg/ecode"
	"github.com/ncobase/ncore/pkg/helper"
	"github.com/ncobase/ncore/pkg/resp"

	"github.com/gin-gonic/gin"
)

// TenantHandlerInterface is the interface for the handler.
type TenantHandlerInterface interface {
	Create(c *gin.Context)
	UserOwn(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	GetMenus(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	ListAttachments(c *gin.Context)
	ListRoles(c *gin.Context)
	GetSetting(c *gin.Context)
	ListUsers(c *gin.Context)
	ListGroups(c *gin.Context)
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

// Create handles creating a tenant.
//
// @Summary Create tenant
// @Description Create a new tenant.
// @Tags iam
// @Accept json
// @Produce json
// @Param body body structs.CreateTenantBody true "CreateTenantBody object"
// @Success 200 {object} structs.ReadTenant "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants [post]
// @Security Bearer
func (h *TenantHandler) Create(c *gin.Context) {
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

// UserOwn handles reading a user's tenant.
//
// @Summary Get user owned tenant
// @Description Retrieve the tenant associated with the specified user.
// @Tags iam
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} structs.ReadTenant "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/users/{username}/tenant [get]
// @Security Bearer
func (h *TenantHandler) UserOwn(c *gin.Context) {
	result, err := h.s.Tenant.UserOwn(c.Request.Context(), c.Param("username"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Update handles updating a tenant.
//
// @Summary Update tenant
// @Description Update the tenant information.
// @Tags iam
// @Accept json
// @Produce json
// @Param slug path string true "Tenant ID"
// @Param body body structs.UpdateTenantBody true "UpdateTenantBody object"
// @Success 200 {object} structs.ReadTenant "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants [put]
// @Security Bearer
func (h *TenantHandler) Update(c *gin.Context) {
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

// Get handles reading tenant information.
//
// @Summary Get tenant
// @Description Retrieve information about a specific tenant.
// @Tags iam
// @Produce json
// @Param slug path string true "Tenant ID"
// @Success 200 {object} structs.ReadTenant "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/{slug} [get]
// @Security Bearer
func (h *TenantHandler) Get(c *gin.Context) {
	result, err := h.s.Tenant.Get(c.Request.Context(), c.Param("slug"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetMenus handles reading tenant menu.
//
// @Summary Get tenant menu
// @Description Retrieve the menu associated with a specific tenant.
// @Tags iam
// @Produce json
// @Param slug path string true "Tenant ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/{slug}/menu [get]
// @Security Bearer
func (h *TenantHandler) GetMenus(c *gin.Context) {
	result, err := h.s.Tenant.Get(c.Request.Context(), c.Param("slug"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetTenantSetting handles reading tenant setting.
//
// @Summary Get tenant setting
// @Description Retrieve the settings associated with a specific tenant.
// @Tags iam
// @Produce json
// @Param slug path string true "Tenant ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/{slug}/setting [get]
// @Security Bearer
func (h *TenantHandler) GetTenantSetting(c *gin.Context) {
	result, err := h.s.Tenant.Get(c.Request.Context(), c.Param("slug"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Delete handles deleting a tenant.
//
// @Summary Delete tenant
// @Description Delete a specific tenant.
// @Tags iam
// @Produce json
// @Param slug path string true "Tenant ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/{slug} [delete]
// @Security Bearer
func (h *TenantHandler) Delete(c *gin.Context) {
	if err := h.s.Tenant.Delete(c.Request.Context(), c.Param("slug")); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// List handles listing tenants.
//
// @Summary List tenants
// @Description Retrieve a list of tenants.
// @Tags iam
// @Produce json
// @Param params query structs.ListTenantParams true "List tenant parameters"
// @Success 200 {array} structs.ReadTenant"success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants [get]
// @Security Bearer
func (h *TenantHandler) List(c *gin.Context) {
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

// ListAttachments handles listing tenant attachments.
// TODO: implement this
// @Summary List tenant attachments
// @Description Retrieve a list of attachments associated with a specific tenant.
// @Tags iam
// @Produce json
// @Param slug path string true "Tenant ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/{slug}/attachments [get]
// @Security Bearer
func (h *TenantHandler) ListAttachments(c *gin.Context) {
	// result, err := h.svc.ListAttachmentss(c.Request.Context(),c.Param("slug"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer)
}

// ListRoles handles listing tenant roles.
// TODO: implement this
// @Summary List tenant roles
// @Description Retrieve a list of roles associated with a specific tenant.
// @Tags iam
// @Produce json
// @Param slug path string true "Tenant ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/{slug}/roles [get]
// @Security Bearer
func (h *TenantHandler) ListRoles(c *gin.Context) {
	// result, err := h.svc.ListRoless(c.Request.Context(),c.Param("slug"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer)
}

// GetSetting handles listing tenant settings.
// TODO: implement this
// @Summary List tenant settings
// @Description Retrieve a list of settings associated with a specific tenant.
// @Tags iam
// @Produce json
// @Param slug path string true "Tenant ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/{slug}/settings [get]
// @Security Bearer
func (h *TenantHandler) GetSetting(c *gin.Context) {
	// result, err := h.svc.GetSettings(c.Request.Context(),c.Param("slug"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer)
}

// ListUsers handles listing tenant users.
// TODO: implement this
// @Summary List tenant users
// @Description Retrieve a list of users associated with a specific tenant.
// @Tags iam
// @Produce json
// @Param slug path string true "Tenant ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/{slug}/users [get]
// @Security Bearer
func (h *TenantHandler) ListUsers(c *gin.Context) {
	// result, err := h.svc.ListUserss(c.Request.Context(),c.Param("slug"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer)
}

// ListGroups handles listing tenant groups.
// TODO: implement this
// @Summary List tenant groups
// @Description Retrieve a list of groups associated with a specific tenant.
// @Tags iam
// @Produce json
// @Param slug path string true "Tenant ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/{slug}/groups [get]
// @Security Bearer
func (h *TenantHandler) ListGroups(c *gin.Context) {
	// result, err := h.svc.ListGroupss(c.Request.Context(),c.Param("slug"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer)
}
