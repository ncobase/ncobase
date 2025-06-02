package service

import (
	"context"
	"errors"
	"fmt"
	"ncobase/tenant/data"
	"ncobase/tenant/data/ent"
	"ncobase/tenant/data/repository"
	"ncobase/tenant/structs"
	"time"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
)

// TenantBillingServiceInterface defines the interface for tenant billing service
type TenantBillingServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateTenantBillingBody) (*structs.ReadTenantBilling, error)
	Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadTenantBilling, error)
	Get(ctx context.Context, id string) (*structs.ReadTenantBilling, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params *structs.ListTenantBillingParams) (paging.Result[*structs.ReadTenantBilling], error)
	ProcessPayment(ctx context.Context, req *structs.PaymentRequest) error
	GetBillingSummary(ctx context.Context, tenantID string) (*structs.BillingSummary, error)
	GetOverdueBilling(ctx context.Context, tenantID string) ([]*structs.ReadTenantBilling, error)
	MarkAsOverdue(ctx context.Context) error
	GenerateInvoice(ctx context.Context, tenantID string, period structs.BillingPeriod) (*structs.ReadTenantBilling, error)
	Serialize(row *ent.TenantBilling) *structs.ReadTenantBilling
	Serializes(rows []*ent.TenantBilling) []*structs.ReadTenantBilling
}

// tenantBillingService implements TenantBillingServiceInterface
type tenantBillingService struct {
	repo repository.TenantBillingRepositoryInterface
}

// NewTenantBillingService creates a new tenant billing service
func NewTenantBillingService(d *data.Data) TenantBillingServiceInterface {
	return &tenantBillingService{
		repo: repository.NewTenantBillingRepository(d),
	}
}

