package structs

import (
	"fmt"
	"time"

	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
)

// BillingStatus represents the status of a billing record
type BillingStatus string

const (
	StatusPending   BillingStatus = "pending"
	StatusPaid      BillingStatus = "paid"
	StatusOverdue   BillingStatus = "overdue"
	StatusCancelled BillingStatus = "cancelled"
	StatusRefunded  BillingStatus = "refunded"
)

// BillingPeriod represents the billing period type
type BillingPeriod string

const (
	PeriodMonthly BillingPeriod = "monthly"
	PeriodYearly  BillingPeriod = "yearly"
	PeriodOneTime BillingPeriod = "one_time"
	PeriodUsage   BillingPeriod = "usage_based"
)

// TenantBillingBody represents billing information for a tenant
type TenantBillingBody struct {
	TenantID      string        `json:"tenant_id,omitempty"`
	BillingPeriod BillingPeriod `json:"billing_period,omitempty"`
	PeriodStart   *int64        `json:"period_start,omitempty"`
	PeriodEnd     *int64        `json:"period_end,omitempty"`
	Amount        float64       `json:"amount,omitempty"`
	Currency      string        `json:"currency,omitempty"`
	Status        BillingStatus `json:"status,omitempty"`
	Description   string        `json:"description,omitempty"`
	InvoiceNumber string        `json:"invoice_number,omitempty"`
	PaymentMethod string        `json:"payment_method,omitempty"`
	PaidAt        *int64        `json:"paid_at,omitempty"`
	DueDate       *int64        `json:"due_date,omitempty"`
	UsageDetails  *types.JSON   `json:"usage_details,omitempty"`
	Extras        *types.JSON   `json:"extras,omitempty"`
	CreatedBy     *string       `json:"created_by,omitempty"`
	UpdatedBy     *string       `json:"updated_by,omitempty"`
}

// CreateTenantBillingBody represents the body for creating tenant billing
type CreateTenantBillingBody struct {
	TenantBillingBody
}

// UpdateTenantBillingBody represents the body for updating tenant billing
type UpdateTenantBillingBody struct {
	ID string `json:"id,omitempty"`
	TenantBillingBody
}

// ReadTenantBilling represents the output schema for retrieving tenant billing
type ReadTenantBilling struct {
	ID            string        `json:"id"`
	TenantID      string        `json:"tenant_id"`
	BillingPeriod BillingPeriod `json:"billing_period"`
	PeriodStart   *int64        `json:"period_start"`
	PeriodEnd     *int64        `json:"period_end"`
	Amount        float64       `json:"amount"`
	Currency      string        `json:"currency"`
	Status        BillingStatus `json:"status"`
	Description   string        `json:"description"`
	InvoiceNumber string        `json:"invoice_number"`
	PaymentMethod string        `json:"payment_method"`
	PaidAt        *int64        `json:"paid_at"`
	DueDate       *int64        `json:"due_date"`
	IsOverdue     bool          `json:"is_overdue"`
	DaysOverdue   int           `json:"days_overdue"`
	UsageDetails  *types.JSON   `json:"usage_details,omitempty"`
	Extras        *types.JSON   `json:"extras,omitempty"`
	CreatedBy     *string       `json:"created_by,omitempty"`
	CreatedAt     *int64        `json:"created_at,omitempty"`
	UpdatedBy     *string       `json:"updated_by,omitempty"`
	UpdatedAt     *int64        `json:"updated_at,omitempty"`
}

// GetCursorValue returns the cursor value
func (r *ReadTenantBilling) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, convert.ToValue(r.CreatedAt))
}

// CalculateOverdue calculates if the billing is overdue
func (r *ReadTenantBilling) CalculateOverdue() {
	if r.DueDate != nil && r.Status == StatusPending {
		now := time.Now().UnixMilli()
		if now > *r.DueDate {
			r.IsOverdue = true
			r.DaysOverdue = int((now - *r.DueDate) / (24 * 60 * 60 * 1000))
		}
	}
}

// PaymentRequest represents a payment request for billing
type PaymentRequest struct {
	BillingID     string  `json:"billing_id" validate:"required"`
	PaymentMethod string  `json:"payment_method" validate:"required"`
	Amount        float64 `json:"amount" validate:"required"`
	TransactionID string  `json:"transaction_id,omitempty"`
}

// ListTenantBillingParams represents the query parameters for listing tenant billing
type ListTenantBillingParams struct {
	TenantID      string        `form:"tenant_id,omitempty" json:"tenant_id,omitempty"`
	Status        BillingStatus `form:"status,omitempty" json:"status,omitempty"`
	BillingPeriod BillingPeriod `form:"billing_period,omitempty" json:"billing_period,omitempty"`
	FromDate      int64         `form:"from_date,omitempty" json:"from_date,omitempty"`
	ToDate        int64         `form:"to_date,omitempty" json:"to_date,omitempty"`
	IsOverdue     *bool         `form:"is_overdue,omitempty" json:"is_overdue,omitempty"`
	Cursor        string        `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit         int           `form:"limit,omitempty" json:"limit,omitempty"`
	Direction     string        `form:"direction,omitempty" json:"direction,omitempty"`
}

// BillingSummary represents billing summary for a tenant
type BillingSummary struct {
	TenantID        string  `json:"tenant_id"`
	TotalAmount     float64 `json:"total_amount"`
	PaidAmount      float64 `json:"paid_amount"`
	PendingAmount   float64 `json:"pending_amount"`
	OverdueAmount   float64 `json:"overdue_amount"`
	Currency        string  `json:"currency"`
	TotalInvoices   int     `json:"total_invoices"`
	PaidInvoices    int     `json:"paid_invoices"`
	OverdueInvoices int     `json:"overdue_invoices"`
}
