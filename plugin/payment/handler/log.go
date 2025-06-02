package handler

import (
	"ncobase/payment/service"
	"ncobase/payment/structs"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"
)

// LogHandlerInterface defines the interface for log handler operations
type LogHandlerInterface interface {
	Get(c *gin.Context)
	GetByOrder(c *gin.Context)
	List(c *gin.Context)
}

// logHandler handles payment log-related requests
type logHandler struct {
	svc service.LogServiceInterface
}

// NewLogHandler creates a new log handler
func NewLogHandler(svc service.LogServiceInterface) LogHandlerInterface {
	return &logHandler{svc: svc}
}

// Get handles getting a payment log by ID
//
// @Summary Get payment log
// @Description Get a payment log by ID
// @Tags payment,logs
// @Produce json
// @Param id path string true "Log ID"
// @Success 200 {object} resp.Exception{data=structs.Log} "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 404 {object} resp.Exception "not found"
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/logs/{id} [get]
// @Security Bearer
func (h *logHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	// Get log
	log, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to get payment log: %v", err)
		resp.Fail(c.Writer, resp.NotFound("Payment log not found"))
		return
	}

	resp.Success(c.Writer, log)
}

// GetByOrder handles getting payment logs for an order
//
// @Summary Get logs by order
// @Description Get payment logs for a specific order
// @Tags payment,logs
// @Produce json
// @Param orderId path string true "Order ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {array} structs.Log "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/logs/order/{orderId} [get]
// @Security Bearer
func (h *logHandler) GetByOrder(c *gin.Context) {
	orderID := c.Param("orderId")
	if orderID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("order_id")))
		return
	}

	// Parse pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSizeStr := c.DefaultQuery("page_size", "20")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 20
	}

	// Get logs
	logs, err := h.svc.GetByOrderID(c.Request.Context(), orderID, page, pageSize)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to get payment logs: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to get payment logs", err))
		return
	}

	resp.Success(c.Writer, logs)
}

// List handles listing payment logs
//
// @Summary List payment logs
// @Description Get a paginated list of payment logs
// @Tags payment,logs
// @Produce json
// @Param order_id query string false "Filter by order ID"
// @Param channel_id query string false "Filter by channel ID"
// @Param type query string false "Filter by log type"
// @Param has_error query bool false "Filter by has error"
// @Param start_date query int64 false "Filter by start date (Unix timestamp)"
// @Param end_date query int64 false "Filter by end date (Unix timestamp)"
// @Param user_id query string false "Filter by user ID"
// @Param cursor query string false "Cursor for pagination"
// @Param page_size query int false "Page size" default(20)
// @Success 200 {array} structs.Log "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/logs [get]
// @Security Bearer
func (h *logHandler) List(c *gin.Context) {
	var query structs.LogQuery
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, &query); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	// Set default page size if not provided
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	// List logs
	result, err := h.svc.List(c.Request.Context(), &query)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to list payment logs: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to list payment logs", err))
		return
	}

	resp.Success(c.Writer, result)
}
