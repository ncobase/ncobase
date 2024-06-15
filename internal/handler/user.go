package handler

import (
	"ncobase/internal/data/structs"
	"ncobase/internal/helper"
	"ncobase/pkg/resp"

	"github.com/gin-gonic/gin"
)

// GetUserHandler handles reading a user.
//
// @Summary Get user
// @Description Retrieve information about a specific user.
// @Tags user
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} structs.User "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/users/{username} [get]
func (h *Handler) GetUserHandler(c *gin.Context) {
	result, err := h.svc.GetUserService(c, c.Param("username"))
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
// @Success 200 {object} structs.User "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/account [get]
// @Security Bearer
func (h *Handler) GetMeHandler(c *gin.Context) {
	result, err := h.svc.GetMeService(c)
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
// @Param body body structs.UserRequestBody true "UserRequestBody object"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/account/password [put]
// @Security Bearer
func (h *Handler) UpdatePasswordHandler(c *gin.Context) {
	body := &structs.UserRequestBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.svc.UpdatePasswordService(c, body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}
