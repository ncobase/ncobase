package repository

import (
	"context"
	"fmt"
	"ncobase/core/payment/data"
	"ncobase/core/payment/data/ent"
	paymentProductEnt "ncobase/core/payment/data/ent/paymentproduct"
	paymentSubscriptionEnt "ncobase/core/payment/data/ent/paymentsubscription"
	"ncobase/core/payment/structs"
	"time"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"
)

// SubscriptionRepositoryInterface defines the interface for subscription repository operations
type SubscriptionRepositoryInterface interface {
	Create(ctx context.Context, subscription *structs.Subscription) (*structs.Subscription, error)
	GetByID(ctx context.Context, id string) (*structs.Subscription, error)
	Update(ctx context.Context, subscription *structs.Subscription) (*structs.Subscription, error)
	List(ctx context.Context, query *structs.SubscriptionQuery) ([]*structs.Subscription, error)
	Count(ctx context.Context, query *structs.SubscriptionQuery) (int64, error)
	GetActiveSubscriptionsForRenewal(ctx context.Context, cutoff time.Time) ([]*structs.Subscription, error)
	GetSubscriptionSummary(ctx context.Context) (*structs.SubscriptionSummary, error)
	GetActiveSubscription(ctx context.Context, userID, productID string) (*structs.Subscription, error)
}

// subscriptionRepository handles subscription persistence
type subscriptionRepository struct {
	data *data.Data
}

// NewSubscriptionRepository creates a new subscription repository
func NewSubscriptionRepository(d *data.Data) SubscriptionRepositoryInterface {
	return &subscriptionRepository{data: d}
}

