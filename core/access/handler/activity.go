package handler

import (
	"ncobase/access/service"
	"ncobase/access/structs"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
)

// ActivityHandlerInterface defines handler operations for activity logs
type ActivityHandlerInterface interface {
	CreateActivity(c *gin.Context)
	GetActivity(c *gin.Context)
	ListActivities(c *gin.Context)
	GetUserActivities(c *gin.Context)
	SearchActivities(c *gin.Context)
}

type activityHandler struct {
	activity service.ActivityServiceInterface
}

func NewActivityHandler(activity service.ActivityServiceInterface) ActivityHandlerInterface {
	return &activityHandler{
		activity: activity,
	}
}

// CreateActivity logs a new activity
//
// @Summary Log a new activity
// @Description Logs a new activity.
// @Tags Activity
// @Accept json
// @Produce json
// @Param activity body structs.CreateActivityRequest true "Activity details"
// @Success 201 {object} structs.Activity
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /sys/activity [post]
func (h *activityHandler) CreateActivity(c *gin.Context) {
	var req structs.CreateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	userID := ctxutil.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found in context",
		})
		return
	}

	activity, err := h.activity.LogActivity(c, userID, &req)
	if err != nil {
		logger.Errorf(c, "Failed to create activity: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create activity",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": activity,
	})
}

// GetActivity retrieves an activity by ID
//
// @Summary Retrieve an activity by ID
// @Description Retrieves an activity by its ID.
// @Tags Activity
// @Produce json
// @Param id path string true "Activity ID"
// @Success 200 {object} structs.Activity
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /sys/activity/{id} [get]
func (h *activityHandler) GetActivity(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Activity ID is required",
		})
		return
	}

	activity, err := h.activity.GetActivity(c, id)
	if err != nil {
		logger.Errorf(c, "Failed to get activity: %v", err)
		if err.Error() == "activity not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Activity not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve activity",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": activity,
	})
}

// ListActivities lists activities
//
// @Summary List activities
// @Description Lists activities.
// @Tags Activity
// @Produce json
// @Param cursor query string false "Cursor for pagination"
// @Param direction query string false "Direction for pagination (default: forward)"
// @Param limit query string false "Number of activities to retrieve (default: 20, max: 100)"
// @Param offset query string false "Number of activities to skip (default: 0)"
// @Param from_date query string false "Unix timestamp to filter activities created after (default: 0)"
// @Param to_date query string false "Unix timestamp to filter activities created before (default: 0)"
// @Success 200 {object} structs.ListActivityResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /sys/activities [get]
func (h *activityHandler) ListActivities(c *gin.Context) {
	params := h.parseListParams(c)

	result, err := h.activity.ListActivity(c, params)
	if err != nil {
		logger.Errorf(c, "Failed to list activities: %v", err)
		if err.Error() == ecode.FieldIsInvalid("cursor") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid cursor parameter",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list activities",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetUserActivities retrieves activities for a specific user
//
// @Summary Retrieve activities for a specific user
// @Description Retrieves activities for a specific user.
// @Tags Activity
// @Produce json
// @Param username path string true "Username of the user"
// @Param limit query string false "Number of activities to retrieve (default: 10, max: 100)"
// @Param offset query string false "Number of activities to skip (default: 0)"
// @Param from_date query string false "Unix timestamp to filter activities created after (default: 0)"
// @Param to_date query string false "Unix timestamp to filter activities created before (default: 0)"
// @Success 200 {object} structs.ListActivityResponse
// @Failure 400 {object} gin.H
// @Failure 403 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /sys/activity/user/{username} [get]
func (h *activityHandler) GetUserActivities(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Username is required",
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10
	}

	activities, err := h.activity.GetUserActivity(c, username, limit)
	if err != nil {
		logger.Errorf(c, "Failed to get user activities: %v", err)
		if err.Error() == "you don't have permission to view this user's activity logs" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve user activities",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": activities,
		"meta": gin.H{
			"username": username,
			"limit":    limit,
			"count":    len(activities),
		},
	})
}

// SearchActivities performs full-text search on activities
//
// @Summary Perform full-text search on activities
// @Description Performs full-text search on activities.
// @Tags Activity
// @Produce json
// @Param query query string true "Search query"
// @Param from query string false "Unix timestamp to filter activities created after (default: 0)"
// @Param size query string false "Number of activities to retrieve (default: 10, max: 100)"
// @Success 200 {object} structs.SearchActivityResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /sys/activity/search [get]
func (h *activityHandler) SearchActivities(c *gin.Context) {
	params := h.parseSearchParams(c)

	activities, total, err := h.activity.SearchActivity(c, params)
	if err != nil {
		logger.Errorf(c, "Failed to search activities: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to search activities",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": activities,
		"meta": gin.H{
			"total":    total,
			"query":    params.Query,
			"from":     params.From,
			"size":     params.Size,
			"returned": len(activities),
		},
	})
}

// parseListParams parses query parameters for list endpoint
func (h *activityHandler) parseListParams(c *gin.Context) *structs.ListActivityParams {
	params := &structs.ListActivityParams{
		Cursor:    c.Query("cursor"),
		Direction: c.DefaultQuery("direction", "forward"),
		UserID:    c.Query("user_id"),
		Type:      c.Query("type"),
	}

	// Parse limit with validation
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			params.Limit = limit
		} else {
			params.Limit = 20 // Default limit
		}
	} else {
		params.Limit = 20
	}

	// Parse offset
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			params.Offset = offset
		}
	}

	// Parse date filters
	if fromDateStr := c.Query("from_date"); fromDateStr != "" {
		if fromDate, err := strconv.ParseInt(fromDateStr, 10, 64); err == nil && fromDate > 0 {
			params.FromDate = fromDate
		}
	}

	if toDateStr := c.Query("to_date"); toDateStr != "" {
		if toDate, err := strconv.ParseInt(toDateStr, 10, 64); err == nil && toDate > 0 {
			params.ToDate = toDate
		}
	}

	return params
}

// parseSearchParams parses query parameters for search endpoint
func (h *activityHandler) parseSearchParams(c *gin.Context) *structs.SearchActivityParams {
	params := &structs.SearchActivityParams{
		Query:  c.Query("q"),
		UserID: c.Query("user_id"),
		Type:   c.Query("type"),
	}

	// Parse from with validation
	if fromStr := c.Query("from"); fromStr != "" {
		if from, err := strconv.Atoi(fromStr); err == nil && from >= 0 {
			params.From = from
		}
	}

	// Parse size with validation
	if sizeStr := c.Query("size"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 && size <= 100 {
			params.Size = size
		} else {
			params.Size = 20 // Default size
		}
	} else {
		params.Size = 20
	}

	// Parse date filters
	if fromDateStr := c.Query("from_date"); fromDateStr != "" {
		if fromDate, err := strconv.ParseInt(fromDateStr, 10, 64); err == nil && fromDate > 0 {
			params.FromDate = fromDate
		}
	}

	if toDateStr := c.Query("to_date"); toDateStr != "" {
		if toDate, err := strconv.ParseInt(toDateStr, 10, 64); err == nil && toDate > 0 {
			params.ToDate = toDate
		}
	}

	return params
}
