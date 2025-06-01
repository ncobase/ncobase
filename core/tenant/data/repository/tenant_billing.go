package repository

import (
	"context"
	"fmt"
	"ncobase/tenant/data"
	"ncobase/tenant/data/ent"
	tenantBillingEnt "ncobase/tenant/data/ent/tenantbilling"
	"ncobase/tenant/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
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
	GetByStatus(ctx context.Context, status structs.BillingStatus) ([]*ent.TenantBilling, error)
	GetByInvoiceNumber(ctx context.Context, invoiceNumber string) (*ent.TenantBilling, error)
	Update(ctx context.Context, id string, updates types.JSON) (*ent.TenantBilling, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params *structs.ListTenantBillingParams) ([]*ent.TenantBilling, error)
	ListWithCount(ctx context.Context, params *structs.ListTenantBillingParams) ([]*ent.TenantBilling, int, error)
	MarkOverdue(ctx context.Context, currentTime int64) error
	CountX(ctx context.Context, params *structs.ListTenantBillingParams) int
}

// tenantBillingRepository implements TenantBillingRepositoryInterface
type tenantBillingRepository struct {
	data                      *data.Data
	billingCache              cache.ICache[ent.TenantBilling]
	tenantBillingsCache       cache.ICache[[]string] // Maps tenant ID to billing IDs
	statusBillingsCache       cache.ICache[[]string] // Maps status to billing IDs
	invoiceNumberBillingCache cache.ICache[string]   // Maps invoice number to billing ID
	overdueBillingsCache      cache.ICache[[]string] // Maps tenant to overdue billing IDs
	billingTTL                time.Duration
}

// NewTenantBillingRepository creates a new tenant billing repository
func NewTenantBillingRepository(d *data.Data) TenantBillingRepositoryInterface {
	redisClient := d.GetRedis()

	return &tenantBillingRepository{
		data:                      d,
		billingCache:              cache.NewCache[ent.TenantBilling](redisClient, "ncse_tenant:billings"),
		tenantBillingsCache:       cache.NewCache[[]string](redisClient, "ncse_tenant:tenant_billing_mappings"),
		statusBillingsCache:       cache.NewCache[[]string](redisClient, "ncse_tenant:status_billing_mappings"),
		invoiceNumberBillingCache: cache.NewCache[string](redisClient, "ncse_tenant:invoice_billing_mappings"),
		overdueBillingsCache:      cache.NewCache[[]string](redisClient, "ncse_tenant:overdue_billing_mappings"),
		billingTTL:                time.Hour * 2, // 2 hours cache TTL (billing data changes frequently)
	}
}

// Create creates a new tenant billing record
func (r *tenantBillingRepository) Create(ctx context.Context, body *structs.CreateTenantBillingBody) (*ent.TenantBilling, error) {
	// Use master for writes
	builder := r.data.GetMasterEntClient().TenantBilling.Create()

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

	// Cache the billing and invalidate related caches
	go func() {
		r.cacheBilling(context.Background(), row)
		r.invalidateTenantBillingsCache(context.Background(), body.TenantID)
		r.invalidateStatusBillingsCache(context.Background(), string(body.Status))
		if body.Status == structs.StatusOverdue {
			r.invalidateOverdueBillingsCache(context.Background(), body.TenantID)
		}
	}()

	return row, nil
}

