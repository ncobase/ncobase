package handler

import (
	"ncobase/user/service"
	"ncobase/user/structs"

	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// UserHandlerInterface is the interface for the handler.
type UserHandlerInterface interface {
	Get(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	UpdatePassword(c *gin.Context)
	GetFiltered(c *gin.Context)
	GetByEmail(c *gin.Context)
	GetByUsername(c *gin.Context)
	GetCurrentUser(c *gin.Context)
	UpdateStatus(c *gin.Context)
	GetProfile(c *gin.Context)
	UpdateProfile(c *gin.Context)
	ResetPassword(c *gin.Context)
}

// userHandler represents the handler.
type userHandler struct {
	s *service.Service
}

// NewUserHandler creates a new handler.
func NewUserHandler(svc *service.Service) UserHandlerInterface {
	return &userHandler{
		s: svc,
	}
}

// Get handles reading a user.
//
// @Summary Get user
// @Description Retrieve information about a specific user.
// @Tags sys
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} structs.ReadUser "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/{username} [get]
func (h *userHandler) Get(c *gin.Context) {
	result, err := h.s.User.Get(c.Request.Context(), c.Param("username"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Create handles creating a new user.
//
// @Summary Create user
// @Description Create a new user.
// @Tags sys
// @Accept json
// @Produce json
// @Param user body structs.UserBody true "User information"
// @Success 200 {object} structs.ReadUser "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users [post]
func (h *userHandler) Create(c *gin.Context) {
	var body structs.UserBody
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.s.User.CreateUser(c.Request.Context(), &body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Update handles updating a user.
//
// @Summary Update user
// @Description Update an existing user.
// @Tags sys
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param user body types.JSON true "User information to update"
// @Success 200 {object} structs.ReadUser "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/{username} [put]
func (h *userHandler) Update(c *gin.Context) {
	var updates types.JSON
	if err := c.ShouldBindJSON(&updates); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	username := c.Param("username")
	user, err := h.s.User.Get(c.Request.Context(), username)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.s.User.UpdateUser(c.Request.Context(), user.ID, updates)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Delete handles deleting a user.
//
// @Summary Delete user
// @Description Delete an existing user.
// @Tags sys
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/{username} [delete]
func (h *userHandler) Delete(c *gin.Context) {
	username := c.Param("username")
	user, err := h.s.User.Get(c.Request.Context(), username)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	err = h.s.User.Delete(c.Request.Context(), user.ID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, nil)
}

// List handles listing users.
//
// @Summary List users
// @Description List all users with pagination.
// @Tags sys
// @Produce json
// @Param cursor query string false "Cursor for pagination"
// @Param limit query int false "Number of items to return"
// @Param direction query string false "Direction of pagination (forward or backward)"
// @Success 200 {array} structs.ReadUser "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users [get]
func (h *userHandler) List(c *gin.Context) {
	params := &structs.ListUserParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.User.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// UpdatePassword handles updating a user's password.
//
// @Summary Update user password
// @Description Update an existing user's password.
// @Tags sys
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param password body structs.UserPassword true "Password information"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/{username}/password [put]
func (h *userHandler) UpdatePassword(c *gin.Context) {
	var body structs.UserPassword
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	body.User = c.Param("username")

	err := h.s.User.UpdatePassword(c.Request.Context(), &body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, nil)
}

// GetFiltered handles filtering users.
//
// @Summary Filter users
// @Description Filter users by search query, role, and status.
// @Tags sys
// @Produce json
// @Param search query string false "Search query for name, email or username"
// @Param role query string false "Role filter"
// @Param status query string false "Status filter"
// @Param sortBy query string false "Sort by field"
// @Success 200 {array} structs.ReadUser "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/filter [get]
func (h *userHandler) GetFiltered(c *gin.Context) {
	searchQuery := c.Query("search")
	roleFilter := c.Query("role")
	statusFilter := c.Query("status")
	sortBy := c.Query("sortBy")

	result, err := h.s.User.GetFiltered(c.Request.Context(), searchQuery, roleFilter, statusFilter, sortBy)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetByEmail handles reading a user by email.
//
// @Summary Get user by email
// @Description Retrieve information about a specific user by email.
// @Tags sys
// @Produce json
// @Param email path string true "Email"
// @Success 200 {object} structs.ReadUser "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/by-email/{email} [get]
func (h *userHandler) GetByEmail(c *gin.Context) {
	result, err := h.s.User.GetUserByEmail(c.Request.Context(), c.Param("email"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetByUsername handles reading a user by username.
//
// @Summary Get user by username
// @Description Retrieve information about a specific user by username.
// @Tags sys
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} structs.ReadUser "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/by-username/{username} [get]
func (h *userHandler) GetByUsername(c *gin.Context) {
	result, err := h.s.User.GetUserByUsername(c.Request.Context(), c.Param("username"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetCurrentUser handles retrieving the current authenticated user.
//
// @Summary Get current user
// @Description Retrieve information about the current authenticated user.
// @Tags sys
// @Produce json
// @Success 200 {object} structs.ReadUser "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/me [get]
func (h *userHandler) GetCurrentUser(c *gin.Context) {
	// This would typically get the user ID from the JWT token or session
	// For simplicity, we'll assume the user ID is available in the context
	userID := c.GetString("userID")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest("User not authenticated"))
		return
	}

	result, err := h.s.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// UpdateStatus handles updating a user's status.
//
// @Summary Update user status
// @Description Update a user's status (active, inactive, pending).
// @Tags sys
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param status body map[string]int true "Status information"
// @Success 200 {object} structs.ReadUser "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/{username}/status [patch]
func (h *userHandler) UpdateStatus(c *gin.Context) {
	var body map[string]int
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	status, ok := body["status"]
	if !ok {
		resp.Fail(c.Writer, resp.BadRequest("Status is required"))
		return
	}

	// Check if the status value is valid
	if status != 0 && status != 1 && status != 2 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid status value. Must be 0 (Active), 1 (Inactive), or 2 (Pending)"))
		return
	}

	username := c.Param("username")
	user, err := h.s.User.Get(c.Request.Context(), username)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.s.User.UpdateStatus(c.Request.Context(), user.ID, status)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetProfile handles retrieving a user's profile.
//
// @Summary Get user profile
// @Description Retrieve profile information for a specific user.
// @Tags sys
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} structs.ReadUserProfile "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/{username}/profile [get]
func (h *userHandler) GetProfile(c *gin.Context) {
	username := c.Param("username")
	user, err := h.s.User.Get(c.Request.Context(), username)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	profile, err := h.s.UserProfile.Get(c.Request.Context(), user.ID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, profile)
}

// UpdateProfile handles updating a user's profile.
//
// @Summary Update user profile
// @Description Update profile information for a specific user.
// @Tags sys
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param profile body structs.UserProfileBody true "Profile information"
// @Success 200 {object} structs.ReadUserProfile "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/{username}/profile [put]
func (h *userHandler) UpdateProfile(c *gin.Context) {
	username := c.Param("username")
	user, err := h.s.User.Get(c.Request.Context(), username)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	var body structs.UserProfileBody
	if err = c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	// Make sure the user ID is set
	body.UserID = user.ID

	// Update the profile, creating it if it doesn't exist
	var profile *structs.ReadUserProfile
	existingProfile, err := h.s.UserProfile.Get(c.Request.Context(), user.ID)
	if err == nil && existingProfile != nil {
		profile, err = h.s.UserProfile.Update(c.Request.Context(), user.ID, types.JSON{
			"display_name": body.DisplayName,
			"short_bio":    body.ShortBio,
			"about":        body.About,
			"thumbnail":    body.Thumbnail,
			"links":        body.Links,
			"extras":       body.Extras,
		})
	} else {
		profile, err = h.s.UserProfile.Create(c.Request.Context(), &body)
	}

	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, profile)
}

// ResetPassword handles requesting a password reset.
//
// @Summary Reset user password
// @Description Request a password reset link to be sent to the user's email.
// @Tags sys
// @Accept json
// @Produce json
// @Param request body structs.PasswordResetRequest true "Password reset request"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/reset-password [post]
func (h *userHandler) ResetPassword(c *gin.Context) {
	var request structs.PasswordResetRequest
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, &request); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	// Validate the email and username
	user, err := h.s.User.FindUser(c.Request.Context(), &structs.FindUser{
		Username: request.Username,
		Email:    request.Email,
	})

	if err != nil {
		// Does not provide specific information to prevent email enumeration attacks
		resp.Success(c.Writer, map[string]any{
			"message": "If a matching account was found, a password reset email was sent.",
		})
		return
	}

	// Send the password reset email
	err = h.s.User.SendPasswordResetEmail(c.Request.Context(), user.ID)
	if err != nil {
		// Record the error, but don't return it
		logger.Errorf(c.Request.Context(), "Failed to send password reset email: %v", err)
	}

	// Return a success message
	resp.Success(c.Writer, map[string]any{
		"message": "If a matching account was found, a password reset email was sent.",
	})
}
