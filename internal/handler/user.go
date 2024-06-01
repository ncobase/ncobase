package handler

import (
	"stocms/pkg/resp"

	"github.com/gin-gonic/gin"
)

// ReadUserHandler Read user handler
func (h *Handler) ReadUserHandler(c *gin.Context) {
	result, err := h.svc.ReadUserService(c, c.Param("id"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// ReadMeHandler Read current user handler
func (h *Handler) ReadMeHandler(c *gin.Context) {
	result, err := h.svc.ReadMeService(c)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// // UpdateUserHandler Update user handler
// func (h *Handler) UpdateUserHandler(c *gin.Context) {
// 	var body *structs.UserRequestBody
// 	if err := c.ShouldBind(&body); err != nil {
// 		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
// 		return
// 	}
// 	result, err := h.svc.UpdateUserService(c, body)
// 	if err != nil {
// 		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
// 		return
// 	}
// 	resp.Success(c.Writer, result)
// }
