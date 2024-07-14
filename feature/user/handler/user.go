package handler

import (
	"ncobase/common/resp"
	"ncobase/feature/user/service"

	"github.com/gin-gonic/gin"
)

// UserHandlerInterface is the interface for the handler.
type UserHandlerInterface interface {
	Get(c *gin.Context)
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
// @Tags user
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} structs.ReadUser "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/users/{username} [get]
func (h *userHandler) Get(c *gin.Context) {
	result, err := h.s.User.Get(c.Request.Context(), c.Param("username"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}
