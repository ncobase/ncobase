package handler

import (
	"ncobase/user/service"
	"ncobase/user/structs"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// UserMeshesHandlerInterface defines user meshes handler interface
type UserMeshesHandlerInterface interface {
	GetUserMeshes(c *gin.Context)
	UpdateUserMeshes(c *gin.Context)
}

// userMeshesHandler implements UserMeshesHandlerInterface
type userMeshesHandler struct {
	s *service.Service
}

// NewUserMeshesHandler creates new user meshes handler
func NewUserMeshesHandler(svc *service.Service) UserMeshesHandlerInterface {
	return &userMeshesHandler{
		s: svc,
	}
}

// GetUserMeshes retrieves aggregated user information
//
// @Summary Get user meshes
// @Description Retrieve aggregated user information
// @Tags sys
// @Produce json
// @Param username path string true "Username"
// @Param include_api_keys query bool false "Include API keys"
// @Success 200 {object} structs.UserMeshes "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/{username}/meshes [get]
// @Security Bearer
func (h *userMeshesHandler) GetUserMeshes(c *gin.Context) {
	username := c.Param("username")
	includeApiKeys := c.Query("include_api_keys") == "true"

	// Check permissions - users can only access own data unless admin
	currentUserID := ctxutil.GetUserID(c.Request.Context())
	isAdmin := ctxutil.GetUserIsAdmin(c.Request.Context())

	// Get target user to check permissions
	targetUser, err := h.s.User.Get(c.Request.Context(), username)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	// Only allow access if admin or own data
	if !isAdmin && currentUserID != targetUser.ID {
		resp.Fail(c.Writer, resp.Forbidden("Access denied"))
		return
	}

	result, err := h.s.UserMeshes.GetUserMeshes(c.Request.Context(), username, includeApiKeys)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// UpdateUserMeshes updates aggregated user information
//
// @Summary Update user meshes
// @Description Update aggregated user information
// @Tags sys
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param meshes body structs.UserMeshes true "User meshes data"
// @Success 200 {object} structs.UserMeshes "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/{username}/meshes [put]
// @Security Bearer
func (h *userMeshesHandler) UpdateUserMeshes(c *gin.Context) {
	username := c.Param("username")

	var body structs.UserMeshes
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, &body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	// Check permissions - users can only update own data unless admin
	currentUserID := ctxutil.GetUserID(c.Request.Context())
	isAdmin := ctxutil.GetUserIsAdmin(c.Request.Context())

	// Get target user to check permissions
	targetUser, err := h.s.User.Get(c.Request.Context(), username)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	// Only allow updates if admin or own data
	if !isAdmin && currentUserID != targetUser.ID {
		resp.Fail(c.Writer, resp.Forbidden("Access denied"))
		return
	}

	result, err := h.s.UserMeshes.UpdateUserMeshes(c.Request.Context(), username, &body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}
