package handler

import (
	"ncobase/core/space/service"
	"ncobase/core/space/structs"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// SpaceBillingHandlerInterface defines the interface for space billing handler
type SpaceBillingHandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	ProcessPayment(c *gin.Context)
	GetSummary(c *gin.Context)
	GetOverdue(c *gin.Context)
	GenerateInvoice(c *gin.Context)
}

// spaceBillingHandler implements SpaceBillingHandlerInterface
type spaceBillingHandler struct {
	s *service.Service
}

// NewSpaceBillingHandler creates a new space billing handler
func NewSpaceBillingHandler(svc *service.Service) SpaceBillingHandlerInterface {
	return &spaceBillingHandler{s: svc}
}

// Create handles creating a space billing record
//
// @Summary Create space billing
// @Description Create a new space billing record
// @Tags sys
// @Accept json
// @Produce json
// @Param body body structs.CreateSpaceBillingBody true "Billing record"
// @Success 200 {object} structs.ReadSpaceBilling "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/billing [post]
// @Security Bearer
func (h *spaceBillingHandler) Create(c *gin.Context) {
	body := &structs.CreateSpaceBillingBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.SpaceBilling.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Update handles updating a space billing record
//
// @Summary Update space billing
// @Description Update an existing space billing record
// @Tags sys
// @Accept json
// @Produce json
// @Param id path string true "Billing ID"
// @Param body body types.JSON true "Update data"
// @Success 200 {object} structs.ReadSpaceBilling "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/billing/{id} [put]
// @Security Bearer
func (h *spaceBillingHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	updates := &types.JSON{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, updates); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.SpaceBilling.Update(c.Request.Context(), id, *updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get handles retrieving a space billing record
//
// @Summary Get space billing
// @Description Retrieve a space billing record by ID
// @Tags sys
// @Produce json
// @Param id path string true "Billing ID"
// @Success 200 {object} structs.ReadSpaceBilling "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/billing/{id} [get]
// @Security Bearer
func (h *spaceBillingHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.s.SpaceBilling.Get(c.Request.Context(), id)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete handles deleting a space billing record
//
// @Summary Delete space billing
// @Description Delete a space billing record
// @Tags sys
// @Produce json
// @Param id path string true "Billing ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/billing/{id} [delete]
// @Security Bearer
func (h *spaceBillingHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.s.SpaceBilling.Delete(c.Request.Context(), id); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List handles listing space billing records
//
// @Summary List space billing
// @Description Retrieve a list of space billing records
// @Tags sys
// @Produce json
// @Param params query structs.ListSpaceBillingParams true "List parameters"
// @Success 200 {array} structs.ReadSpaceBilling "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/billing [get]
// @Security Bearer
func (h *spaceBillingHandler) List(c *gin.Context) {
	params := &structs.ListSpaceBillingParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.SpaceBilling.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// ProcessPayment handles processing payment for billing
//
// @Summary Process payment
// @Description Process payment for a billing record
// @Tags sys
// @Accept json
// @Produce json
// @Param body body structs.PaymentRequest true "Payment request"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/billing/payment [post]
// @Security Bearer
func (h *spaceBillingHandler) ProcessPayment(c *gin.Context) {
	body := &structs.PaymentRequest{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	if err := h.s.SpaceBilling.ProcessPayment(c.Request.Context(), body); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// GetSummary handles retrieving billing summary for a space
//
// @Summary Get billing summary
// @Description Retrieve billing summary for a space
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {object} structs.BillingSummary "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/billing/summary [get]
// @Security Bearer
func (h *spaceBillingHandler) GetSummary(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("space_id")))
		return
	}

	result, err := h.s.SpaceBilling.GetBillingSummary(c.Request.Context(), spaceID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetOverdue handles retrieving overdue billing records
//
// @Summary Get overdue billing
// @Description Retrieve overdue billing records for a space
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {array} structs.ReadSpaceBilling "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/billing/overdue [get]
// @Security Bearer
func (h *spaceBillingHandler) GetOverdue(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("space_id")))
		return
	}

	result, err := h.s.SpaceBilling.GetOverdueBilling(c.Request.Context(), spaceID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GenerateInvoice handles generating an invoice for a space
//
// @Summary Generate invoice
// @Description Generate a new invoice for a space
// @Tags sys
// @Accept json
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param body body map[string]string true "Invoice generation request"
// @Success 200 {object} structs.ReadSpaceBilling "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/billing/invoice [post]
// @Security Bearer
func (h *spaceBillingHandler) GenerateInvoice(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("space_id")))
		return
	}

	var body map[string]string
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	period, ok := body["period"]
	if !ok {
		resp.Fail(c.Writer, resp.BadRequest("Missing billing period"))
		return
	}

	billingPeriod := structs.BillingPeriod(period)
	if billingPeriod != structs.PeriodMonthly && billingPeriod != structs.PeriodYearly {
		resp.Fail(c.Writer, resp.BadRequest("Invalid billing period"))
		return
	}

	result, err := h.s.SpaceBilling.GenerateInvoice(c.Request.Context(), spaceID, billingPeriod)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}
