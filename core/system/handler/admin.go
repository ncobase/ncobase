package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"

	"ncobase/system/service"
	"ncobase/system/structs"
)

// AdminHandlerInterface defines admin panel operations
type AdminHandlerInterface interface {
	GetSystemHealth(c *gin.Context)
	GetSystemMetrics(c *gin.Context)
	GetUserActivity(c *gin.Context)
	GetSystemLogs(c *gin.Context)
	UpdateSystemConfig(c *gin.Context)
	GetSystemConfig(c *gin.Context)
	GetDashboardStats(c *gin.Context)
	ManageUsers(c *gin.Context)
	GetUserDetails(c *gin.Context)
	UpdateUserStatus(c *gin.Context)
}

type adminHandler struct {
	s *service.Service
}

// NewAdminHandler creates admin handler
func NewAdminHandler(svc *service.Service) AdminHandlerInterface {
	return &adminHandler{s: svc}
}

// GetSystemHealth returns system health status
//
// @Summary Get system health
// @Description Get comprehensive system health information
// @Tags admin
// @Produce json
// @Success 200 {object} structs.SystemHealthResponse "System health status"
// @Failure 500 {object} resp.Exception "Internal server error"
// @Security Bearer
// @Router /admin/health [get]
func (h *adminHandler) GetSystemHealth(c *gin.Context) {
	ctx := c.Request.Context()

	health, err := h.s.Admin.GetSystemHealth(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to get system health: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to retrieve system health"))
		return
	}

	resp.Success(c.Writer, health)
}

// GetSystemMetrics returns system performance metrics
//
// @Summary Get system metrics
// @Description Get system performance and usage metrics
// @Tags admin
// @Produce json
// @Param time_range query string false "Time range (1h, 24h, 7d, 30d)" default(24h)
// @Success 200 {object} structs.SystemMetricsResponse "System metrics"
// @Failure 400 {object} resp.Exception "Bad request"
// @Failure 500 {object} resp.Exception "Internal server error"
// @Security Bearer
// @Router /admin/metrics [get]
func (h *adminHandler) GetSystemMetrics(c *gin.Context) {
	ctx := c.Request.Context()
	timeRange := c.DefaultQuery("time_range", "24h")

	metrics, err := h.s.Admin.GetSystemMetrics(ctx, timeRange)
	if err != nil {
		logger.Errorf(ctx, "Failed to get system metrics: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to retrieve system metrics"))
		return
	}

	resp.Success(c.Writer, metrics)
}

// GetUserActivity returns user activity logs
//
// @Summary Get user activity
// @Description Get paginated user activity logs with filtering
// @Tags admin
// @Produce json
// @Param limit query int false "Limit results" default(20)
// @Param offset query int false "Offset results" default(0)
// @Param user_id query string false "Filter by user ID"
// @Param action query string false "Filter by action type"
// @Param from_date query string false "Start date (RFC3339)"
// @Param to_date query string false "End date (RFC3339)"
// @Success 200 {object} structs.UserActivityResponse "User activity logs"
// @Failure 400 {object} resp.Exception "Bad request"
// @Failure 500 {object} resp.Exception "Internal server error"
// @Security Bearer
// @Router /admin/activity [get]
func (h *adminHandler) GetUserActivity(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filters := &structs.ActivityFilters{
		Limit:    limit,
		Offset:   offset,
		UserID:   c.Query("user_id"),
		Action:   c.Query("action"),
		FromDate: c.Query("from_date"),
		ToDate:   c.Query("to_date"),
	}

	activity, err := h.s.Admin.GetUserActivity(ctx, filters)
	if err != nil {
		logger.Errorf(ctx, "Failed to get user activity: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to retrieve user activity"))
		return
	}

	resp.Success(c.Writer, activity)
}

// GetSystemLogs returns system logs
//
// @Summary Get system logs
// @Description Get paginated system logs with filtering
// @Tags admin
// @Produce json
// @Param limit query int false "Limit results" default(50)
// @Param offset query int false "Offset results" default(0)
// @Param level query string false "Log level (debug, info, warn, error)"
// @Param component query string false "Component name"
// @Param from_date query string false "Start date (RFC3339)"
// @Success 200 {object} structs.SystemLogsResponse "System logs"
// @Failure 400 {object} resp.Exception "Bad request"
// @Failure 500 {object} resp.Exception "Internal server error"
// @Security Bearer
// @Router /admin/logs [get]
func (h *adminHandler) GetSystemLogs(c *gin.Context) {
	ctx := c.Request.Context()

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filters := &structs.LogFilters{
		Limit:     limit,
		Offset:    offset,
		Level:     c.Query("level"),
		Component: c.Query("component"),
		FromDate:  c.Query("from_date"),
	}

	logs, err := h.s.Admin.GetSystemLogs(ctx, filters)
	if err != nil {
		logger.Errorf(ctx, "Failed to get system logs: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to retrieve system logs"))
		return
	}

	resp.Success(c.Writer, logs)
}

