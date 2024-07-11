package handler

import (
	"ncobase/common/resp"
	"ncobase/feature/user/service"
	"ncobase/feature/user/structs"
	"ncobase/helper"

	"github.com/gin-gonic/gin"
)

// UserHandlerInterface is the interface for the handler.
type UserHandlerInterface interface {
	GetUserHandler(c *gin.Context)
	GetMeHandler(c *gin.Context)
	UpdatePasswordHandler(c *gin.Context)
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

// GetUserHandler handles reading a user.
//
// @Summary Get user
// @Description Retrieve information about a specific user.
// @Tags user
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} structs.UserMeshes "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/users/{username} [get]
func (h *userHandler) GetUserHandler(c *gin.Context) {
	result, err := h.s.User.GetUserService(c.Request.Context(), c.Param("username"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetMeHandler handles reading the current user.
//
// @Summary Get current user
// @Description Retrieve information about the current user.
// @Tags account
// @Produce json
// @Success 200 {object} structs.UserMeshes "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/account [get]
// @Security Bearer
func (h *userHandler) GetMeHandler(c *gin.Context) {
	result, err := h.s.User.GetMeService(c.Request.Context())
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// UpdatePasswordHandler handles updating user password.
//
// @Summary Update user password
// @Description Update the password of the current user.
// @Tags account
// @Accept json
// @Produce json
// @Param body body structs.UserPassword true "UserPassword object"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/account/password [put]
// @Security Bearer
func (h *userHandler) UpdatePasswordHandler(c *gin.Context) {
	body := &structs.UserPassword{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.User.UpdatePasswordService(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}
