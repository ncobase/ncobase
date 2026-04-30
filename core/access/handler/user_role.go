package handler

import (
	"ncobase/core/access/service"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"
)

// UserRoleHandlerInterface is the interface for user-role relationship endpoints.
type UserRoleHandlerInterface interface {
	GetUserRoles(c *gin.Context)
	AssignRoles(c *gin.Context)
	RemoveRoles(c *gin.Context)
}

type userRoleHandler struct {
	s *service.Service
}

// NewUserRoleHandler creates a new user-role handler.
func NewUserRoleHandler(svc *service.Service) UserRoleHandlerInterface {
	return &userRoleHandler{s: svc}
}

type userRoleBatchBody struct {
	RoleIDs []string `json:"role_ids"`
	RoleIds []string `json:"roleIds"`
}

func (b *userRoleBatchBody) IDs() []string {
	if len(b.RoleIDs) > 0 {
		return b.RoleIDs
	}
	return b.RoleIds
}

// GetUserRoles handles listing roles assigned to a user.
func (h *userRoleHandler) GetUserRoles(c *gin.Context) {
	userID := c.Param("username")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("user")))
		return
	}

	result, err := h.s.UserRole.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// AssignRoles handles assigning roles to a user.
func (h *userRoleHandler) AssignRoles(c *gin.Context) {
	userID := c.Param("username")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("user")))
		return
	}

	body := &userRoleBatchBody{}
	if err := c.ShouldBindJSON(body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	roleIDs := body.IDs()
	if len(roleIDs) == 0 {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("role_ids")))
		return
	}

	assigned := make([]string, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		if roleID == "" {
			continue
		}
		if err := h.s.UserRole.AddRoleToUser(c.Request.Context(), userID, roleID); err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			return
		}
		assigned = append(assigned, roleID)
	}

	resp.Success(c.Writer, gin.H{"user_id": userID, "role_ids": assigned})
}

// RemoveRoles handles removing roles from a user.
func (h *userRoleHandler) RemoveRoles(c *gin.Context) {
	userID := c.Param("username")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("user")))
		return
	}

	body := &userRoleBatchBody{}
	if err := c.ShouldBindJSON(body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	roleIDs := body.IDs()
	if len(roleIDs) == 0 {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("role_ids")))
		return
	}

	removed := make([]string, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		if roleID == "" {
			continue
		}
		if err := h.s.UserRole.RemoveRoleFromUser(c.Request.Context(), userID, roleID); err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			return
		}
		removed = append(removed, roleID)
	}

	resp.Success(c.Writer, gin.H{"user_id": userID, "role_ids": removed})
}
