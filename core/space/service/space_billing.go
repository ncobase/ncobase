package service

import (
	"context"
	"errors"
	"fmt"
	"ncobase/core/space/data"
	"ncobase/core/space/data/repository"
	"ncobase/core/space/structs"
	"time"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
)

// SpaceBillingServiceInterface defines the interface for space billing service
type SpaceBillingServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateSpaceBillingBody) (*structs.ReadSpaceBilling, error)
	Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadSpaceBilling, error)
	Get(ctx context.Context, id string) (*structs.ReadSpaceBilling, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params *structs.ListSpaceBillingParams) (paging.Result[*structs.ReadSpaceBilling], error)
	ProcessPayment(ctx context.Context, req *structs.PaymentRequest) error
	GetBillingSummary(ctx context.Context, spaceID string) (*structs.BillingSummary, error)
	GetOverdueBilling(ctx context.Context, spaceID string) ([]*structs.ReadSpaceBilling, error)
	MarkAsOverdue(ctx context.Context) error
	GenerateInvoice(ctx context.Context, spaceID string, period structs.BillingPeriod) (*structs.ReadSpaceBilling, error)
}

// spaceBillingService implements SpaceBillingServiceInterface
type spaceBillingService struct {
	repo repository.SpaceBillingRepositoryInterface
}

// NewSpaceBillingService creates a new space billing service
func NewSpaceBillingService(d *data.Data) SpaceBillingServiceInterface {
	return &spaceBillingService{
		repo: repository.NewSpaceBillingRepository(d),
	}
}

// Create creates a new space billing record
func (s *spaceBillingService) Create(ctx context.Context, body *structs.CreateSpaceBillingBody) (*structs.ReadSpaceBilling, error) {
	if body.SpaceID == "" {
		return nil, errors.New(ecode.FieldIsRequired("space_id"))
	}
	if body.Amount <= 0 {
		return nil, errors.New(ecode.FieldIsInvalid("amount"))
	}

	row, err := s.repo.Create(ctx, body)
	if err := handleEntError(ctx, "SpaceBilling", err); err != nil {
		return nil, err
	}

	return repository.SerializeSpaceBilling(row), nil
}

// Update updates an existing space billing record
func (s *spaceBillingService) Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadSpaceBilling, error) {
	row, err := s.repo.Update(ctx, id, updates)
	if err := handleEntError(ctx, "SpaceBilling", err); err != nil {
		return nil, err
	}

	return repository.SerializeSpaceBilling(row), nil
}

// Get retrieves a space billing record by ID
func (s *spaceBillingService) Get(ctx context.Context, id string) (*structs.ReadSpaceBilling, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err := handleEntError(ctx, "SpaceBilling", err); err != nil {
		return nil, err
	}

	return repository.SerializeSpaceBilling(row), nil
}

// Delete deletes a space billing record
func (s *spaceBillingService) Delete(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err := handleEntError(ctx, "SpaceBilling", err); err != nil {
		return err
	}
	return nil
}

// List lists space billing records
func (s *spaceBillingService) List(ctx context.Context, params *structs.ListSpaceBillingParams) (paging.Result[*structs.ReadSpaceBilling], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadSpaceBilling, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, total, err := s.repo.ListWithCount(ctx, &lp)
		if err != nil {
			logger.Errorf(ctx, "Error listing space billing: %v", err)
			return nil, 0, err
		}

		return repository.SerializeSpaceBillings(rows), total, nil
	})
}

// ProcessPayment processes a payment for a billing record
func (s *spaceBillingService) ProcessPayment(ctx context.Context, req *structs.PaymentRequest) error {
	billing, err := s.repo.GetByID(ctx, req.BillingID)
	if err != nil {
		return handleEntError(ctx, "SpaceBilling", err)
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
	return handleEntError(ctx, "SpaceBilling", err)
}

// GetBillingSummary retrieves billing summary for a space
func (s *spaceBillingService) GetBillingSummary(ctx context.Context, spaceID string) (*structs.BillingSummary, error) {
	rows, err := s.repo.GetBySpaceID(ctx, spaceID)
	if err != nil {
		return nil, handleEntError(ctx, "SpaceBilling", err)
	}

	summary := &structs.BillingSummary{
		SpaceID:  spaceID,
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

// GetOverdueBilling retrieves overdue billing records for a space
func (s *spaceBillingService) GetOverdueBilling(ctx context.Context, spaceID string) ([]*structs.ReadSpaceBilling, error) {
	rows, err := s.repo.GetOverdueBySpace(ctx, spaceID)
	if err != nil {
		return nil, handleEntError(ctx, "SpaceBilling", err)
	}

	return repository.SerializeSpaceBillings(rows), nil
}

// MarkAsOverdue marks pending billing records as overdue if past due date
func (s *spaceBillingService) MarkAsOverdue(ctx context.Context) error {
	now := time.Now().UnixMilli()
	return s.repo.MarkOverdue(ctx, now)
}

// GenerateInvoice generates a new invoice for a space
func (s *spaceBillingService) GenerateInvoice(ctx context.Context, spaceID string, period structs.BillingPeriod) (*structs.ReadSpaceBilling, error) {
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
	invoiceNumber := s.generateInvoiceNumber(spaceID, now)

	body := &structs.CreateSpaceBillingBody{
		SpaceBillingBody: structs.SpaceBillingBody{
			SpaceID:       spaceID,
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
func (s *spaceBillingService) generateInvoiceNumber(spaceID string, timestamp time.Time) string {
	return fmt.Sprintf("INV-%s-%s", spaceID[:8], timestamp.Format("20060102150405"))
}
