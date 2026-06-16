package handler

import (
	"context"
	"fmt"
	"ncobase/plugin/initialize/service"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/net/resp"

	"github.com/gin-gonic/gin"
)

// InitializeHandlerInterface represents the initialize handler interface.
type InitializeHandlerInterface interface {
	Execute(c *gin.Context)
	GetStatus(c *gin.Context)
	InitializeOrganizations(c *gin.Context)
	InitializeUsers(c *gin.Context)
	ResetInitialization(c *gin.Context)
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

// Execute handles system initialization
func (h *initializeHandler) Execute(c *gin.Context) {
	if !h.ensureInitializationAuthority(c) {
		return
	}

	// Validate and set data mode
	dataMode := c.Query("mode")
	if dataMode != "" {
		if err := h.validateAndSetDataMode(dataMode); err != nil {
			resp.Fail(c.Writer, resp.BadRequest("Invalid data mode: "+err.Error()))
			return
		}
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

// InitializeOrganizations handles organization initialization
func (h *initializeHandler) InitializeOrganizations(c *gin.Context) {
	if !h.ensureInitializationAuthority(c) {
		return
	}

	dataMode := c.Query("mode")
	if dataMode != "" {
		if err := h.validateAndSetDataMode(dataMode); err != nil {
			resp.Fail(c.Writer, resp.BadRequest("Invalid data mode: "+err.Error()))
			return
		}
	}

	state, err := h.s.InitializeOrganizations(c.Request.Context())
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, state)
}

// InitializeUsers handles user initialization
func (h *initializeHandler) InitializeUsers(c *gin.Context) {
	if !h.ensureInitializationAuthority(c) {
		return
	}

	dataMode := c.Query("mode")
	if dataMode != "" {
		if err := h.validateAndSetDataMode(dataMode); err != nil {
			resp.Fail(c.Writer, resp.BadRequest("Invalid data mode: "+err.Error()))
			return
		}
	}

	state, err := h.s.InitializeUsers(c.Request.Context())
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, state)
}

// GetStatus handles status check
func (h *initializeHandler) GetStatus(c *gin.Context) {
	resp.Success(c.Writer, h.s.GetState())
}

// ResetInitialization handles reset
func (h *initializeHandler) ResetInitialization(c *gin.Context) {
	if !h.ensureInitializationAuthority(c) {
		return
	}

	if !h.s.AllowReinitialization() {
		resp.Fail(c.Writer, resp.BadRequest("Reinitialization is not allowed in configuration"))
		return
	}

	state, err := h.s.ResetInitialization(c.Request.Context())
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, state)
}

func (h *initializeHandler) ensureInitializationAuthority(c *gin.Context) bool {
	initToken := c.GetHeader("X-Init-Token")
	if h.s.RequiresInitToken() && initToken == h.s.GetInitToken() {
		return true
	}

	ctx := c.Request.Context()
	if ctxutil.GetUserID(ctx) == "" {
		resp.Fail(c.Writer, resp.UnAuthorized("Valid initialization token or administrator session required"))
		return false
	}

	if hasSystemInitializationPermission(ctx) {
		return true
	}

	resp.Fail(c.Writer, resp.Forbidden("System management permission required"))
	return false
}

func hasSystemInitializationPermission(ctx context.Context) bool {
	if ctxutil.GetUserIsAdmin(ctx) {
		return true
	}

	for _, permission := range ctxutil.GetUserPermissions(ctx) {
		switch permission {
		case "*", "*:*", "*:system", "manage:*", "manage:system", "admin:*", "admin:system":
			return true
		}
	}

	return false
}

// validateAndSetDataMode validates and sets data mode
func (h *initializeHandler) validateAndSetDataMode(mode string) error {
	validModes := map[string]bool{
		"website":    true,
		"company":    true,
		"enterprise": true,
	}

	if !validModes[mode] {
		return fmt.Errorf("invalid mode '%s', must be one of: website, company, enterprise", mode)
	}

	return h.s.SetDataMode(mode)
}
