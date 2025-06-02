package handler

import (
	"ncobase/access/service"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"

	"github.com/gin-gonic/gin"
)

// RolePermissionHandlerInterface is the interface for the handler.
type RolePermissionHandlerInterface interface {
	ListRolePermission(c *gin.Context)
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

// ListRolePermission handles listing permissions for a role.
//
// @Summary List permissions for a role
// @Description Retrieve a list of permissions associated with a role by its ID
// @Tags sys
// @Produce json
// @Param slug path string true "Role ID"
// @Success 200 {array} structs.ReadPermission "success"
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
