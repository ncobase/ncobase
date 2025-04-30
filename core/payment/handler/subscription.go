// ./payment/handler/subscription.go
package handler

import (
	"ncobase/core/payment/service"
	"ncobase/core/payment/structs"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"
)

// SubscriptionHandlerInterface defines the interface for subscription handler operations
type SubscriptionHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Update(c *gin.Context)
	Cancel(c *gin.Context)
	GetByUser(c *gin.Context)
	List(c *gin.Context)
}

// subscriptionHandler handles subscription-related requests
type subscriptionHandler struct {
	svc service.SubscriptionServiceInterface
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(svc service.SubscriptionServiceInterface) SubscriptionHandlerInterface {
	return &subscriptionHandler{svc: svc}
}

// Create handles the creation of a new subscription
//
// @Summary Create subscription
// @Description Create a new subscription
// @Tags payment,subscriptions
// @Accept json
// @Produce json
// @Param body body structs.CreateSubscriptionInput true "Subscription data"
// @Success 200 {object} resp.Exception{data=structs.Subscription} "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/subscriptions [post]
// @Security Bearer
func (h *subscriptionHandler) Create(c *gin.Context) {
	var input structs.CreateSubscriptionInput
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, &input); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	// Set user ID from context if not provided
	if input.UserID == "" {
		if userID, exists := c.Get("user_id"); exists {
			input.UserID = userID.(string)
		}
	}

	// Set tenant ID from context if not provided
	if input.TenantID == "" {
		if tenantID, exists := c.Get("tenant_id"); exists {
			input.TenantID = tenantID.(string)
		}
	}

	// Create subscription
	subscription, err := h.svc.Create(c.Request.Context(), &input)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to create subscription: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to create subscription", err))
		return
	}

	resp.Success(c.Writer, subscription)
}

// Get handles getting a subscription by ID
//
// @Summary Get subscription
// @Description Get a subscription by ID
// @Tags payment,subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} resp.Exception{data=structs.Subscription} "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 404 {object} resp.Exception "not found"
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/subscriptions/{id} [get]
// @Security Bearer
func (h *subscriptionHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	// Get subscription
	subscription, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to get subscription: %v", err)
		resp.Fail(c.Writer, resp.NotFound("Subscription not found"))
		return
	}

	resp.Success(c.Writer, subscription)
}

// Update handles updating a subscription
//
// @Summary Update subscription
// @Description Update an existing subscription
// @Tags payment,subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param body body map[string]interface{} true "Updates"
// @Success 200 {object} resp.Exception{data=structs.Subscription} "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/subscriptions/{id} [put]
// @Security Bearer
func (h *subscriptionHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	var updates map[string]any
	if err := c.ShouldBindJSON(&updates); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Invalid request data", err))
		return
	}

	// Update subscription
	subscription, err := h.svc.Update(c.Request.Context(), id, updates)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to update subscription: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to update subscription", err))
		return
	}

	resp.Success(c.Writer, subscription)
}

// Cancel handles cancelling a subscription
//
// @Summary Cancel subscription
// @Description Cancel an existing subscription
// @Tags payment,subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param body body structs.CancelSubscriptionInput true "Cancel options"
// @Success 200 {object} resp.Exception{data=structs.Subscription} "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/subscriptions/{id}/cancel [post]
// @Security Bearer
func (h *subscriptionHandler) Cancel(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	var input structs.CancelSubscriptionInput
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, &input); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	// Cancel subscription
	subscription, err := h.svc.Cancel(c.Request.Context(), id, input.Immediate, input.Reason)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to cancel subscription: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to cancel subscription", err))
		return
	}

	resp.Success(c.Writer, subscription)
}

// GetByUser handles getting subscriptions for a user
//
// @Summary Get user subscriptions
// @Description Get subscriptions for a specific user
// @Tags payment,subscriptions
// @Produce json
// @Param userId path string true "User ID"
// @Param status query string false "Filter by status"
// @Param active query bool false "Filter by active status"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {array} structs.Subscription "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/subscriptions/user/{userId} [get]
// @Security Bearer
func (h *subscriptionHandler) GetByUser(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("user_id")))
		return
	}

	var query struct {
		Status   string `form:"status"`
		Active   bool   `form:"active"`
		Page     int    `form:"page,default=1"`
		PageSize int    `form:"page_size,default=20"`
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Invalid query parameters", err))
		return
	}

	// Convert to proper query struct
	subscriptionQuery := &structs.SubscriptionQuery{
		UserID: userID,
		PaginationQuery: structs.PaginationQuery{
			PageSize: query.PageSize,
		},
	}

	if query.Status != "" {
		subscriptionQuery.Status = structs.SubscriptionStatus(query.Status)
	}

	activePtr := new(bool)
	*activePtr = query.Active
	subscriptionQuery.Active = activePtr

	// Get subscriptions
	result, err := h.svc.GetByUser(c.Request.Context(), subscriptionQuery)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to get user subscriptions: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to get user subscriptions", err))
		return
	}

	resp.Success(c.Writer, result)
}

// List handles listing subscriptions with pagination
//
// @Summary List subscriptions
// @Description Get a paginated list of subscriptions
// @Tags payment,subscriptions
// @Produce json
// @Param status query string false "Filter by status"
// @Param user_id query string false "Filter by user ID"
// @Param tenant_id query string false "Filter by tenant ID"
// @Param product_id query string false "Filter by product ID"
// @Param channel_id query string false "Filter by channel ID"
// @Param active query bool false "Filter by active status"
// @Param cursor query string false "Cursor for pagination"
// @Param page_size query int false "Page size" default(20)
// @Success 200 {array} structs.Subscription "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/subscriptions [get]
// @Security Bearer
func (h *subscriptionHandler) List(c *gin.Context) {
	var query structs.SubscriptionQuery
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

	// List subscriptions
	result, err := h.svc.List(c.Request.Context(), &query)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to list subscriptions: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to list subscriptions", err))
		return
	}

	resp.Success(c.Writer, result)
}
