package repository

import (
	"context"
	"fmt"
	"ncobase/tenant/data"
	"ncobase/tenant/data/ent"
	tenantBillingEnt "ncobase/tenant/data/ent/tenantbilling"
	"ncobase/tenant/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"
)

// TenantBillingRepositoryInterface defines the interface for tenant billing repository
type TenantBillingRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateTenantBillingBody) (*ent.TenantBilling, error)
	GetByID(ctx context.Context, id string) (*ent.TenantBilling, error)
	GetByTenantID(ctx context.Context, tenantID string) ([]*ent.TenantBilling, error)
	GetOverdueByTenant(ctx context.Context, tenantID string) ([]*ent.TenantBilling, error)
	Update(ctx context.Context, id string, updates types.JSON) (*ent.TenantBilling, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params *structs.ListTenantBillingParams) ([]*ent.TenantBilling, error)
	ListWithCount(ctx context.Context, params *structs.ListTenantBillingParams) ([]*ent.TenantBilling, int, error)
	MarkOverdue(ctx context.Context, currentTime int64) error
}

// tenantBillingRepository implements TenantBillingRepositoryInterface
type tenantBillingRepository struct {
	ec *ent.Client
}

// NewTenantBillingRepository creates a new tenant billing repository
func NewTenantBillingRepository(d *data.Data) TenantBillingRepositoryInterface {
	return &tenantBillingRepository{ec: d.GetMasterEntClient()}
}

// Create creates a new tenant billing record
func (r *tenantBillingRepository) Create(ctx context.Context, body *structs.CreateTenantBillingBody) (*ent.TenantBilling, error) {
	builder := r.ec.TenantBilling.Create()

	builder.SetTenantID(body.TenantID)
	builder.SetBillingPeriod(string(body.BillingPeriod))
	builder.SetAmount(body.Amount)
	builder.SetCurrency(body.Currency)
	builder.SetStatus(string(body.Status))
	builder.SetDescription(body.Description)
	builder.SetInvoiceNumber(body.InvoiceNumber)
	builder.SetPaymentMethod(body.PaymentMethod)
	builder.SetNillableCreatedBy(body.CreatedBy)

	if body.PeriodStart != nil {
		builder.SetPeriodStart(*body.PeriodStart)
	}

	if body.PeriodEnd != nil {
		builder.SetPeriodEnd(*body.PeriodEnd)
	}

	if body.PaidAt != nil {
		builder.SetPaidAt(*body.PaidAt)
	}

	if body.DueDate != nil {
		builder.SetDueDate(*body.DueDate)
	}

	if !validator.IsNil(body.UsageDetails) && !validator.IsEmpty(body.UsageDetails) {
		builder.SetUsageDetails(*body.UsageDetails)
	}

	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantBillingRepo.Create error: %v", err)
		return nil, err
	}

	return row, nil
}

// GetByID retrieves a tenant billing record by ID
func (r *tenantBillingRepository) GetByID(ctx context.Context, id string) (*ent.TenantBilling, error) {
	row, err := r.ec.TenantBilling.Get(ctx, id)
	if err != nil {
		logger.Errorf(ctx, "tenantBillingRepo.GetByID error: %v", err)
		return nil, err
	}

	return row, nil
}

// GetByTenantID retrieves all billing records for a tenant
func (r *tenantBillingRepository) GetByTenantID(ctx context.Context, tenantID string) ([]*ent.TenantBilling, error) {
	rows, err := r.ec.TenantBilling.Query().
		Where(tenantBillingEnt.TenantIDEQ(tenantID)).
		Order(ent.Desc(tenantBillingEnt.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantBillingRepo.GetByTenantID error: %v", err)
		return nil, err
	}

	return rows, nil
}

// GetOverdueByTenant retrieves overdue billing records for a tenant
func (r *tenantBillingRepository) GetOverdueByTenant(ctx context.Context, tenantID string) ([]*ent.TenantBilling, error) {
	rows, err := r.ec.TenantBilling.Query().
		Where(
			tenantBillingEnt.TenantIDEQ(tenantID),
			tenantBillingEnt.StatusEQ(string(structs.StatusOverdue)),
		).
		Order(ent.Desc(tenantBillingEnt.FieldDueDate)).
		All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantBillingRepo.GetOverdueByTenant error: %v", err)
		return nil, err
	}

	return rows, nil
}

