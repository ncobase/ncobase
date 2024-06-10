package handler

import (
	"stocms/internal/data/structs"
	"stocms/pkg/resp"

	"github.com/gin-gonic/gin"
)

// ReadUserHandler handles reading a user.
//
// @Summary Read user
// @Description Retrieve information about a specific user.
// @Tags user
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /user/{username} [get]
func (h *Handler) ReadUserHandler(c *gin.Context) {
	result, err := h.svc.ReadUserService(c, c.Param("username"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// ReadMeHandler handles reading the current user.
//
// @Summary Read current user
// @Description Retrieve information about the current user.
// @Tags account
// @Produce json
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /account [get]
func (h *Handler) ReadMeHandler(c *gin.Context) {
	result, err := h.svc.ReadMeService(c)
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
// @Router /account/password [put]
func (h *Handler) UpdatePasswordHandler(c *gin.Context) {
	var body *structs.UserRequestBody
	if err := c.ShouldBind(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.svc.UpdatePasswordService(c, body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}
