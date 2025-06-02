package handler

import (
	"ncobase/auth/service"
	"ncobase/auth/structs"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// SessionHandlerInterface defines the session handler interface
type SessionHandlerInterface interface {
	List(c *gin.Context)
	Get(c *gin.Context)
	Delete(c *gin.Context)
	DeactivateAll(c *gin.Context)
}

// sessionHandler implements the SessionHandlerInterface
type sessionHandler struct {
	s *service.Service
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(svc *service.Service) SessionHandlerInterface {
	return &sessionHandler{
		s: svc,
	}
}

// List handles listing user sessions
//
// @Summary List user sessions
// @Description List all sessions for the current user
// @Tags auth
// @Produce json
// @Param cursor query string false "Cursor for pagination"
// @Param limit query int false "Number of items to return"
// @Param direction query string false "Direction of pagination"
// @Param is_active query bool false "Filter by active status"
// @Success 200 {array} structs.ReadSession "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sessions [get]
// @Security Bearer
func (h *sessionHandler) List(c *gin.Context) {
	params := &structs.ListSessionParams{
		UserID: ctxutil.GetUserID(c.Request.Context()),
	}

	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Session.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Get handles retrieving a specific session
//
// @Summary Get session
// @Description Retrieve a specific session by ID
// @Tags auth
// @Produce json
// @Param session_id path string true "Session ID"
// @Success 200 {object} structs.ReadSession "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sessions/{session_id} [get]
// @Security Bearer
func (h *sessionHandler) Get(c *gin.Context) {
	sessionID := c.Param("session_id")

	result, err := h.s.Session.GetByID(c.Request.Context(), sessionID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	// Verify the session belongs to the current user
	currentUserID := ctxutil.GetUserID(c.Request.Context())
	if result.UserID != currentUserID {
		resp.Fail(c.Writer, resp.Forbidden("Access denied"))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete handles deleting a specific session
//
// @Summary Delete session
// @Description Delete a specific session (logout from device)
// @Tags auth
// @Produce json
// @Param session_id path string true "Session ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sessions/{session_id} [delete]
// @Security Bearer
func (h *sessionHandler) Delete(c *gin.Context) {
	sessionID := c.Param("session_id")

	// Verify the session belongs to the current user
	session, err := h.s.Session.GetByID(c.Request.Context(), sessionID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	currentUserID := ctxutil.GetUserID(c.Request.Context())
	if session.UserID != currentUserID {
		resp.Fail(c.Writer, resp.Forbidden("Access denied"))
		return
	}

	err = h.s.Session.Delete(c.Request.Context(), sessionID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, nil)
}

// DeactivateAll handles deactivating all sessions for the current user
//
// @Summary Deactivate all sessions
// @Description Deactivate all sessions for the current user (logout from all devices)
// @Tags auth
// @Produce json
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sessions/deactivate-all [post]
// @Security Bearer
func (h *sessionHandler) DeactivateAll(c *gin.Context) {
	currentUserID := ctxutil.GetUserID(c.Request.Context())

	err := h.s.Session.DeactivateByUserID(c.Request.Context(), currentUserID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, nil)
}