// Create creates a new tenant billing record
func (s *tenantBillingService) Create(ctx context.Context, body *structs.CreateTenantBillingBody) (*structs.ReadTenantBilling, error) {
	if body.TenantID == "" {
		return nil, errors.New(ecode.FieldIsRequired("tenant_id"))
	}
	if body.Amount <= 0 {
		return nil, errors.New(ecode.FieldIsInvalid("amount"))
	}

	row, err := s.repo.Create(ctx, body)
	if err := handleEntError(ctx, "TenantBilling", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Update updates an existing tenant billing record
func (s *tenantBillingService) Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadTenantBilling, error) {
	row, err := s.repo.Update(ctx, id, updates)
	if err := handleEntError(ctx, "TenantBilling", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Get retrieves a tenant billing record by ID
func (s *tenantBillingService) Get(ctx context.Context, id string) (*structs.ReadTenantBilling, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err := handleEntError(ctx, "TenantBilling", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Delete deletes a tenant billing record
func (s *tenantBillingService) Delete(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err := handleEntError(ctx, "TenantBilling", err); err != nil {
		return err
	}
	return nil
}

// List lists tenant billing records
func (s *tenantBillingService) List(ctx context.Context, params *structs.ListTenantBillingParams) (paging.Result[*structs.ReadTenantBilling], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadTenantBilling, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, total, err := s.repo.ListWithCount(ctx, &lp)
		if err != nil {
			logger.Errorf(ctx, "Error listing tenant billing: %v", err)
			return nil, 0, err
		}

		return s.Serializes(rows), total, nil
	})
}

// ProcessPayment processes a payment for a billing record
func (s *tenantBillingService) ProcessPayment(ctx context.Context, req *structs.PaymentRequest) error {
	billing, err := s.repo.GetByID(ctx, req.BillingID)
	if err != nil {
		return handleEntError(ctx, "TenantBilling", err)
	}

	if billing.Status != string(structs.StatusPending) && billing.Status != string(structs.StatusOverdue) {
		return errors.New("billing record is not payable")
	}

	now := time.Now().UnixMilli()
	updates := types.JSON{
		"status":         string(structs.StatusPaid),
		"payment_method": req.PaymentMethod,
		"paid_at":        now,
	}

	_, err = s.repo.Update(ctx, req.BillingID, updates)
	return handleEntError(ctx, "TenantBilling", err)
}

// GetBillingSummary retrieves billing summary for a tenant
func (s *tenantBillingService) GetBillingSummary(ctx context.Context, tenantID string) (*structs.BillingSummary, error) {
	rows, err := s.repo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, handleEntError(ctx, "TenantBilling", err)
	}

	summary := &structs.BillingSummary{
		TenantID: tenantID,
		Currency: "USD", // Default currency
	}

	for _, row := range rows {
		summary.TotalInvoices++
		summary.TotalAmount += row.Amount

		switch structs.BillingStatus(row.Status) {
		case structs.StatusPaid:
			summary.PaidInvoices++
			summary.PaidAmount += row.Amount
		case structs.StatusPending:
			summary.PendingAmount += row.Amount
		case structs.StatusOverdue:
			summary.OverdueInvoices++
			summary.OverdueAmount += row.Amount
		}

		// Use currency from first record
		if summary.Currency == "USD" && row.Currency != "" {
			summary.Currency = row.Currency
		}
	}

	return summary, nil
}

// GetOverdueBilling retrieves overdue billing records for a tenant
func (s *tenantBillingService) GetOverdueBilling(ctx context.Context, tenantID string) ([]*structs.ReadTenantBilling, error) {
	rows, err := s.repo.GetOverdueByTenant(ctx, tenantID)
	if err != nil {
		return nil, handleEntError(ctx, "TenantBilling", err)
	}

	return s.Serializes(rows), nil
}

// MarkAsOverdue marks pending billing records as overdue if past due date
func (s *tenantBillingService) MarkAsOverdue(ctx context.Context) error {
	now := time.Now().UnixMilli()
	return s.repo.MarkOverdue(ctx, now)
}

// GenerateInvoice generates a new invoice for a tenant
func (s *tenantBillingService) GenerateInvoice(ctx context.Context, tenantID string, period structs.BillingPeriod) (*structs.ReadTenantBilling, error) {
	now := time.Now()

	var periodStart, periodEnd time.Time
	switch period {
	case structs.PeriodMonthly:
		periodStart = now.AddDate(0, -1, 0)
		periodEnd = now
	case structs.PeriodYearly:
		periodStart = now.AddDate(-1, 0, 0)
		periodEnd = now
	default:
		return nil, errors.New("unsupported billing period")
	}

	// Generate invoice number
	invoiceNumber := s.generateInvoiceNumber(tenantID, now)

	body := &structs.CreateTenantBillingBody{
		TenantBillingBody: structs.TenantBillingBody{
			TenantID:      tenantID,
			BillingPeriod: period,
			PeriodStart:   &[]int64{periodStart.UnixMilli()}[0],
			PeriodEnd:     &[]int64{periodEnd.UnixMilli()}[0],
			Amount:        0, // Should be calculated based on usage
			Currency:      "USD",
			Status:        structs.StatusPending,
			InvoiceNumber: invoiceNumber,
			DueDate:       &[]int64{now.AddDate(0, 0, 30).UnixMilli()}[0], // 30 days due
		},
	}

	return s.Create(ctx, body)
}

// generateInvoiceNumber generates a unique invoice number
func (s *tenantBillingService) generateInvoiceNumber(tenantID string, timestamp time.Time) string {
	return fmt.Sprintf("INV-%s-%s", tenantID[:8], timestamp.Format("20060102150405"))
}

// Serialize converts entity to struct
func (s *tenantBillingService) Serialize(row *ent.TenantBilling) *structs.ReadTenantBilling {
	result := &structs.ReadTenantBilling{
		ID:            row.ID,
		TenantID:      row.TenantID,
		BillingPeriod: structs.BillingPeriod(row.BillingPeriod),
		PeriodStart:   &row.PeriodStart,
		PeriodEnd:     &row.PeriodEnd,
		Amount:        row.Amount,
		Currency:      row.Currency,
		Status:        structs.BillingStatus(row.Status),
		Description:   row.Description,
		InvoiceNumber: row.InvoiceNumber,
		PaymentMethod: row.PaymentMethod,
		PaidAt:        &row.PaidAt,
		DueDate:       &row.DueDate,
		UsageDetails:  &row.UsageDetails,
		Extras:        &row.Extras,
		CreatedBy:     &row.CreatedBy,
		CreatedAt:     &row.CreatedAt,
		UpdatedBy:     &row.UpdatedBy,
		UpdatedAt:     &row.UpdatedAt,
	}

	result.CalculateOverdue()
	return result
}

// Serializes converts multiple entities to structs
func (s *tenantBillingService) Serializes(rows []*ent.TenantBilling) []*structs.ReadTenantBilling {
	result := make([]*structs.ReadTenantBilling, len(rows))
	for i, row := range rows {
		result[i] = s.Serialize(row)
	}
	return result
}
