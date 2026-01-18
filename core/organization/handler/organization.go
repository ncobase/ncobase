package handler

import (
	"ncobase/core/organization/service"
	"ncobase/core/organization/structs"

	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// OrganizationHandlerInterface represents the organization handler interface.
type OrganizationHandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	GetMembers(c *gin.Context)
	AddMember(c *gin.Context)
	UpdateMember(c *gin.Context)
	RemoveMember(c *gin.Context)
	IsUserMember(c *gin.Context)
	IsUserOwner(c *gin.Context)
	GetUserRole(c *gin.Context)
}

// organizationHandler represents the organization handler.
type organizationHandler struct {
	s *service.Service
}

// NewOrganizationHandler creates new organization handler.
func NewOrganizationHandler(svc *service.Service) OrganizationHandlerInterface {
	return &organizationHandler{
		s: svc,
	}
}

// Create handles creating a new organization.
//
// @Summary Create organization
// @Description Create a new organization.
// @Tags org
// @Accept json
// @Produce json
// @Param body body structs.OrganizationBody true "OrganizationBody object"
// @Success 200 {object} structs.ReadOrganization "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/orgs [post]
// @Security Bearer
func (h *organizationHandler) Create(c *gin.Context) {
	body := &structs.CreateOrganizationBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Organization.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Update handles updating an organization.
//
// @Summary Update organization
// @Description Update an existing organization.
// @Tags org
// @Accept json
// @Produce json
// @Param body body structs.UpdateOrganizationBody true "UpdateOrganizationBody object"
// @Success 200 {object} structs.ReadOrganization "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/orgs [put]
// @Security Bearer
func (h *organizationHandler) Update(c *gin.Context) {
	body := types.JSON{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Organization.Update(c.Request.Context(), body["id"].(string), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Get handles retrieving an organization by ID or slug.
//
// @Summary Get organization
// @Description Retrieve an organization by ID or slug.
// @Tags org
// @Produce json
// @Param orgId path string true "Organization ID or slug"
// @Param params query structs.FindOrganization true "FindOrganization parameters"
// @Success 200 {object} structs.ReadOrganization "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/orgs/{orgId} [get]
// @Security Bearer
func (h *organizationHandler) Get(c *gin.Context) {
	params := &structs.FindOrganization{Organization: c.Param("orgId")}

	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Organization.Get(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Delete handles deleting an organization by ID or slug.
//
// @Summary Delete organization
// @Description Delete an organization by ID or slug.
// @Tags org
// @Produce json
// @Param orgId path string true "Organization ID or slug"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/orgs/{orgId} [delete]
// @Security Bearer
func (h *organizationHandler) Delete(c *gin.Context) {
	params := &structs.FindOrganization{Organization: c.Param("orgId")}
	err := h.s.Organization.Delete(c.Request.Context(), params.Organization)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// List handles listing all orgs.
//
// @Summary List orgs
// @Description Retrieve a list or tree structure of orgs.
// @Tags org
// @Produce json
// @Param params query structs.ListOrganizationParams true "List organization parameters"
// @Success 200 {array} structs.ReadOrganization "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/orgs [get]
// @Security Bearer
func (h *organizationHandler) List(c *gin.Context) {
	params := &structs.ListOrganizationParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Organization.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetMembers handles getting members for an organization.
//
// @Summary Get organization members
// @Description Get all members of a specific organization
// @Tags org
// @Produce json
// @Param orgId path string true "Organization ID or Slug"
// @Success 200 {array} structs.OrganizationMember "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/orgs/{orgId}/members [get]
// @Security Bearer
func (h *organizationHandler) GetMembers(c *gin.Context) {
	organizationID := c.Param("orgId")
	if organizationID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Organization ID or Slug is required"))
		return
	}

	members, err := h.s.UserOrganization.GetOrganizationMembers(c.Request.Context(), organizationID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, members)
}

// AddMember handles adding a member to an organization.
//
// @Summary Add member to organization
// @Description Add a user to an organization with a specified role
// @Tags org
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID or Slug"
// @Param body body structs.AddMemberRequest true "User details to add"
// @Success 200 {object} structs.OrganizationMember "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/orgs/{orgId}/members [post]
// @Security Bearer
func (h *organizationHandler) AddMember(c *gin.Context) {
	organizationID := c.Param("orgId")
	if organizationID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Organization ID or Slug is required"))
		return
	}

	var req structs.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	// Validate role
	if !structs.IsValidUserRole(req.Role) {
		resp.Fail(c.Writer, resp.BadRequest("Invalid role provided"))
		return
	}

	// Resolve organization ID if slug was provided
	organization, err := h.s.Organization.Get(c.Request.Context(), &structs.FindOrganization{Organization: organizationID})
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Organization not found"))
		return
	}

	// Check if user already exists in the organization
	isMember, _ := h.s.UserOrganization.IsUserMember(c.Request.Context(), organization.ID, req.UserID)
	if isMember {
		resp.Fail(c.Writer, resp.BadRequest("User is already a member of this organization"))
		return
	}

	// Add the user to the organization
	member, err := h.s.UserOrganization.AddUserToOrganization(c.Request.Context(), req.UserID, organization.ID, req.Role)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, member)
}

// UpdateMember handles updating a member in an organization.
//
// @Summary Update organization member
// @Description Update a member's role in an organization
// @Tags org
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID or Slug"
// @Param userId path string true "User ID"
// @Param body body structs.UpdateMemberRequest true "Role update"
// @Success 200 {object} structs.OrganizationMember "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/orgs/{orgId}/members/{userId} [put]
// @Security Bearer
func (h *organizationHandler) UpdateMember(c *gin.Context) {
	organizationID := c.Param("orgId")
	if organizationID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Organization ID or Slug is required"))
		return
	}

	userID := c.Param("userId")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest("User ID is required"))
		return
	}

	var req structs.UpdateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	// Validate role
	if !structs.IsValidUserRole(req.Role) {
		resp.Fail(c.Writer, resp.BadRequest("Invalid role provided"))
		return
	}

	// Resolve organization ID if slug was provided
	organization, err := h.s.Organization.Get(c.Request.Context(), &structs.FindOrganization{Organization: organizationID})
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Organization not found"))
		return
	}

	// Check if this is the only owner and trying to change role
	if req.Role != structs.RoleOwner {
		isOwner, _ := h.s.UserOrganization.HasRole(c.Request.Context(), organization.ID, userID, structs.RoleOwner)
		if isOwner {
			// Check if there's only one owner
			owners, _ := h.s.UserOrganization.GetMembersByRole(c.Request.Context(), organization.ID, structs.RoleOwner)
			if len(owners) <= 1 {
				resp.Fail(c.Writer, resp.BadRequest("Cannot change role of the only owner"))
				return
			}
		}
	}

	// Check if user exists in the organization
	isMember, _ := h.s.UserOrganization.IsUserMember(c.Request.Context(), organization.ID, userID)
	if !isMember {
		resp.Fail(c.Writer, resp.BadRequest("User is not a member of this organization"))
		return
	}

	// Update the member's role
	err = h.s.UserOrganization.RemoveUserFromOrganization(c.Request.Context(), userID, organization.ID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	member, err := h.s.UserOrganization.AddUserToOrganization(c.Request.Context(), userID, organization.ID, req.Role)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, member)
}

