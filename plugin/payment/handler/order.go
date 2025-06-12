package handler

import (
	"ncobase/payment/service"
	"ncobase/payment/structs"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"
)

// OrderHandlerInterface defines the interface for order handler operations
type OrderHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	GetByOrderNumber(c *gin.Context)
	GeneratePaymentURL(c *gin.Context)
	VerifyPayment(c *gin.Context)
	RefundPayment(c *gin.Context)
	List(c *gin.Context)
}

// orderHandler handles payment order-related requests
type orderHandler struct {
	svc service.OrderServiceInterface
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(svc service.OrderServiceInterface) OrderHandlerInterface {
	return &orderHandler{svc: svc}
}

// Create handles the creation of a new payment order
//
// @Summary Create payment order
// @Description Create a new payment order
// @Tags payment,orders
// @Accept json
// @Produce json
// @Param body body structs.CreateOrderInput true "Order data"
// @Success 200 {object} resp.Exception{data=structs.Order} "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/orders [post]
// @Security Bearer
func (h *orderHandler) Create(c *gin.Context) {
	var input structs.CreateOrderInput
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

	// Set space ID from context if not provided
	if input.SpaceID == "" {
		if spaceID, exists := c.Get("space_id"); exists {
			input.SpaceID = spaceID.(string)
		}
	}

	// Create order
	order, err := h.svc.Create(c.Request.Context(), &input)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to create payment order: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to create payment order", err))
		return
	}

	resp.Success(c.Writer, order)
}

// Get handles getting a payment order by ID
//
// @Summary Get payment order
// @Description Get a payment order by ID
// @Tags payment,orders
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} resp.Exception{data=structs.Order} "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 404 {object} resp.Exception "not found"
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/orders/{id} [get]
// @Security Bearer
func (h *orderHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	// Get order
	order, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to get payment order: %v", err)
		resp.Fail(c.Writer, resp.NotFound("Payment order not found"))
		return
	}

	resp.Success(c.Writer, order)
}

// GetByOrderNumber handles getting a payment order by order number
//
// @Summary Get order by number
// @Description Get a payment order by order number
// @Tags payment,orders
// @Produce json
// @Param orderNumber path string true "Order Number"
// @Success 200 {object} resp.Exception{data=structs.Order} "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 404 {object} resp.Exception "not found"
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/orders/number/{orderNumber} [get]
// @Security Bearer
func (h *orderHandler) GetByOrderNumber(c *gin.Context) {
	orderNumber := c.Param("orderNumber")
	if orderNumber == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("order_number")))
		return
	}

	// Get order
	order, err := h.svc.GetByOrderNumber(c.Request.Context(), orderNumber)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to get payment order by number: %v", err)
		resp.Fail(c.Writer, resp.NotFound("Payment order not found"))
		return
	}

	resp.Success(c.Writer, order)
}

// GeneratePaymentURL handles generating a payment URL for a payment order
//
// @Summary Generate payment URL
// @Description Generate a payment URL for a payment order
// @Tags payment,orders
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} resp.Exception{data=map[string]any} "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/orders/{id}/payment-url [post]
// @Security Bearer
func (h *orderHandler) GeneratePaymentURL(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	// Generate payment URL
	providerRef, paymentData, err := h.svc.GeneratePaymentURL(c.Request.Context(), id)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to generate payment URL: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to generate payment URL", err))
		return
	}

	result := map[string]any{
		"provider_ref": providerRef,
		"payment_data": paymentData,
	}

	resp.Success(c.Writer, result)
}

// VerifyPayment handles verifying a payment with the provider
//
// @Summary Verify payment
// @Description Verify a payment with the payment provider
// @Tags payment,orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param body body map[string]any true "Verification data"
// @Success 200 {object} resp.Exception{data=structs.Order} "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/orders/{id}/verify [post]
// @Security Bearer
func (h *orderHandler) VerifyPayment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	var verificationData map[string]any
	if err := c.ShouldBindJSON(&verificationData); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Invalid verification data", err))
		return
	}

	// Verify payment
	if err := h.svc.VerifyPayment(c.Request.Context(), id, verificationData); err != nil {
		logger.Errorf(c.Request.Context(), "Failed to verify payment: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to verify payment", err))
		return
	}

	// Get updated order to return
	order, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to get updated order: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to get updated order", err))
		return
	}

	resp.Success(c.Writer, order)
}

// RefundPayment handles requesting a refund for a payment
//
// @Summary Refund payment
// @Description Request a refund for a payment
// @Tags payment,orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param body body structs.RefundOrderInput true "Refund data"
// @Success 200 {object} resp.Exception{data=structs.Order} "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/orders/{id}/refund [post]
// @Security Bearer
func (h *orderHandler) RefundPayment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	var input structs.RefundOrderInput
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, &input); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	// Refund payment
	if err := h.svc.RefundPayment(c.Request.Context(), id, input.Amount, input.Reason); err != nil {
		logger.Errorf(c.Request.Context(), "Failed to refund payment: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to refund payment", err))
		return
	}

	// Get updated order to return
	order, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to get updated order: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to get updated order", err))
		return
	}

	resp.Success(c.Writer, order)
}

// List handles listing payment orders
//
// @Summary List payment orders
// @Description Get a paginated list of payment orders
// @Tags payment,orders
// @Produce json
// @Param status query string false "Filter by status"
// @Param type query string false "Filter by type"
// @Param channel_id query string false "Filter by channel ID"
// @Param user_id query string false "Filter by user ID"
// @Param space_id query string false "Filter by space ID"
// @Param product_id query string false "Filter by product ID"
// @Param subscription_id query string false "Filter by subscription ID"
// @Param start_date query int64 false "Filter by start date (Unix timestamp)"
// @Param end_date query int64 false "Filter by end date (Unix timestamp)"
// @Param cursor query string false "Cursor for pagination"
// @Param page_size query int false "Page size" default(20)
// @Success 200 {array} structs.Order "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/orders [get]
// @Security Bearer
func (h *orderHandler) List(c *gin.Context) {
	var query structs.OrderQuery
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

	// List orders
	result, err := h.svc.List(c.Request.Context(), &query)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to list payment orders: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to list payment orders", err))
		return
	}

	resp.Success(c.Writer, result)
}