// Create creates a new subscription
func (r *subscriptionRepository) Create(ctx context.Context, sub *structs.Subscription) (*structs.Subscription, error) {
	builder := r.data.EC.PaymentSubscription.Create().
		SetStatus(string(sub.Status)).
		SetUserID(sub.UserID).
		SetProductID(sub.ProductID).
		SetChannelID(sub.ChannelID).
		SetCurrentPeriodStart(sub.CurrentPeriodStart).
		SetCurrentPeriodEnd(sub.CurrentPeriodEnd)

	// Set optional fields
	if sub.TenantID != "" {
		builder.SetTenantID(sub.TenantID)
	}

	if sub.CancelAt != nil {
		builder.SetCancelAt(*sub.CancelAt)
	}

	if sub.CancelledAt != nil {
		builder.SetCancelledAt(*sub.CancelledAt)
	}

	if sub.TrialStart != nil {
		builder.SetTrialStart(*sub.TrialStart)
	}

	if sub.TrialEnd != nil {
		builder.SetTrialEnd(*sub.TrialEnd)
	}

	if sub.ProviderRef != "" {
		builder.SetProviderRef(sub.ProviderRef)
	}

	if validator.IsNotEmpty(sub.Metadata) {
		builder.SetExtras(sub.Metadata)
	}

	// Create subscription
	row, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	// Convert to struct
	result, err := r.entToStruct(row)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetByID gets a subscription by ID
func (r *subscriptionRepository) GetByID(ctx context.Context, id string) (*structs.Subscription, error) {
	sub, err := r.data.EC.PaymentSubscription.Query().
		Where(paymentSubscriptionEnt.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return r.entToStruct(sub)
}

// Update updates a subscription
func (r *subscriptionRepository) Update(ctx context.Context, sub *structs.Subscription) (*structs.Subscription, error) {
	builder := r.data.EC.PaymentSubscription.UpdateOneID(sub.ID).
		SetStatus(string(sub.Status)).
		SetCurrentPeriodStart(sub.CurrentPeriodStart).
		SetCurrentPeriodEnd(sub.CurrentPeriodEnd).
		SetUpdatedAt(sub.UpdatedAt)

	// Set optional fields
	if sub.CancelAt != nil {
		builder.SetCancelAt(*sub.CancelAt)
	} else {
		builder.ClearCancelAt()
	}

	if sub.CancelledAt != nil {
		builder.SetCancelledAt(*sub.CancelledAt)
	} else {
		builder.ClearCancelledAt()
	}

	if sub.TrialStart != nil {
		builder.SetTrialStart(*sub.TrialStart)
	} else {
		builder.ClearTrialStart()
	}

	if sub.TrialEnd != nil {
		builder.SetTrialEnd(*sub.TrialEnd)
	} else {
		builder.ClearTrialEnd()
	}

	if sub.ProviderRef != "" {
		builder.SetProviderRef(sub.ProviderRef)
	}

	if validator.IsNotEmpty(sub.Metadata) {
		builder.SetExtras(sub.Metadata)
	}

	// Update subscription
	row, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	// Convert to struct
	result, err := r.entToStruct(row)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// List lists subscriptions with pagination
func (r *subscriptionRepository) List(ctx context.Context, query *structs.SubscriptionQuery) ([]*structs.Subscription, error) {
	// Build query
	q := r.data.EC.PaymentSubscription.Query()

	// Apply filters
	if query.Status != "" {
		q = q.Where(paymentSubscriptionEnt.Status(string(query.Status)))
	}

	if query.UserID != "" {
		q = q.Where(paymentSubscriptionEnt.UserID(query.UserID))
	}

	if query.TenantID != "" {
		q = q.Where(paymentSubscriptionEnt.TenantID(query.TenantID))
	}

	if query.ProductID != "" {
		q = q.Where(paymentSubscriptionEnt.ProductID(query.ProductID))
	}

	if query.ChannelID != "" {
		q = q.Where(paymentSubscriptionEnt.ChannelID(query.ChannelID))
	}

	if query.Active != nil {
		if *query.Active {
			// Active subscriptions: status is active and not expired
			now := time.Now()
			q = q.Where(
				paymentSubscriptionEnt.Status(string(structs.SubscriptionStatusActive)),
				paymentSubscriptionEnt.CurrentPeriodEndGT(now),
			)
		} else {
			// Inactive subscriptions: status is not active or expired
			now := time.Now()
			q = q.Where(
				paymentSubscriptionEnt.Or(
					paymentSubscriptionEnt.StatusNEQ(string(structs.SubscriptionStatusActive)),
					paymentSubscriptionEnt.CurrentPeriodEndLT(now),
				),
			)
		}
	}

	// Apply cursor pagination
	if query.Cursor != "" {
		id, timestamp, err := paging.DecodeCursor(query.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %v", err)
		}

		if !nanoid.IsPrimaryKey(id) {
			return nil, fmt.Errorf("invalid id in cursor: %s", id)
		}

		// Default direction is forward
		q.Where(
			paymentSubscriptionEnt.Or(
				paymentSubscriptionEnt.CreatedAtLT(timestamp),
				paymentSubscriptionEnt.And(
					paymentSubscriptionEnt.CreatedAtEQ(timestamp),
					paymentSubscriptionEnt.IDLT(id),
				),
			),
		)
	}

	// Set order - most recent first by default
	q.Order(ent.Desc(paymentSubscriptionEnt.FieldCreatedAt), ent.Desc(paymentSubscriptionEnt.FieldID))

	// Set limit
	if query.PageSize > 0 {
		q.Limit(query.PageSize)
	} else {
		q.Limit(20) // Default page size
	}

	// Execute query
	rows, err := q.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	// Convert to structs
	var subscriptions []*structs.Subscription
	for _, row := range rows {
		subscription, err := r.entToStruct(row)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, subscription)
	}

	return subscriptions, nil
}

// Count counts subscriptions
func (r *subscriptionRepository) Count(ctx context.Context, query *structs.SubscriptionQuery) (int64, error) {
	// Build query without cursor, limit, and order
	q := r.data.EC.PaymentSubscription.Query()

	// Apply filters
	if query.Status != "" {
		q = q.Where(paymentSubscriptionEnt.Status(string(query.Status)))
	}

	if query.UserID != "" {
		q = q.Where(paymentSubscriptionEnt.UserID(query.UserID))
	}

	if query.TenantID != "" {
		q = q.Where(paymentSubscriptionEnt.TenantID(query.TenantID))
	}

	if query.ProductID != "" {
		q = q.Where(paymentSubscriptionEnt.ProductID(query.ProductID))
	}

	if query.ChannelID != "" {
		q = q.Where(paymentSubscriptionEnt.ChannelID(query.ChannelID))
	}

	if query.Active != nil {
		if *query.Active {
			// Active subscriptions: status is active and not expired
			now := time.Now()
			q = q.Where(
				paymentSubscriptionEnt.Status(string(structs.SubscriptionStatusActive)),
				paymentSubscriptionEnt.CurrentPeriodEndGT(now),
			)
		} else {
			// Inactive subscriptions: status is not active or expired
			now := time.Now()
			q = q.Where(
				paymentSubscriptionEnt.Or(
					paymentSubscriptionEnt.StatusNEQ(string(structs.SubscriptionStatusActive)),
					paymentSubscriptionEnt.CurrentPeriodEndLT(now),
				),
			)
		}
	}

	// Execute count
	count, err := q.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count subscriptions: %w", err)
	}

	return int64(count), nil
}

// GetSubscriptionSummary calculates subscription statistics
func (r *subscriptionRepository) GetSubscriptionSummary(ctx context.Context) (*structs.SubscriptionSummary, error) {
	// Get past due count (continued)
	pastDueCount, err := r.data.EC.PaymentSubscription.Query().
		Where(
			paymentSubscriptionEnt.Status(string(structs.SubscriptionStatusPastDue)),
		).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get past due subscription count: %w", err)
	}

	// Calculate MRR and ARR
	// MRR (Monthly Recurring Revenue): sum(monthly price * active subscriptions)
	var mrr float64
	var arr float64

	// Get active subscriptions with products
	activeSubscriptions, err := r.data.EC.PaymentSubscription.Query().
		Where(
			paymentSubscriptionEnt.Status(string(structs.SubscriptionStatusActive)),
			paymentSubscriptionEnt.CurrentPeriodEndGT(time.Now()),
		).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active subscriptions: %w", err)
	}

	for _, sub := range activeSubscriptions {
		// Get the product for this subscription
		product, err := r.data.EC.PaymentProduct.Query().
			Where(paymentProductEnt.ID(sub.ProductID)).
			Only(ctx)
		if err != nil {
			continue // Skip if product not found
		}

		// Calculate monthly price based on billing interval
		var monthlyPrice float64
		switch product.BillingInterval {
		case string(structs.BillingIntervalDaily):
			monthlyPrice = product.Price * 30 // Approximate
		case string(structs.BillingIntervalWeekly):
			monthlyPrice = product.Price * 4.3 // Approximate
		case string(structs.BillingIntervalMonthly):
			monthlyPrice = product.Price
		case string(structs.BillingIntervalYearly):
			monthlyPrice = product.Price / 12
		default:
			monthlyPrice = product.Price // Default to assuming it's monthly
		}

		// Add to MRR
		mrr += monthlyPrice
	}

	// ARR is MRR * 12
	arr = mrr * 12

	// Create summary
	summary := &structs.SubscriptionSummary{
		// TotalCount:     int64(totalCount),
		// ActiveCount:    int64(activeCount),
		// TrialingCount:  int64(trialingCount),
		// CancelledCount: int64(cancelledCount),
		// ExpiredCount:   int64(expiredCount),
		PastDueCount: int64(pastDueCount),
		MRR:          mrr,
		ARR:          arr,
	}

	return summary, nil
}