// Update updates a tenant billing record
func (r *tenantBillingRepository) Update(ctx context.Context, id string, updates types.JSON) (*ent.TenantBilling, error) {
	billing, err := r.ec.TenantBilling.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	builder := billing.Update()

	for field, value := range updates {
		switch field {
		case "billing_period":
			builder.SetBillingPeriod(value.(string))
		case "period_start":
			builder.SetPeriodStart(int64(value.(float64)))
		case "period_end":
			builder.SetPeriodEnd(int64(value.(float64)))
		case "amount":
			builder.SetAmount(value.(float64))
		case "currency":
			builder.SetCurrency(value.(string))
		case "status":
			builder.SetStatus(value.(string))
		case "description":
			builder.SetDescription(value.(string))
		case "invoice_number":
			builder.SetInvoiceNumber(value.(string))
		case "payment_method":
			builder.SetPaymentMethod(value.(string))
		case "paid_at":
			builder.SetPaidAt(int64(value.(float64)))
		case "due_date":
			builder.SetDueDate(int64(value.(float64)))
		case "usage_details":
			builder.SetUsageDetails(value.(types.JSON))
		case "extras":
			builder.SetExtras(value.(types.JSON))
		case "updated_by":
			builder.SetUpdatedBy(value.(string))
		}
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantBillingRepo.Update error: %v", err)
		return nil, err
	}

	return row, nil
}

// Delete deletes a tenant billing record
func (r *tenantBillingRepository) Delete(ctx context.Context, id string) error {
	if err := r.ec.TenantBilling.DeleteOneID(id).Exec(ctx); err != nil {
		logger.Errorf(ctx, "tenantBillingRepo.Delete error: %v", err)
		return err
	}

	return nil
}

// List lists tenant billing records
func (r *tenantBillingRepository) List(ctx context.Context, params *structs.ListTenantBillingParams) ([]*ent.TenantBilling, error) {
	builder := r.buildListQuery(params)

	if params.Cursor != "" {
		id, timestamp, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %v", err)
		}

		if !nanoid.IsPrimaryKey(id) {
			return nil, fmt.Errorf("invalid id in cursor: %s", id)
		}

		if params.Direction == "backward" {
			builder.Where(tenantBillingEnt.Or(
				tenantBillingEnt.CreatedAtGT(timestamp),
				tenantBillingEnt.And(
					tenantBillingEnt.CreatedAtEQ(timestamp),
					tenantBillingEnt.IDGT(id),
				),
			))
		} else {
			builder.Where(tenantBillingEnt.Or(
				tenantBillingEnt.CreatedAtLT(timestamp),
				tenantBillingEnt.And(
					tenantBillingEnt.CreatedAtEQ(timestamp),
					tenantBillingEnt.IDLT(id),
				),
			))
		}
	}

	if params.Direction == "backward" {
		builder.Order(ent.Asc(tenantBillingEnt.FieldCreatedAt), ent.Asc(tenantBillingEnt.FieldID))
	} else {
		builder.Order(ent.Desc(tenantBillingEnt.FieldCreatedAt), ent.Desc(tenantBillingEnt.FieldID))
	}

	builder.Limit(params.Limit)

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantBillingRepo.List error: %v", err)
		return nil, err
	}

	return rows, nil
}

// ListWithCount lists tenant billing records with count
func (r *tenantBillingRepository) ListWithCount(ctx context.Context, params *structs.ListTenantBillingParams) ([]*ent.TenantBilling, int, error) {
	builder := r.buildListQuery(params)

	total, err := builder.Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantBillingRepo.ListWithCount count error: %v", err)
		return nil, 0, err
	}

	rows, err := r.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

// MarkOverdue marks pending billing records as overdue
func (r *tenantBillingRepository) MarkOverdue(ctx context.Context, currentTime int64) error {
	_, err := r.ec.TenantBilling.Update().
		Where(
			tenantBillingEnt.StatusEQ(string(structs.StatusPending)),
			tenantBillingEnt.DueDateLT(currentTime),
		).
		SetStatus(string(structs.StatusOverdue)).
		Save(ctx)

	if err != nil {
		logger.Errorf(ctx, "tenantBillingRepo.MarkOverdue error: %v", err)
		return err
	}

	return nil
}

// buildListQuery builds the list query based on parameters
func (r *tenantBillingRepository) buildListQuery(params *structs.ListTenantBillingParams) *ent.TenantBillingQuery {
	builder := r.ec.TenantBilling.Query()

	if params.TenantID != "" {
		builder.Where(tenantBillingEnt.TenantIDEQ(params.TenantID))
	}

	if params.Status != "" {
		builder.Where(tenantBillingEnt.StatusEQ(string(params.Status)))
	}

	if params.BillingPeriod != "" {
		builder.Where(tenantBillingEnt.BillingPeriodEQ(string(params.BillingPeriod)))
	}

	if params.FromDate > 0 {
		builder.Where(tenantBillingEnt.CreatedAtGTE(params.FromDate))
	}

	if params.ToDate > 0 {
		builder.Where(tenantBillingEnt.CreatedAtLTE(params.ToDate))
	}

	if params.IsOverdue != nil && *params.IsOverdue {
		builder.Where(tenantBillingEnt.StatusEQ(string(structs.StatusOverdue)))
	}

	return builder
}
