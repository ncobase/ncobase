package repository

import (
	"context"
	"fmt"
	"ncobase/space/data"
	"ncobase/space/data/ent"
	spaceBillingEnt "ncobase/space/data/ent/spacebilling"
	"ncobase/space/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"
)

// SpaceBillingRepositoryInterface defines the interface for space billing repository
type SpaceBillingRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateSpaceBillingBody) (*ent.SpaceBilling, error)
	GetByID(ctx context.Context, id string) (*ent.SpaceBilling, error)
	GetBySpaceID(ctx context.Context, spaceID string) ([]*ent.SpaceBilling, error)
	GetOverdueBySpace(ctx context.Context, spaceID string) ([]*ent.SpaceBilling, error)
	GetByStatus(ctx context.Context, status structs.BillingStatus) ([]*ent.SpaceBilling, error)
	GetByInvoiceNumber(ctx context.Context, invoiceNumber string) (*ent.SpaceBilling, error)
	Update(ctx context.Context, id string, updates types.JSON) (*ent.SpaceBilling, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params *structs.ListSpaceBillingParams) ([]*ent.SpaceBilling, error)
	ListWithCount(ctx context.Context, params *structs.ListSpaceBillingParams) ([]*ent.SpaceBilling, int, error)
	MarkOverdue(ctx context.Context, currentTime int64) error
	CountX(ctx context.Context, params *structs.ListSpaceBillingParams) int
}

// spaceBillingRepository implements SpaceBillingRepositoryInterface
type spaceBillingRepository struct {
	data                      *data.Data
	billingCache              cache.ICache[ent.SpaceBilling]
	spaceBillingsCache        cache.ICache[[]string] // Maps space ID to billing IDs
	statusBillingsCache       cache.ICache[[]string] // Maps status to billing IDs
	invoiceNumberBillingCache cache.ICache[string]   // Maps invoice number to billing ID
	overdueBillingsCache      cache.ICache[[]string] // Maps space to overdue billing IDs
	billingTTL                time.Duration
}

// NewSpaceBillingRepository creates a new space billing repository
func NewSpaceBillingRepository(d *data.Data) SpaceBillingRepositoryInterface {
	redisClient := d.GetRedis()

	return &spaceBillingRepository{
		data:                      d,
		billingCache:              cache.NewCache[ent.SpaceBilling](redisClient, "ncse_space:billings"),
		spaceBillingsCache:        cache.NewCache[[]string](redisClient, "ncse_space:space_billing_mappings"),
		statusBillingsCache:       cache.NewCache[[]string](redisClient, "ncse_space:status_billing_mappings"),
		invoiceNumberBillingCache: cache.NewCache[string](redisClient, "ncse_space:invoice_billing_mappings"),
		overdueBillingsCache:      cache.NewCache[[]string](redisClient, "ncse_space:overdue_billing_mappings"),
		billingTTL:                time.Hour * 2, // 2 hours cache TTL (billing data changes frequently)
	}
}

