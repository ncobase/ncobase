package handler

import (
	"ncobase/space/service"
	"ncobase/space/structs"

	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// GroupHandlerInterface represents the group handler interface.
type GroupHandlerInterface interface {
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

// groupHandler represents the group handler.
type groupHandler struct {
	s *service.Service
}

// NewGroupHandler creates new group handler.
func NewGroupHandler(svc *service.Service) GroupHandlerInterface {
	return &groupHandler{
		s: svc,
	}
}

// Create handles creating a new group.
//
// @Summary Create group
// @Description Create a new group.
// @Tags org
// @Accept json
// @Produce json
// @Param body body structs.GroupBody true "GroupBody object"
// @Success 200 {object} structs.ReadGroup "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /org/groups [post]
// @Security Bearer
func (h *groupHandler) Create(c *gin.Context) {
	body := &structs.CreateGroupBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Group.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Update handles updating a group.
//
// @Summary Update group
// @Description Update an existing group.
// @Tags org
// @Accept json
// @Produce json
// @Param body body structs.UpdateGroupBody true "UpdateGroupBody object"
// @Success 200 {object} structs.ReadGroup "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /org/groups [put]
// @Security Bearer
func (h *groupHandler) Update(c *gin.Context) {
	body := types.JSON{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Group.Update(c.Request.Context(), body["id"].(string), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Get handles retrieving a group by ID or slug.
//
// @Summary Get group
// @Description Retrieve a group by ID or slug.
// @Tags org
// @Produce json
// @Param slug path string true "Group ID or slug"
// @Param params query structs.FindGroup true "FindGroup parameters"
// @Success 200 {object} structs.ReadGroup "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /org/groups/{slug} [get]
// @Security Bearer
func (h *groupHandler) Get(c *gin.Context) {
	params := &structs.FindGroup{Group: c.Param("slug")}

	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Group.Get(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Delete handles deleting a group by ID or slug.
//
// @Summary Delete group
// @Description Delete a group by ID or slug.
// @Tags org
// @Produce json
// @Param slug path string true "Group ID or slug"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /org/groups/{slug} [delete]
// @Security Bearer
func (h *groupHandler) Delete(c *gin.Context) {
	params := &structs.FindGroup{Group: c.Param("slug")}
	err := h.s.Group.Delete(c.Request.Context(), params.Group)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// List handles listing all groups.
//
// @Summary List groups
// @Description Retrieve a list or tree structure of groups.
// @Tags org
// @Produce json
// @Param params query structs.ListGroupParams true "List group parameters"
// @Success 200 {array} structs.ReadGroup "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /org/groups [get]
// @Security Bearer
func (h *groupHandler) List(c *gin.Context) {
	params := &structs.ListGroupParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Group.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// // GetTree handles retrieving the group tree.
// //
// // @Summary Get group tree
// // @Description Retrieve the group tree structure.
// // @Tags org
// // @Produce json
// // @Param params query structs.FindGroup true "FindGroup parameters"
// // @Success 200 {object} structs.ReadGroup "success"
// // @Failure 400 {object} resp.Exception "bad request"
// // @Router /org/groups/tree [get]
// // @Security Bearer
// func (h *Handler) GetTree(c *gin.Context) {
// 	params := &structs.FindGroup{}
// 	if validationErrors, err := helper.ShouldBindAndValidateStruct(c,params); err != nil {
// 		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
// 		return
// 	} else if len(validationErrors) > 0 {
// 		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
// 		return
// 	}
//
// 	result, err := h.s.Group.GetTree(c.Request.Context(),params)
// 	if err != nil {
// 		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
// 		return
// 	}
// 	resp.Success(c.Writer, result)
// }

// GetMembers handles getting members for a group.
//
// @Summary Get group members
// @Description Get all members of a specific group
// @Tags org
// @Produce json
// @Param slug path string true "Group ID or Slug"
// @Success 200 {array} structs.GroupMember "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /org/groups/{slug}/members [get]
// @Security Bearer
func (h *groupHandler) GetMembers(c *gin.Context) {
	groupID := c.Param("slug")
	if groupID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Group ID or Slug is required"))
		return
	}

	members, err := h.s.UserGroup.GetGroupMembers(c.Request.Context(), groupID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, members)
}

// AddMember handles adding a member to a group.
//
// @Summary Add member to group
// @Description Add a user to a group with a specified role
// @Tags org
// @Accept json
// @Produce json
// @Param slug path string true "Group ID or Slug"
// @Param body body structs.AddMemberRequest true "User details to add"
// @Success 200 {object} structs.GroupMember "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /org/groups/{slug}/members [post]
// @Security Bearer
func (h *groupHandler) AddMember(c *gin.Context) {
	groupID := c.Param("slug")
	if groupID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Group ID or Slug is required"))
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

	// Resolve group ID if slug was provided
	group, err := h.s.Group.Get(c.Request.Context(), &structs.FindGroup{Group: groupID})
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Group not found"))
		return
	}

	// Check if user already exists in the group
	isMember, _ := h.s.UserGroup.IsUserMember(c.Request.Context(), group.ID, req.UserID)
	if isMember {
		resp.Fail(c.Writer, resp.BadRequest("User is already a member of this group"))
		return
	}

	// Add the user to the group
	member, err := h.s.UserGroup.AddUserToGroup(c.Request.Context(), req.UserID, group.ID, req.Role)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, member)
}

// UpdateMember handles updating a member in a group.
//
// @Summary Update group member
// @Description Update a member's role in a group
// @Tags org
// @Accept json
// @Produce json
// @Param slug path string true "Group ID or Slug"
// @Param userId path string true "User ID"
// @Param body body structs.UpdateMemberRequest true "Role update"
// @Success 200 {object} structs.GroupMember "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /org/groups/{slug}/members/{userId} [put]
// @Security Bearer
func (h *groupHandler) UpdateMember(c *gin.Context) {
	groupID := c.Param("slug")
	if groupID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Group ID or Slug is required"))
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

	// Resolve group ID if slug was provided
	group, err := h.s.Group.Get(c.Request.Context(), &structs.FindGroup{Group: groupID})
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Group not found"))
		return
	}

	// Check if this is the only owner and trying to change role
	if req.Role != structs.RoleOwner {
		isOwner, _ := h.s.UserGroup.HasRole(c.Request.Context(), group.ID, userID, structs.RoleOwner)
		if isOwner {
			// Check if there's only one owner
			owners, _ := h.s.UserGroup.GetMembersByRole(c.Request.Context(), group.ID, structs.RoleOwner)
			if len(owners) <= 1 {
				resp.Fail(c.Writer, resp.BadRequest("Cannot change role of the only owner"))
				return
			}
		}
	}

	// Check if user exists in the group
	isMember, _ := h.s.UserGroup.IsUserMember(c.Request.Context(), group.ID, userID)
	if !isMember {
		resp.Fail(c.Writer, resp.BadRequest("User is not a member of this group"))
		return
	}

	// Update the member's role
	err = h.s.UserGroup.RemoveUserFromGroup(c.Request.Context(), userID, group.ID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	member, err := h.s.UserGroup.AddUserToGroup(c.Request.Context(), userID, group.ID, req.Role)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, member)
}

// RemoveMember handles removing a member from a group.
//
// @Summary Remove group member
// @Description Remove a user from a group
// @Tags org
// @Produce json
// @Param slug path string true "Group ID or Slug"
// @Param userId path string true "User ID"
// @Success 200 {object} resp.Success "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /org/groups/{slug}/members/{userId} [delete]
// @Security Bearer
func (h *groupHandler) RemoveMember(c *gin.Context) {
	groupID := c.Param("slug")
	if groupID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Group ID or Slug is required"))
		return
	}

	userID := c.Param("userId")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest("User ID is required"))
		return
	}

	// Resolve group ID if slug was provided
	group, err := h.s.Group.Get(c.Request.Context(), &structs.FindGroup{Group: groupID})
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Group not found"))
		return
	}

	// Check if this is the only owner
	isOwner, _ := h.s.UserGroup.HasRole(c.Request.Context(), group.ID, userID, structs.RoleOwner)
	if isOwner {
		// Check if there's only one owner
		owners, _ := h.s.UserGroup.GetMembersByRole(c.Request.Context(), group.ID, structs.RoleOwner)
		if len(owners) <= 1 {
			resp.Fail(c.Writer, resp.BadRequest("Cannot remove the only owner"))
			return
		}
	}

	err = h.s.UserGroup.RemoveUserFromGroup(c.Request.Context(), userID, group.ID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]bool{"success": true})
}

// IsUserMember handles checking if a user is a member of a group.
//
// @Summary Check if user is a member
// @Description Check if a user is a member of a group
// @Tags org
// @Produce json
// @Param slug path string true "Group ID or Slug"
// @Param userId path string true "User ID"
// @Success 200 {object} map[string]bool "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /org/groups/{slug}/members/{userId}/check [get]
// @Security Bearer
func (h *groupHandler) IsUserMember(c *gin.Context) {
	groupID := c.Param("slug")
	if groupID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Group ID or Slug is required"))
		return
	}

	userID := c.Param("userId")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest("User ID is required"))
		return
	}

	// Resolve group ID if slug was provided
	group, err := h.s.Group.Get(c.Request.Context(), &structs.FindGroup{Group: groupID})
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Group not found"))
		return
	}

	isMember, err := h.s.UserGroup.IsUserMember(c.Request.Context(), group.ID, userID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]bool{"isMember": isMember})
}

