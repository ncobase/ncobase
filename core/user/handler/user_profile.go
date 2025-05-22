package handler

import (
	"ncobase/user/service"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"
)

// UserProfileHandlerInterface is the interface for the handler.
type UserProfileHandlerInterface interface {
	Get(c *gin.Context)
	Update(c *gin.Context)
}

type userProfileHandler struct {
	s *service.Service
}

func NewUserProfileHandler(svc *service.Service) UserProfileHandlerInterface {
	return &userProfileHandler{
		s: svc,
	}
}

// Get handles reading a user profile.
//
// @Summary Get user profile
// @Description Retrieve information about a specific user profile.
// @Tags sys
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} structs.ReadUserProfile "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/{username}/profile [get]
func (h *userProfileHandler) Get(c *gin.Context) {
	result, err := h.s.UserProfile.Get(c.Request.Context(), c.Param("username"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Update handles updating a user profile.
//
// @Summary Update user profile
// @Description Update an existing user profile.
// @Tags sys
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param user_profile body types.JSON true "User profile information to update"
// @Success 200 {object} structs.ReadUserProfile "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/{username}/profile [put]
func (h *userProfileHandler) Update(c *gin.Context) {
	var updates types.JSON
	if err := c.ShouldBindJSON(&updates); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	username := c.Param("username")
	result, err := h.s.UserProfile.Update(c.Request.Context(), username, updates)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}