// RemoveMember handles removing a member from an organization.
//
// @Summary Remove organization member
// @Description Remove a user from an organization
// @Tags org
// @Produce json
// @Param orgId path string true "Organization ID or Slug"
// @Param userId path string true "User ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/orgs/{orgId}/members/{userId} [delete]
// @Security Bearer
func (h *organizationHandler) RemoveMember(c *gin.Context) {
	organizationID := c.Param("orgId")
	if organizationID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Organization ID or Slug is required"))
		return
	}

	userID := c.Param("userId")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest("User ID is required"))
		return
	}

	// Resolve organization ID if slug was provided
	organization, err := h.s.Organization.Get(c.Request.Context(), &structs.FindOrganization{Organization: organizationID})
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Organization not found"))
		return
	}

	// Check if this is the only owner
	isOwner, _ := h.s.UserOrganization.HasRole(c.Request.Context(), organization.ID, userID, structs.RoleOwner)
	if isOwner {
		// Check if there's only one owner
		owners, _ := h.s.UserOrganization.GetMembersByRole(c.Request.Context(), organization.ID, structs.RoleOwner)
		if len(owners) <= 1 {
			resp.Fail(c.Writer, resp.BadRequest("Cannot remove the only owner"))
			return
		}
	}

	err = h.s.UserOrganization.RemoveUserFromOrganization(c.Request.Context(), userID, organization.ID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]bool{"success": true})
}

