package handler

import (
	"ncobase/initialize/service"

	"github.com/ncobase/ncore/net/resp"

	"github.com/gin-gonic/gin"
)

// InitializeHandlerInterface represents the initialize handler interface.
type InitializeHandlerInterface interface {
	Execute(c *gin.Context)
	GetStatus(c *gin.Context)
	InitializeOrganizations(c *gin.Context)
	InitializeUsers(c *gin.Context)
}

// initializeHandler represents the initialize handler.
type initializeHandler struct {
	s *service.Service
}

// NewInitializeHandler creates new initialize handler.
func NewInitializeHandler(svc *service.Service) InitializeHandlerInterface {
	return &initializeHandler{
		s: svc,
	}
}

// Execute handles system initialization.
//
// @Summary Initialize system
// @Description Initialize the entire system with roles, permissions, users, etc.
// @Tags system
// @Accept json
// @Produce json
// @Param X-Init-Token header string false "Initialization Token"
// @Success 200 {object} service.InitState "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/initialize [post]
// @Security Bearer
func (h *initializeHandler) Execute(c *gin.Context) {
	// Token validation logic
	initToken := c.GetHeader("X-Init-Token")
	if h.s.RequiresInitToken() && initToken != h.s.GetInitToken() {
		resp.Fail(c.Writer, resp.UnAuthorized("Invalid initialization token"))
		return
	}

	state, err := h.s.Execute(c.Request.Context(), h.s.AllowReinitialization())
	if err != nil {
		if err.Error() == "system is already initialized" {
			resp.Fail(c.Writer, resp.BadRequest("System is already initialized"))
			return
		}
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, state)
}

// InitializeOrganizations handles initializing only organizations.
//
// @Summary Initialize organizations
// @Description Initialize only the organizations
// @Tags system
// @Accept json
// @Produce json
// @Param X-Init-Token header string false "Initialization Token"
// @Success 200 {object} service.InitState "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/initialize/organizations [post]
// @Security Bearer
func (h *initializeHandler) InitializeOrganizations(c *gin.Context) {
	// Token validation logic
	initToken := c.GetHeader("X-Init-Token")
	if h.s.RequiresInitToken() && initToken != h.s.GetInitToken() {
		resp.Fail(c.Writer, resp.UnAuthorized("Invalid initialization token"))
		return
	}

	state, err := h.s.InitializeOrganizations(c.Request.Context())
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, state)
}

// InitializeUsers handles initializing only users.
//
// @Summary Initialize users
// @Description Initialize only the users
// @Tags system
// @Accept json
// @Produce json
// @Param X-Init-Token header string false "Initialization Token"
// @Success 200 {object} service.InitState "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/initialize/users [post]
// @Security Bearer
func (h *initializeHandler) InitializeUsers(c *gin.Context) {
	// Token validation logic
	initToken := c.GetHeader("X-Init-Token")
	if h.s.RequiresInitToken() && initToken != h.s.GetInitToken() {
		resp.Fail(c.Writer, resp.UnAuthorized("Invalid initialization token"))
		return
	}

	state, err := h.s.InitializeUsers(c.Request.Context())
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, state)
}

// GetStatus handles checking initialization status.
//
// @Summary Get initialization status
// @Description Check if the system has been initialized
// @Tags system
// @Produce json
// @Success 200 {object} service.InitState "success"
// @Router /sys/initialize/status [get]
// @Security Bearer
func (h *initializeHandler) GetStatus(c *gin.Context) {
	resp.Success(c.Writer, h.s.GetState())
}
