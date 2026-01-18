package handler

import (
	"ncobase/core/space/service"
	"ncobase/core/space/structs"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// SpaceSettingHandlerInterface defines the interface for space setting handler
type SpaceSettingHandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	BulkUpdate(c *gin.Context)
	GetSpaceSettings(c *gin.Context)
	GetPublicSettings(c *gin.Context)
	SetSetting(c *gin.Context)
	GetSetting(c *gin.Context)
}

// spaceSettingHandler implements SpaceSettingHandlerInterface
type spaceSettingHandler struct {
	s *service.Service
}

// NewSpaceSettingHandler creates a new space setting handler
func NewSpaceSettingHandler(svc *service.Service) SpaceSettingHandlerInterface {
	return &spaceSettingHandler{s: svc}
}

// Create handles creating a space setting
//
// @Summary Create space setting
// @Description Create a new space setting configuration
// @Tags sys
// @Accept json
// @Produce json
// @Param body body structs.CreateSpaceSettingBody true "Setting configuration"
// @Success 200 {object} structs.ReadSpaceSetting "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/settings [post]
// @Security Bearer
func (h *spaceSettingHandler) Create(c *gin.Context) {
	body := &structs.CreateSpaceSettingBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.SpaceSetting.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Update handles updating a space setting
//
// @Summary Update space setting
// @Description Update an existing space setting configuration
// @Tags sys
// @Accept json
// @Produce json
// @Param id path string true "Setting ID"
// @Param body body types.JSON true "Update data"
// @Success 200 {object} structs.ReadSpaceSetting "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/settings/{id} [put]
// @Security Bearer
func (h *spaceSettingHandler) Update(c *gin.Context) {
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

	result, err := h.s.SpaceSetting.Update(c.Request.Context(), id, *updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get handles retrieving a space setting
//
// @Summary Get space setting
// @Description Retrieve a space setting by ID
// @Tags sys
// @Produce json
// @Param id path string true "Setting ID"
// @Success 200 {object} structs.ReadSpaceSetting "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/settings/{id} [get]
// @Security Bearer
func (h *spaceSettingHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.s.SpaceSetting.Get(c.Request.Context(), id)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete handles deleting a space setting
//
// @Summary Delete space setting
// @Description Delete a space setting configuration
// @Tags sys
// @Produce json
// @Param id path string true "Setting ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/settings/{id} [delete]
// @Security Bearer
func (h *spaceSettingHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.s.SpaceSetting.Delete(c.Request.Context(), id); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List handles listing space settings
//
// @Summary List space settings
// @Description Retrieve a list of space settings
// @Tags sys
// @Produce json
// @Param params query structs.ListSpaceSettingParams true "List parameters"
// @Success 200 {array} structs.ReadSpaceSetting "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/settings [get]
// @Security Bearer
func (h *spaceSettingHandler) List(c *gin.Context) {
	params := &structs.ListSpaceSettingParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.SpaceSetting.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// BulkUpdate handles bulk updating space settings
//
// @Summary Bulk update space settings
// @Description Update multiple space settings at once
// @Tags sys
// @Accept json
// @Produce json
// @Param body body structs.BulkUpdateSettingsRequest true "Bulk update request"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/settings/bulk [post]
// @Security Bearer
func (h *spaceSettingHandler) BulkUpdate(c *gin.Context) {
	body := &structs.BulkUpdateSettingsRequest{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	if err := h.s.SpaceSetting.BulkUpdate(c.Request.Context(), body); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// GetSpaceSettings handles retrieving all settings for a space
//
// @Summary Get all space settings
// @Description Retrieve all settings for a specific space
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {object} map[string]interface{} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/settings [get]
// @Security Bearer
func (h *spaceSettingHandler) GetSpaceSettings(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug or space_id")))
		return
	}

	result, err := h.s.SpaceSetting.GetSpaceSettings(c.Request.Context(), spaceID, false)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetPublicSettings handles retrieving public settings for a space
//
// @Summary Get public space settings
// @Description Retrieve public settings for a specific space
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {object} map[string]interface{} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/settings/public [get]
func (h *spaceSettingHandler) GetPublicSettings(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug or space_id")))
		return
	}

	result, err := h.s.SpaceSetting.GetSpaceSettings(c.Request.Context(), spaceID, true)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// SetSetting handles setting a specific space setting
//
// @Summary Set space setting
// @Description Set a specific setting for a space
// @Tags sys
// @Accept json
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param key path string true "Setting Key"
// @Param body body map[string]string true "Setting value"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/settings/{key} [put]
// @Security Bearer
func (h *spaceSettingHandler) SetSetting(c *gin.Context) {
	spaceID := c.Param("spaceId")
	key := c.Param("key")
	if spaceID == "" || key == "" {
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

	if err := h.s.SpaceSetting.SetSetting(c.Request.Context(), spaceID, key, value); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// GetSetting handles retrieving a specific space setting
//
// @Summary Get specific space setting
// @Description Retrieve a specific setting for a space
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param key path string true "Setting Key"
// @Success 200 {object} map[string]interface{} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/settings/{key} [get]
// @Security Bearer
func (h *spaceSettingHandler) GetSetting(c *gin.Context) {
	spaceID := c.Param("spaceId")
	key := c.Param("key")
	if spaceID == "" || key == "" {
		resp.Fail(c.Writer, resp.BadRequest("Missing required parameters"))
		return
	}

	value, err := h.s.SpaceSetting.GetSettingValue(c.Request.Context(), spaceID, key)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]any{
		"key":   key,
		"value": value,
	})
}
