package handler

import (
	"ncobase/core/space/service"
	"ncobase/core/space/structs"

	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// SpaceOrganizationHandlerInterface represents the space group handler interface.
type SpaceOrganizationHandlerInterface interface {
	AddGroupToSpace(c *gin.Context)
	RemoveGroupFromSpace(c *gin.Context)
	GetSpaceOrganizations(c *gin.Context)
	GetOrganizationSpaces(c *gin.Context)
	IsGroupInSpace(c *gin.Context)
}

// spaceGroupHandler represents the space group handler.
type spaceGroupHandler struct {
	s *service.Service
}

// NewSpaceOrganizationHandler creates new space group handler.
func NewSpaceOrganizationHandler(svc *service.Service) SpaceOrganizationHandlerInterface {
	return &spaceGroupHandler{
		s: svc,
	}
}

// AddGroupToSpace handles adding a organization to a space.
//
// @Summary Add group to space
// @Description Add a organization to a specific space
// @Tags sys
// @Accept json
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param body body structs.AddSpaceOrganizationRequest true "AddSpaceOrganizationRequest object"
// @Success 200 {object} structs.SpaceOrganizationRelation "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/orgs [post]
// @Security Bearer
func (h *spaceGroupHandler) AddGroupToSpace(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	var req structs.AddSpaceOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	// Check if group already exists in space
	exists, _ := h.s.SpaceOrganization.IsGroupInSpace(c.Request.Context(), spaceID, req.OrgID)
	if exists {
		resp.Fail(c.Writer, resp.BadRequest("Group already exists in this space"))
		return
	}

	// Add the group to space
	relation, err := h.s.SpaceOrganization.AddGroupToSpace(c.Request.Context(), spaceID, req.OrgID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, relation)
}

// RemoveGroupFromSpace handles removing a organization from a space.
//
// @Summary Remove group from space
// @Description Remove a organization from a specific space
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param orgId path string true "Organization ID"
// @Success 200 {object} resp.Exception"success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/orgs/{orgId} [delete]
// @Security Bearer
func (h *spaceGroupHandler) RemoveGroupFromSpace(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	orgID := c.Param("orgId")
	if orgID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Organization ID is required"))
		return
	}

	err := h.s.SpaceOrganization.RemoveGroupFromSpace(c.Request.Context(), spaceID, orgID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]bool{"success": true})
}

// GetSpaceOrganizations handles getting all orgs for a space.
//
// @Summary Get space orgs
// @Description Get all orgs belonging to a specific space
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param params query structs.ListOrganizationParams true "List group parameters"
// @Success 200 {array} structs.ReadOrganization "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/orgs [get]
// @Security Bearer
func (h *spaceGroupHandler) GetSpaceOrganizations(c *gin.Context) {
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

// GetOrganizationSpaces handles getting all spaces that have a specific group.
//
// @Summary Get group spaces
// @Description Get all spaces that have a specific group
// @Tags sys
// @Produce json
// @Param orgId path string true "Organization ID"
// @Success 200 {array} string "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/orgs/{orgId}/spaces [get]
// @Security Bearer
func (h *spaceGroupHandler) GetOrganizationSpaces(c *gin.Context) {
	orgID := c.Param("orgId")
	if orgID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Organization ID is required"))
		return
	}

	spaces, err := h.s.SpaceOrganization.GetOrganizationSpaces(c.Request.Context(), orgID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, spaces)
}

// IsGroupInSpace handles checking if a organization belongs to a space.
//
// @Summary Check if group is in space
// @Description Check if a organization belongs to a specific space
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param orgId path string true "Organization ID"
// @Success 200 {object} map[string]bool "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/orgs/{orgId}/check [get]
// @Security Bearer
func (h *spaceGroupHandler) IsGroupInSpace(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	orgID := c.Param("orgId")
	if orgID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Organization ID is required"))
		return
	}

	exists, err := h.s.SpaceOrganization.IsGroupInSpace(c.Request.Context(), spaceID, orgID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]bool{"exists": exists})
}
