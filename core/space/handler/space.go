package handler

import (
	"ncobase/core/space/service"
	"ncobase/core/space/structs"
	resourceStructs "ncobase/plugin/resource/structs"
	"strings"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// SpaceHandlerInterface is the interface for the handler.
type SpaceHandlerInterface interface {
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
	ListOrganizations(c *gin.Context)
}

// SpaceHandler represents the handler.
type SpaceHandler struct {
	s *service.Service
}

// NewSpaceHandler creates a new handler.
func NewSpaceHandler(svc *service.Service) SpaceHandlerInterface {
	return &SpaceHandler{
		s: svc,
	}
}

// Create handles creating a space.
//
// @Summary Create space
// @Description Create a new space.
// @Tags sys
// @Accept json
// @Produce json
// @Param body body structs.CreateSpaceBody true "CreateSpaceBody object"
// @Success 200 {object} structs.ReadSpace "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces [post]
// @Security Bearer
func (h *SpaceHandler) Create(c *gin.Context) {
	body := &structs.CreateSpaceBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Space.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// UserOwn handles reading a user's space.
//
// @Summary Get user owned space
// @Description Retrieve the space associated with the specified user.
// @Tags sys
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} structs.ReadSpace "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/{username}/space [get]
// @Security Bearer
func (h *SpaceHandler) UserOwn(c *gin.Context) {
	result, err := h.s.Space.UserOwn(c.Request.Context(), c.Param("username"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Update handles updating a space.
//
// @Summary Update space
// @Description Update the space information.
// @Tags sys
// @Accept json
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param body body structs.UpdateSpaceBody true "UpdateSpaceBody object"
// @Success 200 {object} structs.ReadSpace "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId} [put]
// @Security Bearer
func (h *SpaceHandler) Update(c *gin.Context) {
	slug := c.Param("spaceId")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}
	body := &structs.UpdateSpaceBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Space.Update(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Get handles reading space information.
//
// @Summary Get space
// @Description Retrieve information about a specific space.
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {object} structs.ReadSpace "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId} [get]
// @Security Bearer
func (h *SpaceHandler) Get(c *gin.Context) {
	result, err := h.s.Space.Get(c.Request.Context(), c.Param("spaceId"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetMenus handles reading space menu.
//
// @Summary Get space menu
// @Description Retrieve the menu associated
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/menu [get]
// @Security Bearer
func (h *SpaceHandler) GetMenus(c *gin.Context) {
	result, err := h.s.Space.Get(c.Request.Context(), c.Param("spaceId"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetSpaceSetting handles reading space setting.
//
// @Summary Get space setting
// @Description Retrieve the settings associated
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/setting [get]
// @Security Bearer
func (h *SpaceHandler) GetSpaceSetting(c *gin.Context) {
	result, err := h.s.Space.Get(c.Request.Context(), c.Param("spaceId"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Delete handles deleting a space.
//
// @Summary Delete space
// @Description Delete a specific space.
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId} [delete]
// @Security Bearer
func (h *SpaceHandler) Delete(c *gin.Context) {
	if err := h.s.Space.Delete(c.Request.Context(), c.Param("spaceId")); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// List handles listing spaces.
//
// @Summary List spaces
// @Description Retrieve a list of spaces.
// @Tags sys
// @Produce json
// @Param params query structs.ListSpaceParams true "List space parameters"
// @Success 200 {array} structs.ReadSpace"success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces [get]
// @Security Bearer
func (h *SpaceHandler) List(c *gin.Context) {
	params := &structs.ListSpaceParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Space.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// ListAttachments handles listing space attachments.
// @Summary List space attachments
// @Description Retrieve a list of attachments associated
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/attachments [get]
// @Security Bearer
func (h *SpaceHandler) ListAttachments(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	// Ensure current user belongs to the space or is the owner
	if _, err := h.s.Space.Get(c.Request.Context(), spaceID); err != nil {
		resp.Fail(c.Writer, resp.Forbidden("Space access denied"))
		return
	}

	resourceWrapper := h.s.ResourceFileWrapper()
	if resourceWrapper == nil || !resourceWrapper.HasFileService() {
		resp.Fail(c.Writer, resp.ServiceUnavailable("Resource service not available"))
		return
	}

	params := &resourceStructs.ListFileParams{}
	if err := c.ShouldBindQuery(params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	params.OwnerID = spaceID
	params.User = strings.TrimSpace(params.User)

	if params.Limit <= 0 {
		params.Limit = 50
	}

	result, err := resourceWrapper.ListFiles(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// ListRoles handles listing space roles.
// @Summary List space roles
// @Description Retrieve a list of roles associated
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/roles [get]
// @Security Bearer
func (h *SpaceHandler) ListRoles(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	result, err := h.s.UserSpaceRole.ListSpaceRoleIDs(c.Request.Context(), spaceID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetSetting handles listing space settings.
// @Summary List space settings
// @Description Retrieve a list of settings associated
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/settings [get]
// @Security Bearer
func (h *SpaceHandler) GetSetting(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	result, err := h.s.SpaceSetting.GetSpaceSettings(c.Request.Context(), spaceID, false)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// ListUsers handles listing space users.
// @Summary List space users
// @Description Retrieve a list of users associated
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/users [get]
// @Security Bearer
func (h *SpaceHandler) ListUsers(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	params := &structs.ListSpaceUsersParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.UserSpaceRole.ListSpaceUsers(c.Request.Context(), spaceID, params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// ListOrganizations handles listing space orgs.
// @Summary List space orgs
// @Description Retrieve a list of orgs associated
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/orgs [get]
// @Security Bearer
func (h *SpaceHandler) ListOrganizations(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	params := &structs.ListOrganizationParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.SpaceOrganization.GetSpaceOrganizations(c.Request.Context(), spaceID, params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}