// IsUserOwner handles checking if a user is an owner of a group.
//
// @Summary Check if user is an owner
// @Description Check if a user has owner role in a group
// @Tags org
// @Produce json
// @Param slug path string true "Group ID or Slug"
// @Param userId path string true "User ID"
// @Success 200 {object} map[string]bool "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /org/groups/{slug}/members/{userId}/is-owner [get]
// @Security Bearer
func (h *groupHandler) IsUserOwner(c *gin.Context) {
	groupID := c.Param("slug")
	if groupID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Group ID or Slug is required"))
		return
	}

	userID := c.Param("userId")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest("User ID is required"))
		return
	}

	// Resolve group ID if slug was provided
	group, err := h.s.Group.Get(c.Request.Context(), &structs.FindGroup{Group: groupID})
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Group not found"))
		return
	}

	isOwner, err := h.s.UserGroup.HasRole(c.Request.Context(), group.ID, userID, structs.RoleOwner)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]bool{"isOwner": isOwner})
}

// GetUserRole handles getting a user's role in a group.
//
// @Summary Get user role
// @Description Get a user's role in a group
// @Tags org
// @Produce json
// @Param slug path string true "Group ID or Slug"
// @Param userId path string true "User ID"
// @Success 200 {object} map[string]string "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /org/groups/{slug}/members/{userId}/role [get]
// @Security Bearer
func (h *groupHandler) GetUserRole(c *gin.Context) {
	groupID := c.Param("slug")
	if groupID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Group ID or Slug is required"))
		return
	}

	userID := c.Param("userId")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest("User ID is required"))
		return
	}

	// Resolve group ID if slug was provided
	group, err := h.s.Group.Get(c.Request.Context(), &structs.FindGroup{Group: groupID})
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Group not found"))
		return
	}

	role, err := h.s.UserGroup.GetUserRole(c.Request.Context(), group.ID, userID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]string{"role": string(role)})
}