// Create creates a new space billing record
func (r *spaceBillingRepository) Create(ctx context.Context, body *structs.CreateSpaceBillingBody) (*ent.SpaceBilling, error) {
	// Use master for writes
	builder := r.data.GetMasterEntClient().SpaceBilling.Create()

	builder.SetSpaceID(body.SpaceID)
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
		logger.Errorf(ctx, "spaceBillingRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the billing and invalidate related caches
	go func() {
		r.cacheBilling(context.Background(), row)
		r.invalidateSpaceBillingsCache(context.Background(), body.SpaceID)
		r.invalidateStatusBillingsCache(context.Background(), string(body.Status))
		if body.Status == structs.StatusOverdue {
			r.invalidateOverdueBillingsCache(context.Background(), body.SpaceID)
		}
	}()

	return row, nil
}

// GetByID retrieves a space billing record by ID
func (r *spaceBillingRepository) GetByID(ctx context.Context, id string) (*ent.SpaceBilling, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("id:%s", id)
	if cached, err := r.billingCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Use slave for reads
	row, err := r.data.GetSlaveEntClient().SpaceBilling.Get(ctx, id)
	if err != nil {
		logger.Errorf(ctx, "spaceBillingRepo.GetByID error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheBilling(context.Background(), row)

	return row, nil
}

// GetBySpaceID retrieves all billing records for a space
func (r *spaceBillingRepository) GetBySpaceID(ctx context.Context, spaceID string) ([]*ent.SpaceBilling, error) {
	// Try to get billing IDs from cache
	cacheKey := fmt.Sprintf("space_billings:%s", spaceID)
	var billingIDs []string
	if err := r.spaceBillingsCache.GetArray(ctx, cacheKey, &billingIDs); err == nil && len(billingIDs) > 0 {
		// Get billings by IDs
		billings := make([]*ent.SpaceBilling, 0, len(billingIDs))
		for _, billingID := range billingIDs {
			if billing, err := r.GetByID(ctx, billingID); err == nil {
				billings = append(billings, billing)
			}
		}
		return billings, nil
	}

	// Fallback to database
	rows, err := r.data.GetSlaveEntClient().SpaceBilling.Query().
		Where(spaceBillingEnt.SpaceIDEQ(spaceID)).
		Order(ent.Desc(spaceBillingEnt.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceBillingRepo.GetBySpaceID error: %v", err)
		return nil, err
	}

	// Cache billings and billing IDs
	go func() {
		billingIDs := make([]string, 0, len(rows))
		for _, billing := range rows {
			r.cacheBilling(context.Background(), billing)
			billingIDs = append(billingIDs, billing.ID)
		}

		if err := r.spaceBillingsCache.SetArray(context.Background(), cacheKey, billingIDs, r.billingTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache space billings %s: %v", spaceID, err)
		}
	}()

	return rows, nil
}

// GetOverdueBySpace retrieves overdue billing records for a space
func (r *spaceBillingRepository) GetOverdueBySpace(ctx context.Context, spaceID string) ([]*ent.SpaceBilling, error) {
	// Try to get overdue billing IDs from cache
	cacheKey := fmt.Sprintf("overdue_billings:%s", spaceID)
	var billingIDs []string
	if err := r.overdueBillingsCache.GetArray(ctx, cacheKey, &billingIDs); err == nil && len(billingIDs) > 0 {
		// Get billings by IDs
		billings := make([]*ent.SpaceBilling, 0, len(billingIDs))
		for _, billingID := range billingIDs {
			if billing, err := r.GetByID(ctx, billingID); err == nil {
				billings = append(billings, billing)
			}
		}
		return billings, nil
	}

	// Fallback to database
	rows, err := r.data.GetSlaveEntClient().SpaceBilling.Query().
		Where(
			spaceBillingEnt.SpaceIDEQ(spaceID),
			spaceBillingEnt.StatusEQ(string(structs.StatusOverdue)),
		).
		Order(ent.Desc(spaceBillingEnt.FieldDueDate)).
		All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceBillingRepo.GetOverdueBySpace error: %v", err)
		return nil, err
	}

	// Cache billings and overdue billing IDs
	go func() {
		billingIDs := make([]string, 0, len(rows))
		for _, billing := range rows {
			r.cacheBilling(context.Background(), billing)
			billingIDs = append(billingIDs, billing.ID)
		}

		if err := r.overdueBillingsCache.SetArray(context.Background(), cacheKey, billingIDs, r.billingTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache overdue billings %s: %v", spaceID, err)
		}
	}()

	return rows, nil
}

// GetByStatus retrieves billing records by status
func (r *spaceBillingRepository) GetByStatus(ctx context.Context, status structs.BillingStatus) ([]*ent.SpaceBilling, error) {
	// Try to get billing IDs from cache
	cacheKey := fmt.Sprintf("status_billings:%s", string(status))
	var billingIDs []string
	if err := r.statusBillingsCache.GetArray(ctx, cacheKey, &billingIDs); err == nil && len(billingIDs) > 0 {
		// Get billings by IDs
		billings := make([]*ent.SpaceBilling, 0, len(billingIDs))
		for _, billingID := range billingIDs {
			if billing, err := r.GetByID(ctx, billingID); err == nil {
				billings = append(billings, billing)
			}
		}
		return billings, nil
	}

	// Fallback to database
	rows, err := r.data.GetSlaveEntClient().SpaceBilling.Query().
		Where(spaceBillingEnt.StatusEQ(string(status))).
		Order(ent.Desc(spaceBillingEnt.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceBillingRepo.GetByStatus error: %v", err)
		return nil, err
	}

	// Cache billings and status billing IDs
	go func() {
		billingIDs := make([]string, 0, len(rows))
		for _, billing := range rows {
			r.cacheBilling(context.Background(), billing)
			billingIDs = append(billingIDs, billing.ID)
		}

		if err := r.statusBillingsCache.SetArray(context.Background(), cacheKey, billingIDs, r.billingTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache status billings %s: %v", string(status), err)
		}
	}()

	return rows, nil
}

// GetByInvoiceNumber retrieves a billing record by invoice number
func (r *spaceBillingRepository) GetByInvoiceNumber(ctx context.Context, invoiceNumber string) (*ent.SpaceBilling, error) {
	// Try to get billing ID from invoice number mapping cache
	if billingID, err := r.getBillingIDByInvoiceNumber(ctx, invoiceNumber); err == nil && billingID != "" {
		return r.GetByID(ctx, billingID)
	}

	// Fallback to database
	row, err := r.data.GetSlaveEntClient().SpaceBilling.Query().
		Where(spaceBillingEnt.InvoiceNumberEQ(invoiceNumber)).
		Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceBillingRepo.GetByInvoiceNumber error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheBilling(context.Background(), row)

	return row, nil
}

// Update updates a space billing record
func (r *spaceBillingRepository) Update(ctx context.Context, id string, updates types.JSON) (*ent.SpaceBilling, error) {
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
		logger.Errorf(ctx, "spaceBillingRepo.Update error: %v", err)
		return nil, err
	}

	// Invalidate and re-cache
	go func() {
		r.invalidateBillingCache(context.Background(), originalBilling)
		r.cacheBilling(context.Background(), row)

		// Invalidate related caches
		r.invalidateSpaceBillingsCache(context.Background(), row.SpaceID)

		// Invalidate status caches for both old and new statuses
		if originalBilling.Status != row.Status {
			r.invalidateStatusBillingsCache(context.Background(), originalBilling.Status)
			r.invalidateStatusBillingsCache(context.Background(), row.Status)

			// Handle overdue cache invalidation
			if originalBilling.Status == string(structs.StatusOverdue) || row.Status == string(structs.StatusOverdue) {
				r.invalidateOverdueBillingsCache(context.Background(), row.SpaceID)
			}
		} else {
			r.invalidateStatusBillingsCache(context.Background(), row.Status)
			if row.Status == string(structs.StatusOverdue) {
				r.invalidateOverdueBillingsCache(context.Background(), row.SpaceID)
			}
		}
	}()

	return row, nil
}

// Delete deletes a space billing record
func (r *spaceBillingRepository) Delete(ctx context.Context, id string) error {
	// Get billing first for cache invalidation
	billing, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Use master for writes
	if err := r.data.GetMasterEntClient().SpaceBilling.DeleteOneID(id).Exec(ctx); err != nil {
		logger.Errorf(ctx, "spaceBillingRepo.Delete error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateBillingCache(context.Background(), billing)
		r.invalidateSpaceBillingsCache(context.Background(), billing.SpaceID)
		r.invalidateStatusBillingsCache(context.Background(), billing.Status)
		if billing.Status == string(structs.StatusOverdue) {
			r.invalidateOverdueBillingsCache(context.Background(), billing.SpaceID)
		}
	}()

	return nil
}

// List lists space billing records
func (r *spaceBillingRepository) List(ctx context.Context, params *structs.ListSpaceBillingParams) ([]*ent.SpaceBilling, error) {
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
			builder.Where(spaceBillingEnt.Or(
				spaceBillingEnt.CreatedAtGT(timestamp),
				spaceBillingEnt.And(
					spaceBillingEnt.CreatedAtEQ(timestamp),
					spaceBillingEnt.IDGT(id),
				),
			))
		} else {
			builder.Where(spaceBillingEnt.Or(
				spaceBillingEnt.CreatedAtLT(timestamp),
				spaceBillingEnt.And(
					spaceBillingEnt.CreatedAtEQ(timestamp),
					spaceBillingEnt.IDLT(id),
				),
			))
		}
	}

	if params.Direction == "backward" {
		builder.Order(ent.Asc(spaceBillingEnt.FieldCreatedAt), ent.Asc(spaceBillingEnt.FieldID))
	} else {
		builder.Order(ent.Desc(spaceBillingEnt.FieldCreatedAt), ent.Desc(spaceBillingEnt.FieldID))
	}

	builder.Limit(params.Limit)

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceBillingRepo.List error: %v", err)
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

// ListWithCount lists space billing records with count
func (r *spaceBillingRepository) ListWithCount(ctx context.Context, params *structs.ListSpaceBillingParams) ([]*ent.SpaceBilling, int, error) {
	builder := r.buildListQuery(params)

	total, err := builder.Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceBillingRepo.ListWithCount count error: %v", err)
		return nil, 0, err
	}

	rows, err := r.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

// MarkOverdue marks pending billing records as overdue
func (r *spaceBillingRepository) MarkOverdue(ctx context.Context, currentTime int64) error {
	// Get billings that will be marked as overdue for cache invalidation
	overdueRows, err := r.data.GetSlaveEntClient().SpaceBilling.Query().
		Where(
			spaceBillingEnt.StatusEQ(string(structs.StatusPending)),
			spaceBillingEnt.DueDateLT(currentTime),
		).
		All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get overdue billings for cache invalidation: %v", err)
	}

	// Use master for writes
	_, err = r.data.GetMasterEntClient().SpaceBilling.Update().
		Where(
			spaceBillingEnt.StatusEQ(string(structs.StatusPending)),
			spaceBillingEnt.DueDateLT(currentTime),
		).
		SetStatus(string(structs.StatusOverdue)).
		Save(ctx)

	if err != nil {
		logger.Errorf(ctx, "spaceBillingRepo.MarkOverdue error: %v", err)
		return err
	}

	// Invalidate caches for affected billings
	go func() {
		for _, billing := range overdueRows {
			// Invalidate individual billing cache
			r.invalidateBillingCache(context.Background(), billing)

			// Invalidate related caches
			r.invalidateSpaceBillingsCache(context.Background(), billing.SpaceID)
			r.invalidateOverdueBillingsCache(context.Background(), billing.SpaceID)
		}

		// Invalidate status caches
		r.invalidateStatusBillingsCache(context.Background(), string(structs.StatusPending))
		r.invalidateStatusBillingsCache(context.Background(), string(structs.StatusOverdue))
	}()

	return nil
}

// CountX counts space billing records
func (r *spaceBillingRepository) CountX(ctx context.Context, params *structs.ListSpaceBillingParams) int {
	builder := r.buildListQuery(params)
	return builder.CountX(ctx)
}

// buildListQuery builds the list query based on parameters
func (r *spaceBillingRepository) buildListQuery(params *structs.ListSpaceBillingParams) *ent.SpaceBillingQuery {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().SpaceBilling.Query()

	if params.SpaceID != "" {
		builder.Where(spaceBillingEnt.SpaceIDEQ(params.SpaceID))
	}

	if params.Status != "" {
		builder.Where(spaceBillingEnt.StatusEQ(string(params.Status)))
	}

	if params.BillingPeriod != "" {
		builder.Where(spaceBillingEnt.BillingPeriodEQ(string(params.BillingPeriod)))
	}

	if params.FromDate > 0 {
		builder.Where(spaceBillingEnt.CreatedAtGTE(params.FromDate))
	}

	if params.ToDate > 0 {
		builder.Where(spaceBillingEnt.CreatedAtLTE(params.ToDate))
	}

	if params.IsOverdue != nil && *params.IsOverdue {
		builder.Where(spaceBillingEnt.StatusEQ(string(structs.StatusOverdue)))
	}

	return builder
}

// cacheBilling caches billing
func (r *spaceBillingRepository) cacheBilling(ctx context.Context, billing *ent.SpaceBilling) {
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
func (r *spaceBillingRepository) invalidateBillingCache(ctx context.Context, billing *ent.SpaceBilling) {
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

// invalidateSpaceBillingsCache invalidates the cache for space billings
func (r *spaceBillingRepository) invalidateSpaceBillingsCache(ctx context.Context, spaceID string) {
	cacheKey := fmt.Sprintf("space_billings:%s", spaceID)
	if err := r.spaceBillingsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate space billings cache %s: %v", spaceID, err)
	}
}

// invalidateStatusBillingsCache invalidates the cache for status billings
func (r *spaceBillingRepository) invalidateStatusBillingsCache(ctx context.Context, status string) {
	cacheKey := fmt.Sprintf("status_billings:%s", status)
	if err := r.statusBillingsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate status billings cache %s: %v", status, err)
	}
}

// invalidateOverdueBillingsCache invalidates the cache for overdue billings
func (r *spaceBillingRepository) invalidateOverdueBillingsCache(ctx context.Context, spaceID string) {
	cacheKey := fmt.Sprintf("overdue_billings:%s", spaceID)
	if err := r.overdueBillingsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate overdue billings cache %s: %v", spaceID, err)
	}
}

// getBillingIDByInvoiceNumber gets the billing ID by invoice number
func (r *spaceBillingRepository) getBillingIDByInvoiceNumber(ctx context.Context, invoiceNumber string) (string, error) {
	cacheKey := fmt.Sprintf("invoice:%s", invoiceNumber)
	billingID, err := r.invoiceNumberBillingCache.Get(ctx, cacheKey)
	if err != nil || billingID == nil {
		return "", err
	}
	return *billingID, nil
}
