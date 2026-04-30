package handler

import (
	"ncobase/core/access/service"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"

	"github.com/gin-gonic/gin"
)

// RolePermissionHandlerInterface is the interface for the handler.
type RolePermissionHandlerInterface interface {
	ListRolePermission(c *gin.Context)
	AddPermissionsToRole(c *gin.Context)
	RemovePermissionsFromRole(c *gin.Context)
}

// rolePermissionHandler represents the handler.
type rolePermissionHandler struct {
	s *service.Service
}

// NewRolePermissionHandler creates a new handler.
func NewRolePermissionHandler(svc *service.Service) RolePermissionHandlerInterface {
	return &rolePermissionHandler{
		s: svc,
	}
}

type rolePermissionBatchBody struct {
	PermissionIDs []string `json:"permission_ids"`
	PermissionIds []string `json:"permissionIds"`
}

func (b *rolePermissionBatchBody) IDs() []string {
	if len(b.PermissionIDs) > 0 {
		return b.PermissionIDs
	}
	return b.PermissionIds
}

// ListRolePermission handles listing permissions for a role.
//
// @Summary List permissions for a role
// @Description Retrieve a list of permissions associated with a role by its ID
// @Tags sys
// @Produce json
// @Param slug path string true "Role ID"
// @Success 200 {array} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/roles/{slug}/permissions [get]
// @Security Bearer
func (h *rolePermissionHandler) ListRolePermission(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.s.RolePermission.GetRolePermissions(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// AddPermissionsToRole handles assigning permissions to a role.
func (h *rolePermissionHandler) AddPermissionsToRole(c *gin.Context) {
	roleID := c.Param("slug")
	if roleID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("role")))
		return
	}

	body := &rolePermissionBatchBody{}
	if err := c.ShouldBindJSON(body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	permissionIDs := body.IDs()
	if len(permissionIDs) == 0 {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("permission_ids")))
		return
	}

	assigned := make([]string, 0, len(permissionIDs))
	for _, permissionID := range permissionIDs {
		if permissionID == "" {
			continue
		}
		if _, err := h.s.RolePermission.AddPermissionToRole(c.Request.Context(), roleID, permissionID); err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			return
		}
		assigned = append(assigned, permissionID)
	}

	resp.Success(c.Writer, gin.H{"role_id": roleID, "permission_ids": assigned})
}

// RemovePermissionsFromRole handles removing permissions from a role.
func (h *rolePermissionHandler) RemovePermissionsFromRole(c *gin.Context) {
	roleID := c.Param("slug")
	if roleID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("role")))
		return
	}

	body := &rolePermissionBatchBody{}
	if err := c.ShouldBindJSON(body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	permissionIDs := body.IDs()
	if len(permissionIDs) == 0 {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("permission_ids")))
		return
	}

	removed := make([]string, 0, len(permissionIDs))
	for _, permissionID := range permissionIDs {
		if permissionID == "" {
			continue
		}
		if err := h.s.RolePermission.RemovePermissionFromRole(c.Request.Context(), roleID, permissionID); err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			return
		}
		removed = append(removed, permissionID)
	}

	resp.Success(c.Writer, gin.H{"role_id": roleID, "permission_ids": removed})
}