// GetByID retrieves a tenant billing record by ID
func (r *tenantBillingRepository) GetByID(ctx context.Context, id string) (*ent.TenantBilling, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("id:%s", id)
	if cached, err := r.billingCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Use slave for reads
	row, err := r.data.GetSlaveEntClient().TenantBilling.Get(ctx, id)
	if err != nil {
		logger.Errorf(ctx, "tenantBillingRepo.GetByID error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheBilling(context.Background(), row)

	return row, nil
}

// GetByTenantID retrieves all billing records for a tenant
func (r *tenantBillingRepository) GetByTenantID(ctx context.Context, tenantID string) ([]*ent.TenantBilling, error) {
	// Try to get billing IDs from cache
	cacheKey := fmt.Sprintf("tenant_billings:%s", tenantID)
	var billingIDs []string
	if err := r.tenantBillingsCache.GetArray(ctx, cacheKey, &billingIDs); err == nil && len(billingIDs) > 0 {
		// Get billings by IDs
		billings := make([]*ent.TenantBilling, 0, len(billingIDs))
		for _, billingID := range billingIDs {
			if billing, err := r.GetByID(ctx, billingID); err == nil {
				billings = append(billings, billing)
			}
		}
		return billings, nil
	}

	// Fallback to database
	rows, err := r.data.GetSlaveEntClient().TenantBilling.Query().
		Where(tenantBillingEnt.TenantIDEQ(tenantID)).
		Order(ent.Desc(tenantBillingEnt.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantBillingRepo.GetByTenantID error: %v", err)
		return nil, err
	}

	// Cache billings and billing IDs
	go func() {
		billingIDs := make([]string, len(rows))
		for i, billing := range rows {
			r.cacheBilling(context.Background(), billing)
			billingIDs[i] = billing.ID
		}

		if err := r.tenantBillingsCache.SetArray(context.Background(), cacheKey, billingIDs, r.billingTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache tenant billings %s: %v", tenantID, err)
		}
	}()

	return rows, nil
}

// GetOverdueByTenant retrieves overdue billing records for a tenant
func (r *tenantBillingRepository) GetOverdueByTenant(ctx context.Context, tenantID string) ([]*ent.TenantBilling, error) {
	// Try to get overdue billing IDs from cache
	cacheKey := fmt.Sprintf("overdue_billings:%s", tenantID)
	var billingIDs []string
	if err := r.overdueBillingsCache.GetArray(ctx, cacheKey, &billingIDs); err == nil && len(billingIDs) > 0 {
		// Get billings by IDs
		billings := make([]*ent.TenantBilling, 0, len(billingIDs))
		for _, billingID := range billingIDs {
			if billing, err := r.GetByID(ctx, billingID); err == nil {
				billings = append(billings, billing)
			}
		}
		return billings, nil
	}

	// Fallback to database
	rows, err := r.data.GetSlaveEntClient().TenantBilling.Query().
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

	// Cache billings and overdue billing IDs
	go func() {
		billingIDs := make([]string, len(rows))
		for i, billing := range rows {
			r.cacheBilling(context.Background(), billing)
			billingIDs[i] = billing.ID
		}

		if err := r.overdueBillingsCache.SetArray(context.Background(), cacheKey, billingIDs, r.billingTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache overdue billings %s: %v", tenantID, err)
		}
	}()

	return rows, nil
}

// GetByStatus retrieves billing records by status
func (r *tenantBillingRepository) GetByStatus(ctx context.Context, status structs.BillingStatus) ([]*ent.TenantBilling, error) {
	// Try to get billing IDs from cache
	cacheKey := fmt.Sprintf("status_billings:%s", string(status))
	var billingIDs []string
	if err := r.statusBillingsCache.GetArray(ctx, cacheKey, &billingIDs); err == nil && len(billingIDs) > 0 {
		// Get billings by IDs
		billings := make([]*ent.TenantBilling, 0, len(billingIDs))
		for _, billingID := range billingIDs {
			if billing, err := r.GetByID(ctx, billingID); err == nil {
				billings = append(billings, billing)
			}
		}
		return billings, nil
	}

	// Fallback to database
	rows, err := r.data.GetSlaveEntClient().TenantBilling.Query().
		Where(tenantBillingEnt.StatusEQ(string(status))).
		Order(ent.Desc(tenantBillingEnt.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantBillingRepo.GetByStatus error: %v", err)
		return nil, err
	}

	// Cache billings and status billing IDs
	go func() {
		billingIDs := make([]string, len(rows))
		for i, billing := range rows {
			r.cacheBilling(context.Background(), billing)
			billingIDs[i] = billing.ID
		}

		if err := r.statusBillingsCache.SetArray(context.Background(), cacheKey, billingIDs, r.billingTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache status billings %s: %v", string(status), err)
		}
	}()

	return rows, nil
}

// GetByInvoiceNumber retrieves a billing record by invoice number
func (r *tenantBillingRepository) GetByInvoiceNumber(ctx context.Context, invoiceNumber string) (*ent.TenantBilling, error) {
	// Try to get billing ID from invoice number mapping cache
	if billingID, err := r.getBillingIDByInvoiceNumber(ctx, invoiceNumber); err == nil && billingID != "" {
		return r.GetByID(ctx, billingID)
	}

	// Fallback to database
	row, err := r.data.GetSlaveEntClient().TenantBilling.Query().
		Where(tenantBillingEnt.InvoiceNumberEQ(invoiceNumber)).
		Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantBillingRepo.GetByInvoiceNumber error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheBilling(context.Background(), row)

	return row, nil
}

// Update updates a tenant billing record
func (r *tenantBillingRepository) Update(ctx context.Context, id string, updates types.JSON) (*ent.TenantBilling, error) {
	// Get original billing for cache invalidation
	originalBilling, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Use master for writes
	builder := originalBilling.Update()

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

	// Invalidate and re-cache
	go func() {
		r.invalidateBillingCache(context.Background(), originalBilling)
		r.cacheBilling(context.Background(), row)

		// Invalidate related caches
		r.invalidateTenantBillingsCache(context.Background(), row.TenantID)

		// Invalidate status caches for both old and new statuses
		if originalBilling.Status != row.Status {
			r.invalidateStatusBillingsCache(context.Background(), originalBilling.Status)
			r.invalidateStatusBillingsCache(context.Background(), row.Status)

			// Handle overdue cache invalidation
			if originalBilling.Status == string(structs.StatusOverdue) || row.Status == string(structs.StatusOverdue) {
				r.invalidateOverdueBillingsCache(context.Background(), row.TenantID)
			}
		} else {
			r.invalidateStatusBillingsCache(context.Background(), row.Status)
			if row.Status == string(structs.StatusOverdue) {
				r.invalidateOverdueBillingsCache(context.Background(), row.TenantID)
			}
		}
	}()

	return row, nil
}

// Delete deletes a tenant billing record
func (r *tenantBillingRepository) Delete(ctx context.Context, id string) error {
	// Get billing first for cache invalidation
	billing, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Use master for writes
	if err := r.data.GetMasterEntClient().TenantBilling.DeleteOneID(id).Exec(ctx); err != nil {
		logger.Errorf(ctx, "tenantBillingRepo.Delete error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateBillingCache(context.Background(), billing)
		r.invalidateTenantBillingsCache(context.Background(), billing.TenantID)
		r.invalidateStatusBillingsCache(context.Background(), billing.Status)
		if billing.Status == string(structs.StatusOverdue) {
			r.invalidateOverdueBillingsCache(context.Background(), billing.TenantID)
		}
	}()

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

	// Cache billings in background
	go func() {
		for _, billing := range rows {
			r.cacheBilling(context.Background(), billing)
		}
	}()

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
	// Get billings that will be marked as overdue for cache invalidation
	overdueRows, err := r.data.GetSlaveEntClient().TenantBilling.Query().
		Where(
			tenantBillingEnt.StatusEQ(string(structs.StatusPending)),
			tenantBillingEnt.DueDateLT(currentTime),
		).
		All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get overdue billings for cache invalidation: %v", err)
	}

	// Use master for writes
	_, err = r.data.GetMasterEntClient().TenantBilling.Update().
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

	// Invalidate caches for affected billings
	go func() {
		for _, billing := range overdueRows {
			// Invalidate individual billing cache
			r.invalidateBillingCache(context.Background(), billing)

			// Invalidate related caches
			r.invalidateTenantBillingsCache(context.Background(), billing.TenantID)
			r.invalidateOverdueBillingsCache(context.Background(), billing.TenantID)
		}

		// Invalidate status caches
		r.invalidateStatusBillingsCache(context.Background(), string(structs.StatusPending))
		r.invalidateStatusBillingsCache(context.Background(), string(structs.StatusOverdue))
	}()

	return nil
}

// CountX counts tenant billing records
func (r *tenantBillingRepository) CountX(ctx context.Context, params *structs.ListTenantBillingParams) int {
	builder := r.buildListQuery(params)
	return builder.CountX(ctx)
}

// buildListQuery builds the list query based on parameters
func (r *tenantBillingRepository) buildListQuery(params *structs.ListTenantBillingParams) *ent.TenantBillingQuery {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().TenantBilling.Query()

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

// cacheBilling caches billing
func (r *tenantBillingRepository) cacheBilling(ctx context.Context, billing *ent.TenantBilling) {
	// Cache by ID
	idKey := fmt.Sprintf("id:%s", billing.ID)
	if err := r.billingCache.Set(ctx, idKey, billing, r.billingTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache billing by ID %s: %v", billing.ID, err)
	}

	// Cache invoice number to ID mapping
	if billing.InvoiceNumber != "" {
		invoiceKey := fmt.Sprintf("invoice:%s", billing.InvoiceNumber)
		if err := r.invoiceNumberBillingCache.Set(ctx, invoiceKey, &billing.ID, r.billingTTL); err != nil {
			logger.Debugf(ctx, "Failed to cache invoice mapping %s: %v", billing.InvoiceNumber, err)
		}
	}
}

// invalidateBillingCache invalidates billing cache
func (r *tenantBillingRepository) invalidateBillingCache(ctx context.Context, billing *ent.TenantBilling) {
	// Invalidate ID cache
	idKey := fmt.Sprintf("id:%s", billing.ID)
	if err := r.billingCache.Delete(ctx, idKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate billing ID cache %s: %v", billing.ID, err)
	}

	// Invalidate invoice number mapping
	if billing.InvoiceNumber != "" {
		invoiceKey := fmt.Sprintf("invoice:%s", billing.InvoiceNumber)
		if err := r.invoiceNumberBillingCache.Delete(ctx, invoiceKey); err != nil {
			logger.Debugf(ctx, "Failed to invalidate invoice mapping cache %s: %v", billing.InvoiceNumber, err)
		}
	}
}

// invalidateTenantBillingsCache invalidates the cache for tenant billings
func (r *tenantBillingRepository) invalidateTenantBillingsCache(ctx context.Context, tenantID string) {
	cacheKey := fmt.Sprintf("tenant_billings:%s", tenantID)
	if err := r.tenantBillingsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate tenant billings cache %s: %v", tenantID, err)
	}
}

// invalidateStatusBillingsCache invalidates the cache for status billings
func (r *tenantBillingRepository) invalidateStatusBillingsCache(ctx context.Context, status string) {
	cacheKey := fmt.Sprintf("status_billings:%s", status)
	if err := r.statusBillingsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate status billings cache %s: %v", status, err)
	}
}

// invalidateOverdueBillingsCache invalidates the cache for overdue billings
func (r *tenantBillingRepository) invalidateOverdueBillingsCache(ctx context.Context, tenantID string) {
	cacheKey := fmt.Sprintf("overdue_billings:%s", tenantID)
	if err := r.overdueBillingsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate overdue billings cache %s: %v", tenantID, err)
	}
}

// getBillingIDByInvoiceNumber gets the billing ID by invoice number
func (r *tenantBillingRepository) getBillingIDByInvoiceNumber(ctx context.Context, invoiceNumber string) (string, error) {
	cacheKey := fmt.Sprintf("invoice:%s", invoiceNumber)
	billingID, err := r.invoiceNumberBillingCache.Get(ctx, cacheKey)
	if err != nil || billingID == nil {
		return "", err
	}
	return *billingID, nil
}
