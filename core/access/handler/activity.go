package handler

import (
	"ncobase/access/service"
	"ncobase/access/structs"
	"strconv"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// ActivityHandlerInterface defines handler operations for activity logs
type ActivityHandlerInterface interface {
	LogActivity(c *gin.Context)
	GetActivity(c *gin.Context)
	ListActivity(c *gin.Context)
	GetUserActivity(c *gin.Context)
}

// activityHandler implements ActivityHandlerInterface
type activityHandler struct {
	s *service.Service
}

// NewActivityHandler creates a new activity log handler
func NewActivityHandler(svc *service.Service) ActivityHandlerInterface {
	return &activityHandler{
		s: svc,
	}
}

// LogActivity records a new activity
//
// @Summary Log activity
// @Description Record a new activity log entry
// @Tags sys
// @Accept json
// @Produce json
// @Param activity body structs.CreateActivityRequest true "Activity log information"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/activities [post]
// @Security Bearer
func (h *activityHandler) LogActivity(c *gin.Context) {
	var body structs.CreateActivityRequest
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, &body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	userID := ctxutil.GetUserID(c.Request.Context())
	_, err := h.s.Activity.LogActivity(c.Request.Context(), userID, &body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// GetActivity retrieves an activity log entry
//
// @Summary Get activity
// @Description Retrieve an activity log entry by its ID
// @Tags sys
// @Produce json
// @Param id path string true "Activity log ID"
// @Success 200 {object} structs.ActivityEntry "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/activities/{id} [get]
// @Security Bearer
func (h *activityHandler) GetActivity(c *gin.Context) {
	result, err := h.s.Activity.GetActivity(c.Request.Context(), c.Param("id"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// ListActivity lists activity logs with filters
//
// @Summary List activities
// @Description List activity log entries with pagination and filters
// @Tags sys
// @Produce json
// @Param user_id query string false "Filter by user ID"
// @Param type query string false "Filter by activity type"
// @Param from_date query int false "Filter from date (timestamp)"
// @Param to_date query int false "Filter to date (timestamp)"
// @Param limit query int false "Number of items per page"
// @Param offset query int false "Offset for pagination"
// @Success 200 {array} structs.ActivityEntry "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/activities [get]
// @Security Bearer
func (h *activityHandler) ListActivity(c *gin.Context) {
	params := &structs.ListActivityParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	// Non-admin users can only see their own activity logs
	userID := ctxutil.GetUserID(c.Request.Context())
	isAdmin := ctxutil.GetUserIsAdmin(c.Request.Context())
	if !isAdmin && params.UserID == "" || params.UserID != userID {
		params.UserID = userID
	}

	result, total, err := h.s.Activity.ListActivity(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]any{
		"items": result,
		"total": total,
	})
}

// GetUserActivity retrieves activity logs for a user
//
// @Summary Get user activity
// @Description Retrieve recent activity log entries for a specific user
// @Tags sys
// @Produce json
// @Param username path string true "Username"
// @Param limit query int false "Number of items to retrieve"
// @Success 200 {array} structs.ActivityEntry "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/{username}/activity [get]
// @Security Bearer
func (h *activityHandler) GetUserActivity(c *gin.Context) {
	// Get and validate username parameter
	username := c.Param("username")
	if username == "" {
		resp.Fail(c.Writer, resp.BadRequest("Username is required"))
		return
	}

	// Parse limit parameter
	limit := 10 // Default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	result, err := h.s.Activity.GetUserActivity(c.Request.Context(), username, limit)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}
