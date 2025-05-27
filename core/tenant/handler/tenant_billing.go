package handler

import (
	"ncobase/tenant/service"
	"ncobase/tenant/structs"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// TenantBillingHandlerInterface defines the interface for tenant billing handler
type TenantBillingHandlerInterface interface {
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

// tenantBillingHandler implements TenantBillingHandlerInterface
type tenantBillingHandler struct {
	s *service.Service
}

// NewTenantBillingHandler creates a new tenant billing handler
func NewTenantBillingHandler(svc *service.Service) TenantBillingHandlerInterface {
	return &tenantBillingHandler{s: svc}
}

// Create handles creating a tenant billing record
//
// @Summary Create tenant billing
// @Description Create a new tenant billing record
// @Tags iam
// @Accept json
// @Produce json
// @Param body body structs.CreateTenantBillingBody true "Billing record"
// @Success 200 {object} structs.ReadTenantBilling "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/billing [post]
// @Security Bearer
func (h *tenantBillingHandler) Create(c *gin.Context) {
	body := &structs.CreateTenantBillingBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.TenantBilling.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Update handles updating a tenant billing record
//
// @Summary Update tenant billing
// @Description Update an existing tenant billing record
// @Tags iam
// @Accept json
// @Produce json
// @Param id path string true "Billing ID"
// @Param body body types.JSON true "Update data"
// @Success 200 {object} structs.ReadTenantBilling "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/billing/{id} [put]
// @Security Bearer
func (h *tenantBillingHandler) Update(c *gin.Context) {
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

	result, err := h.s.TenantBilling.Update(c.Request.Context(), id, *updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get handles retrieving a tenant billing record
//
// @Summary Get tenant billing
// @Description Retrieve a tenant billing record by ID
// @Tags iam
// @Produce json
// @Param id path string true "Billing ID"
// @Success 200 {object} structs.ReadTenantBilling "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/billing/{id} [get]
// @Security Bearer
func (h *tenantBillingHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.s.TenantBilling.Get(c.Request.Context(), id)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete handles deleting a tenant billing record
//
// @Summary Delete tenant billing
// @Description Delete a tenant billing record
// @Tags iam
// @Produce json
// @Param id path string true "Billing ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/billing/{id} [delete]
// @Security Bearer
func (h *tenantBillingHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.s.TenantBilling.Delete(c.Request.Context(), id); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List handles listing tenant billing records
//
// @Summary List tenant billing
// @Description Retrieve a list of tenant billing records
// @Tags iam
// @Produce json
// @Param params query structs.ListTenantBillingParams true "List parameters"
// @Success 200 {array} structs.ReadTenantBilling "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/billing [get]
// @Security Bearer
func (h *tenantBillingHandler) List(c *gin.Context) {
	params := &structs.ListTenantBillingParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.TenantBilling.List(c.Request.Context(), params)
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
// @Tags iam
// @Accept json
// @Produce json
// @Param body body structs.PaymentRequest true "Payment request"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/billing/payment [post]
// @Security Bearer
func (h *tenantBillingHandler) ProcessPayment(c *gin.Context) {
	body := &structs.PaymentRequest{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	if err := h.s.TenantBilling.ProcessPayment(c.Request.Context(), body); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// GetSummary handles retrieving billing summary for a tenant
//
// @Summary Get billing summary
// @Description Retrieve billing summary for a tenant
// @Tags iam
// @Produce json
// @Param slug path string true "Tenant ID"
// @Success 200 {object} structs.BillingSummary "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/{slug}/billing/summary [get]
// @Security Bearer
func (h *tenantBillingHandler) GetSummary(c *gin.Context) {
	tenantID := c.Param("slug")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("tenant_id")))
		return
	}

	result, err := h.s.TenantBilling.GetBillingSummary(c.Request.Context(), tenantID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetOverdue handles retrieving overdue billing records
//
// @Summary Get overdue billing
// @Description Retrieve overdue billing records for a tenant
// @Tags iam
// @Produce json
// @Param slug path string true "Tenant ID"
// @Success 200 {array} structs.ReadTenantBilling "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/{slug}/billing/overdue [get]
// @Security Bearer
func (h *tenantBillingHandler) GetOverdue(c *gin.Context) {
	tenantID := c.Param("slug")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("tenant_id")))
		return
	}

	result, err := h.s.TenantBilling.GetOverdueBilling(c.Request.Context(), tenantID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GenerateInvoice handles generating an invoice for a tenant
//
// @Summary Generate invoice
// @Description Generate a new invoice for a tenant
// @Tags iam
// @Accept json
// @Produce json
// @Param slug path string true "Tenant ID"
// @Param body body map[string]string true "Invoice generation request"
// @Success 200 {object} structs.ReadTenantBilling "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/{slug}/billing/invoice [post]
// @Security Bearer
func (h *tenantBillingHandler) GenerateInvoice(c *gin.Context) {
	tenantID := c.Param("slug")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("tenant_id")))
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

	result, err := h.s.TenantBilling.GenerateInvoice(c.Request.Context(), tenantID, billingPeriod)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}
