// user/handler/user.go

package handler

import (
	"ncobase/core/user/service"
	"ncobase/core/user/structs"
	"ncobase/ncore/helper"
	"ncobase/ncore/resp"
	"ncobase/ncore/types"

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
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
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