// entToStruct converts an Ent Subscription to a structs.Subscription
func (r *subscriptionRepository) entToStruct(sub *ent.PaymentSubscription) (*structs.Subscription, error) {
	// Convert optional time fields
	var cancelAt *time.Time
	if !sub.CancelAt.IsZero() {
		cancelAt = sub.CancelAt
	}

	var cancelledAt *time.Time
	if !sub.CancelledAt.IsZero() {
		cancelledAt = sub.CancelledAt
	}

	var trialStart *time.Time
	if !sub.TrialStart.IsZero() {
		trialStart = sub.TrialStart
	}

	var trialEnd *time.Time
	if !sub.TrialEnd.IsZero() {
		trialEnd = sub.TrialEnd
	}

	return &structs.Subscription{
		Status:             structs.SubscriptionStatus(sub.Status),
		UserID:             sub.UserID,
		TenantID:           sub.TenantID,
		ProductID:          sub.ProductID,
		ChannelID:          sub.ChannelID,
		CurrentPeriodStart: sub.CurrentPeriodStart,
		CurrentPeriodEnd:   sub.CurrentPeriodEnd,
		CancelAt:           cancelAt,
		CancelledAt:        cancelledAt,
		TrialStart:         trialStart,
		TrialEnd:           trialEnd,
		ProviderRef:        sub.ProviderRef,
		Metadata:           sub.Extras,
	}, nil
}

// GetActiveSubscriptionsForRenewal gets active subscriptions due for renewal
func (r *subscriptionRepository) GetActiveSubscriptionsForRenewal(ctx context.Context, cutoff time.Time) ([]*structs.Subscription, error) {
	// Get subscriptions that are active, not cancelled, and ending before the cutoff time
	entSubscriptions, err := r.data.EC.PaymentSubscription.Query().
		Where(
			paymentSubscriptionEnt.Status(string(structs.SubscriptionStatusActive)),
			paymentSubscriptionEnt.CancelAtIsNil(),
			paymentSubscriptionEnt.CurrentPeriodEndLT(cutoff),
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions for renewal: %w", err)
	}

	// Convert to structs
	var subscriptions []*structs.Subscription
	for _, sub := range entSubscriptions {
		sub, err := r.entToStruct(sub)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

// GetByUser gets subscriptions for a user
func (r *subscriptionRepository) GetByUser(ctx context.Context, userID string, active bool) ([]*structs.Subscription, error) {
	// Build query
	q := r.data.EC.PaymentSubscription.Query().
		Where(paymentSubscriptionEnt.UserID(userID))

	// Apply active filter if requested
	if active {
		now := time.Now()
		q = q.Where(
			paymentSubscriptionEnt.Status(string(structs.SubscriptionStatusActive)),
			paymentSubscriptionEnt.CurrentPeriodEndGT(now),
		)
	}

	// Execute query
	entSubscriptions, err := q.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user subscriptions: %w", err)
	}

	// Convert to structs
	var subscriptions []*structs.Subscription
	for _, sub := range entSubscriptions {
		sub, err := r.entToStruct(sub)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

// GetActiveSubscription gets the active subscription for a user and product
func (r *subscriptionRepository) GetActiveSubscription(ctx context.Context, userID, productID string) (*structs.Subscription, error) {
	now := time.Now()
	sub, err := r.data.EC.PaymentSubscription.Query().
		Where(
			paymentSubscriptionEnt.UserID(userID),
			paymentSubscriptionEnt.ProductID(productID),
			paymentSubscriptionEnt.Status(string(structs.SubscriptionStatusActive)),
			paymentSubscriptionEnt.CurrentPeriodEndGT(now),
		).
		Only(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get active subscription: %w", err)
	}

	return r.entToStruct(sub)
}
