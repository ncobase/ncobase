package handler

import (
	"ncobase/tenant/service"
	"ncobase/tenant/structs"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// TenantSettingHandlerInterface defines the interface for tenant setting handler
type TenantSettingHandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	BulkUpdate(c *gin.Context)
	GetTenantSettings(c *gin.Context)
	GetPublicSettings(c *gin.Context)
	SetSetting(c *gin.Context)
	GetSetting(c *gin.Context)
}

// tenantSettingHandler implements TenantSettingHandlerInterface
type tenantSettingHandler struct {
	s *service.Service
}

// NewTenantSettingHandler creates a new tenant setting handler
func NewTenantSettingHandler(svc *service.Service) TenantSettingHandlerInterface {
	return &tenantSettingHandler{s: svc}
}

// Create handles creating a tenant setting
//
// @Summary Create tenant setting
// @Description Create a new tenant setting configuration
// @Tags iam
// @Accept json
// @Produce json
// @Param body body structs.CreateTenantSettingBody true "Setting configuration"
// @Success 200 {object} structs.ReadTenantSetting "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/settings [post]
// @Security Bearer
func (h *tenantSettingHandler) Create(c *gin.Context) {
	body := &structs.CreateTenantSettingBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.TenantSetting.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Update handles updating a tenant setting
//
// @Summary Update tenant setting
// @Description Update an existing tenant setting configuration
// @Tags iam
// @Accept json
// @Produce json
// @Param id path string true "Setting ID"
// @Param body body types.JSON true "Update data"
// @Success 200 {object} structs.ReadTenantSetting "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/settings/{id} [put]
// @Security Bearer
func (h *tenantSettingHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	updates := &types.JSON{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, updates); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.TenantSetting.Update(c.Request.Context(), id, *updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get handles retrieving a tenant setting
//
// @Summary Get tenant setting
// @Description Retrieve a tenant setting by ID
// @Tags iam
// @Produce json
// @Param id path string true "Setting ID"
// @Success 200 {object} structs.ReadTenantSetting "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/settings/{id} [get]
// @Security Bearer
func (h *tenantSettingHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.s.TenantSetting.Get(c.Request.Context(), id)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete handles deleting a tenant setting
//
// @Summary Delete tenant setting
// @Description Delete a tenant setting configuration
// @Tags iam
// @Produce json
// @Param id path string true "Setting ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/settings/{id} [delete]
// @Security Bearer
func (h *tenantSettingHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.s.TenantSetting.Delete(c.Request.Context(), id); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List handles listing tenant settings
//
// @Summary List tenant settings
// @Description Retrieve a list of tenant settings
// @Tags iam
// @Produce json
// @Param params query structs.ListTenantSettingParams true "List parameters"
// @Success 200 {array} structs.ReadTenantSetting "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/settings [get]
// @Security Bearer
func (h *tenantSettingHandler) List(c *gin.Context) {
	params := &structs.ListTenantSettingParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.TenantSetting.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// BulkUpdate handles bulk updating tenant settings
//
// @Summary Bulk update tenant settings
// @Description Update multiple tenant settings at once
// @Tags iam
// @Accept json
// @Produce json
// @Param body body structs.BulkUpdateSettingsRequest true "Bulk update request"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/settings/bulk [post]
// @Security Bearer
func (h *tenantSettingHandler) BulkUpdate(c *gin.Context) {
	body := &structs.BulkUpdateSettingsRequest{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	if err := h.s.TenantSetting.BulkUpdate(c.Request.Context(), body); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// GetTenantSettings handles retrieving all settings for a tenant
//
// @Summary Get all tenant settings
// @Description Retrieve all settings for a specific tenant
// @Tags iam
// @Produce json
// @Param slug path string true "Tenant ID"
// @Success 200 {object} map[string]interface{} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/{slug}/settings [get]
// @Security Bearer
func (h *tenantSettingHandler) GetTenantSettings(c *gin.Context) {
	tenantID := c.Param("slug")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug or tenant_id")))
		return
	}

	result, err := h.s.TenantSetting.GetTenantSettings(c.Request.Context(), tenantID, false)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetPublicSettings handles retrieving public settings for a tenant
//
// @Summary Get public tenant settings
// @Description Retrieve public settings for a specific tenant
// @Tags iam
// @Produce json
// @Param slug path string true "Tenant ID"
// @Success 200 {object} map[string]interface{} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/{slug}/settings/public [get]
func (h *tenantSettingHandler) GetPublicSettings(c *gin.Context) {
	tenantID := c.Param("slug")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug or tenant_id")))
		return
	}

	result, err := h.s.TenantSetting.GetTenantSettings(c.Request.Context(), tenantID, true)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// SetSetting handles setting a specific tenant setting
//
// @Summary Set tenant setting
// @Description Set a specific setting for a tenant
// @Tags iam
// @Accept json
// @Produce json
// @Param slug path string true "Tenant ID"
// @Param key path string true "Setting Key"
// @Param body body map[string]string true "Setting value"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/{slug}/settings/{key} [put]
// @Security Bearer
func (h *tenantSettingHandler) SetSetting(c *gin.Context) {
	tenantID := c.Param("slug")
	key := c.Param("key")
	if tenantID == "" || key == "" {
		resp.Fail(c.Writer, resp.BadRequest("Missing required parameters"))
		return
	}

	var body map[string]string
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	value, ok := body["value"]
	if !ok {
		resp.Fail(c.Writer, resp.BadRequest("Missing value field"))
		return
	}

	if err := h.s.TenantSetting.SetSetting(c.Request.Context(), tenantID, key, value); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// GetSetting handles retrieving a specific tenant setting
//
// @Summary Get specific tenant setting
// @Description Retrieve a specific setting for a tenant
// @Tags iam
// @Produce json
// @Param slug path string true "Tenant ID"
// @Param key path string true "Setting Key"
// @Success 200 {object} map[string]interface{} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/{slug}/settings/{key} [get]
// @Security Bearer
func (h *tenantSettingHandler) GetSetting(c *gin.Context) {
	tenantID := c.Param("slug")
	key := c.Param("key")
	if tenantID == "" || key == "" {
		resp.Fail(c.Writer, resp.BadRequest("Missing required parameters"))
		return
	}

	value, err := h.s.TenantSetting.GetSettingValue(c.Request.Context(), tenantID, key)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]interface{}{
		"key":   key,
		"value": value,
	})
}