// UpdateSystemConfig updates system configuration
//
// @Summary Update system config
// @Description Update system configuration settings
// @Tags admin
// @Accept json
// @Produce json
// @Param config body structs.SystemConfigUpdate true "Configuration updates"
// @Success 200 {object} map[string]interface{} "Updated configuration"
// @Failure 400 {object} resp.Exception "Bad request"
// @Failure 500 {object} resp.Exception "Internal server error"
// @Security Bearer
// @Router /admin/config [put]
func (h *adminHandler) UpdateSystemConfig(c *gin.Context) {
	ctx := c.Request.Context()

	var configUpdate structs.SystemConfigUpdate
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, &configUpdate); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid configuration", validationErrors))
		return
	}

	config, err := h.s.Admin.UpdateSystemConfig(ctx, &configUpdate)
	if err != nil {
		logger.Errorf(ctx, "Failed to update system config: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to update system configuration"))
		return
	}

	resp.Success(c.Writer, config)
}

// GetSystemConfig returns current system configuration
//
// @Summary Get system config
// @Description Get current system configuration
// @Tags admin
// @Produce json
// @Success 200 {object} structs.SystemConfigResponse "System configuration"
// @Failure 500 {object} resp.Exception "Internal server error"
// @Security Bearer
// @Router /admin/config [get]
func (h *adminHandler) GetSystemConfig(c *gin.Context) {
	ctx := c.Request.Context()

	config, err := h.s.Admin.GetSystemConfig(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to get system config: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to retrieve system configuration"))
		return
	}

	resp.Success(c.Writer, config)
}

// GetDashboardStats returns admin dashboard statistics
//
// @Summary Get dashboard stats
// @Description Get comprehensive dashboard statistics for admin panel
// @Tags admin
// @Produce json
// @Success 200 {object} structs.DashboardStatsResponse "Dashboard statistics"
// @Failure 500 {object} resp.Exception "Internal server error"
// @Security Bearer
// @Router /admin/dashboard/stats [get]
func (h *adminHandler) GetDashboardStats(c *gin.Context) {
	ctx := c.Request.Context()

	stats, err := h.s.Admin.GetDashboardStats(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to get dashboard stats: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to retrieve dashboard statistics"))
		return
	}

	resp.Success(c.Writer, stats)
}

// ManageUsers returns paginated user list for management
//
// @Summary Manage users
// @Description Get paginated list of users for management
// @Tags admin
// @Produce json
// @Param limit query int false "Limit results" default(20)
// @Param offset query int false "Offset results" default(0)
// @Param search query string false "Search users by name or email"
// @Param status query string false "Filter by status (active, inactive, suspended)"
// @Param role query string false "Filter by role"
// @Success 200 {object} structs.UserManagementResponse "User management data"
// @Failure 400 {object} resp.Exception "Bad request"
// @Failure 500 {object} resp.Exception "Internal server error"
// @Security Bearer
// @Router /admin/users [get]
func (h *adminHandler) ManageUsers(c *gin.Context) {
	ctx := c.Request.Context()

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filters := &structs.UserFilters{
		Limit:  limit,
		Offset: offset,
		Search: c.Query("search"),
		Status: c.Query("status"),
		Role:   c.Query("role"),
	}

	users, err := h.s.Admin.ManageUsers(ctx, filters)
	if err != nil {
		logger.Errorf(ctx, "Failed to get users for management: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to retrieve users"))
		return
	}

	resp.Success(c.Writer, users)
}

// GetUserDetails returns detailed user information
//
// @Summary Get user details
// @Description Get detailed information about a specific user
// @Tags admin
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} structs.UserDetailsResponse "User details"
// @Failure 400 {object} resp.Exception "Bad request"
// @Failure 404 {object} resp.Exception "User not found"
// @Failure 500 {object} resp.Exception "Internal server error"
// @Security Bearer
// @Router /admin/users/{user_id} [get]
func (h *adminHandler) GetUserDetails(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Param("user_id")

	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest("User ID is required"))
		return
	}

	details, err := h.s.Admin.GetUserDetails(ctx, userID)
	if err != nil {
		logger.Errorf(ctx, "Failed to get user details for %s: %v", userID, err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to retrieve user details"))
		return
	}

	resp.Success(c.Writer, details)
}

// UpdateUserStatus updates user status (active, suspended, etc.)
//
// @Summary Update user status
// @Description Update user status and access permissions
// @Tags admin
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param status body structs.UserStatusUpdate true "Status update"
// @Success 200 {object} map[string]interface{} "Updated user status"
// @Failure 400 {object} resp.Exception "Bad request"
// @Failure 404 {object} resp.Exception "User not found"
// @Failure 500 {object} resp.Exception "Internal server error"
// @Security Bearer
// @Router /admin/users/{user_id}/status [put]
func (h *adminHandler) UpdateUserStatus(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Param("user_id")

	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest("User ID is required"))
		return
	}

	var statusUpdate structs.UserStatusUpdate
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, &statusUpdate); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid status update", validationErrors))
		return
	}

	result, err := h.s.Admin.UpdateUserStatus(ctx, userID, &statusUpdate)
	if err != nil {
		logger.Errorf(ctx, "Failed to update user status for %s: %v", userID, err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to update user status"))
		return
	}

	resp.Success(c.Writer, result)
}