// IsUserMember handles checking if a user is a member of an organization.
//
// @Summary Check if user is a member
// @Description Check if a user is a member of an organization
// @Tags org
// @Produce json
// @Param orgId path string true "Organization ID or Slug"
// @Param userId path string true "User ID"
// @Success 200 {object} map[string]bool "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/orgs/{orgId}/members/{userId}/check [get]
// @Security Bearer
func (h *organizationHandler) IsUserMember(c *gin.Context) {
	organizationID := c.Param("orgId")
	if organizationID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Organization ID or Slug is required"))
		return
	}

	userID := c.Param("userId")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest("User ID is required"))
		return
	}

	// Resolve organization ID if slug was provided
	organization, err := h.s.Organization.Get(c.Request.Context(), &structs.FindOrganization{Organization: organizationID})
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Organization not found"))
		return
	}

	isMember, err := h.s.UserOrganization.IsUserMember(c.Request.Context(), organization.ID, userID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]bool{"isMember": isMember})
}

// IsUserOwner handles checking if a user is an owner of an organization.
//
// @Summary Check if user is an owner
// @Description Check if a user has owner role in an organization
// @Tags org
// @Produce json
// @Param orgId path string true "Organization ID or Slug"
// @Param userId path string true "User ID"
// @Success 200 {object} map[string]bool "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/orgs/{orgId}/members/{userId}/is-owner [get]
// @Security Bearer
func (h *organizationHandler) IsUserOwner(c *gin.Context) {
	organizationID := c.Param("orgId")
	if organizationID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Organization ID or Slug is required"))
		return
	}

	userID := c.Param("userId")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest("User ID is required"))
		return
	}

	// Resolve organization ID if slug was provided
	organization, err := h.s.Organization.Get(c.Request.Context(), &structs.FindOrganization{Organization: organizationID})
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Organization not found"))
		return
	}

	isOwner, err := h.s.UserOrganization.HasRole(c.Request.Context(), organization.ID, userID, structs.RoleOwner)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]bool{"isOwner": isOwner})
}

// GetUserRole handles getting a user's role in an organization.
//
// @Summary Get user role
// @Description Get a user's role in an organization
// @Tags org
// @Produce json
// @Param orgId path string true "Organization ID or Slug"
// @Param userId path string true "User ID"
// @Success 200 {object} map[string]string "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/orgs/{orgId}/members/{userId}/role [get]
// @Security Bearer
func (h *organizationHandler) GetUserRole(c *gin.Context) {
	organizationID := c.Param("orgId")
	if organizationID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Organization ID or Slug is required"))
		return
	}

	userID := c.Param("userId")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest("User ID is required"))
		return
	}

	// Resolve organization ID if slug was provided
	organization, err := h.s.Organization.Get(c.Request.Context(), &structs.FindOrganization{Organization: organizationID})
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Organization not found"))
		return
	}

	role, err := h.s.UserOrganization.GetUserRole(c.Request.Context(), organization.ID, userID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]string{"role": string(role)})
}
