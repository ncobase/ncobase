package service

import (
	"context"
	"fmt"
	"ncobase/plugin/payment/data/repository"
	"ncobase/plugin/payment/event"
	"ncobase/plugin/payment/structs"
	"time"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils/nanoid"
)

// SubscriptionServiceInterface defines the interface for subscription service operations
type SubscriptionServiceInterface interface {
	Create(ctx context.Context, input *structs.CreateSubscriptionInput) (*structs.Subscription, error)
	GetByID(ctx context.Context, id string) (*structs.Subscription, error)
	Update(ctx context.Context, id string, updates map[string]any) (*structs.Subscription, error)
	Cancel(ctx context.Context, id string, immediate bool, reason string) (*structs.Subscription, error)
	List(ctx context.Context, query *structs.SubscriptionQuery) (paging.Result[*structs.Subscription], error)
	GetByUser(ctx context.Context, query *structs.SubscriptionQuery) (paging.Result[*structs.Subscription], error)
	CheckUserHasActiveSubscription(ctx context.Context, userID, productID string) (bool, error)
	ProcessRenewals(ctx context.Context) error
	ProcessExpiredSubscriptions(ctx context.Context) error
	GetSubscriptionStats(ctx context.Context) (*structs.SubscriptionSummary, error)
	Serialize(subscription *structs.Subscription) *structs.Subscription
	Serializes(subscriptions []*structs.Subscription) []*structs.Subscription
}

// subscriptionService provides operations for subscriptions
type subscriptionService struct {
	repo        repository.SubscriptionRepositoryInterface
	productRepo repository.ProductRepositoryInterface
	channelRepo repository.ChannelRepositoryInterface
	orderRepo   repository.OrderRepositoryInterface
	publisher   event.PublisherInterface
	providerSvc ProviderServiceInterface
}

// NewSubscriptionService creates a new subscription service
func NewSubscriptionService(
	repo repository.SubscriptionRepositoryInterface,
	productRepo repository.ProductRepositoryInterface,
	channelRepo repository.ChannelRepositoryInterface,
	orderRepo repository.OrderRepositoryInterface,
	publisher event.PublisherInterface,
	providerSvc ProviderServiceInterface,
) SubscriptionServiceInterface {
	return &subscriptionService{
		repo:        repo,
		productRepo: productRepo,
		channelRepo: channelRepo,
		orderRepo:   orderRepo,
		publisher:   publisher,
		providerSvc: providerSvc,
	}
}

// Create creates a new subscription
func (s *subscriptionService) Create(ctx context.Context, input *structs.CreateSubscriptionInput) (*structs.Subscription, error) {
	// Validate input
	if input.UserID == "" {
		return nil, fmt.Errorf(ecode.FieldIsRequired("user_id"))
	}

	if input.ProductID == "" {
		return nil, fmt.Errorf(ecode.FieldIsRequired("product_id"))
	}

	if input.ChannelID == "" {
		return nil, fmt.Errorf(ecode.FieldIsRequired("channel_id"))
	}

	// Get the product
	product, err := s.productRepo.GetByID(ctx, input.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Check if the product supports subscriptions
	if product.PricingType != structs.PricingTypeRecurring {
		return nil, fmt.Errorf("product does not support subscriptions")
	}

	// Get the channel
	channel, err := s.channelRepo.GetByID(ctx, input.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment channel: %w", err)
	}

	// Check if the channel is active
	if channel.Status != structs.ChannelStatusActive {
		return nil, fmt.Errorf("payment channel is not active")
	}

	// Check if the channel supports subscriptions
	supportsSubscriptions := false
	for _, t := range channel.SupportedType {
		if t == structs.PaymentTypeSubscription {
			supportsSubscriptions = true
			break
		}
	}

	if !supportsSubscriptions {
		return nil, fmt.Errorf("payment channel does not support subscriptions")
	}

	// Get payment provider
	paymentProvider, err := s.providerSvc.GetProvider(channel.Provider, channel.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment provider: %w", err)
	}

	// Determine trial period
	now := time.Now()
	trialDays := product.TrialDays
	if input.TrialDays > 0 {
		trialDays = input.TrialDays
	}

	var trialStart, trialEnd *time.Time
	var currentPeriodStart, currentPeriodEnd time.Time

	if trialDays > 0 {
		ts := now
		trialStart = &ts
		te := now.AddDate(0, 0, trialDays)
		trialEnd = &te
		currentPeriodStart = now
		currentPeriodEnd = te
	} else {
		currentPeriodStart = now
		currentPeriodEnd = s.calculateNextBillingDate(now, product.BillingInterval)
	}

	// Create subscription entity
	subscription := &structs.Subscription{
		Status:             structs.SubscriptionStatusActive,
		UserID:             input.UserID,
		SpaceID:            input.SpaceID,
		ProductID:          input.ProductID,
		ChannelID:          input.ChannelID,
		CurrentPeriodStart: currentPeriodStart,
		CurrentPeriodEnd:   currentPeriodEnd,
		TrialStart:         trialStart,
		TrialEnd:           trialEnd,
		Metadata:           input.Metadata,
	}

	// Create subscription with provider if it's not a trial
	if trialDays == 0 {
		providerRef, _, err := paymentProvider.CreateSubscription(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to create subscription with provider: %w", err)
		}
		subscription.ProviderRef = providerRef
	}

	// Save to database
	created, err := s.repo.Create(ctx, subscription)
	if err != nil {
		return nil, handleEntError(ctx, "Subscription", err)
	}

	// If it's a trial, no payment is required
	// If it's not a trial, create a payment order
	if trialDays == 0 {
		// Create payment order for the first billing period
		orderInput := &structs.CreateOrderInput{
			OrderNumber:    time.Now().Format("20060102150405") + nanoid.String(6),
			Amount:         product.Price,
			Currency:       product.Currency,
			Status:         structs.PaymentStatusPending,
			Type:           structs.PaymentTypeSubscription,
			ChannelID:      input.ChannelID,
			UserID:         input.UserID,
			SpaceID:        input.SpaceID,
			ProductID:      input.ProductID,
			SubscriptionID: created.ID,
			ExpiresAt:      now.Add(24 * time.Hour), // 24 hours to pay
			Description:    fmt.Sprintf("Subscription to %s", product.Name),
			Metadata: map[string]any{
				"subscription_id":      created.ID,
				"billing_period_start": currentPeriodStart.Format(time.RFC3339),
				"billing_period_end":   currentPeriodEnd.Format(time.RFC3339),
			},
		}

		_, err = s.orderRepo.Create(ctx, orderInput)
		if err != nil {
			logger.Warnf(ctx, "Failed to create order for subscription: %v", err)
			// Continue even if order creation fails
		}
	}

	// Publish event using the new event system
	if s.publisher != nil {
		// Create event data
		eventData := event.NewSubscriptionEventData(
			created.ID,
			created.UserID,
			created.SpaceID,
			created.ProductID,
			created.ChannelID,
			channel.Provider,
			created.Status,
			created.CurrentPeriodStart,
			created.CurrentPeriodEnd,
			created.Metadata,
		)

		// Publish event
		s.publisher.PublishSubscriptionCreated(ctx, eventData)
	}

	return s.Serialize(created), nil
}

// GetByID gets a subscription by ID
func (s *subscriptionService) GetByID(ctx context.Context, id string) (*structs.Subscription, error) {
	if id == "" {
		return nil, fmt.Errorf("subscription ID is required")
	}

	return s.repo.GetByID(ctx, id)
}

// Update updates a subscription
func (s *subscriptionService) Update(ctx context.Context, id string, updates map[string]any) (*structs.Subscription, error) {
	if id == "" {
		return nil, fmt.Errorf("subscription ID is required")
	}

	// Get existing subscription
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates to subscription entity
	if status, ok := updates["status"].(string); ok {
		existing.Status = structs.SubscriptionStatus(status)
	}

	if periodStart, ok := updates["current_period_start"].(time.Time); ok {
		existing.CurrentPeriodStart = periodStart
	}

	if periodEnd, ok := updates["current_period_end"].(time.Time); ok {
		existing.CurrentPeriodEnd = periodEnd
	}

	if cancelAt, ok := updates["cancel_at"].(time.Time); ok {
		existing.CancelAt = &cancelAt
	}

	if cancelledAt, ok := updates["cancelled_at"].(time.Time); ok {
		existing.CancelledAt = &cancelledAt
	}

	if trialStart, ok := updates["trial_start"].(time.Time); ok {
		existing.TrialStart = &trialStart
	}

	if trialEnd, ok := updates["trial_end"].(time.Time); ok {
		existing.TrialEnd = &trialEnd
	}

	if providerRef, ok := updates["provider_ref"].(string); ok {
		existing.ProviderRef = providerRef
	}

	if metadata, ok := updates["metadata"].(map[string]any); ok {
		existing.Metadata = metadata
	}

	existing.UpdatedAt = time.Now().UnixMilli()

	// If the subscription has a provider reference and it's not a trial,
	// update with the provider
	if existing.ProviderRef != "" && (existing.TrialEnd == nil || time.Now().After(*existing.TrialEnd)) {
		// Get the channel
		channel, err := s.channelRepo.GetByID(ctx, existing.ChannelID)
		if err != nil {
			return nil, fmt.Errorf("failed to get payment channel: %w", err)
		}

		// Get payment provider
		paymentProvider, err := s.providerSvc.GetProvider(channel.Provider, channel.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to get payment provider: %w", err)
		}

		// Update subscription with provider
		err = paymentProvider.UpdateSubscription(ctx, existing.ProviderRef, updates)
		if err != nil {
			return nil, fmt.Errorf("failed to update subscription with provider: %w", err)
		}
	}

	// Save to database
	updated, err := s.repo.Update(ctx, existing)
	if err != nil {
		return nil, err
	}

	// Publish event
	if s.publisher != nil {
		// Get the channel for event data
		channel, err := s.channelRepo.GetByID(ctx, updated.ChannelID)
		if err == nil {
			eventData := &event.SubscriptionEventData{
				SubscriptionID:     updated.ID,
				UserID:             updated.UserID,
				SpaceID:            updated.SpaceID,
				ProductID:          updated.ProductID,
				ChannelID:          updated.ChannelID,
				Provider:           channel.Provider,
				Status:             updated.Status,
				CurrentPeriodStart: updated.CurrentPeriodStart,
				CurrentPeriodEnd:   updated.CurrentPeriodEnd,
				Metadata:           updated.Metadata,
			}

			s.publisher.PublishSubscriptionUpdated(ctx, eventData)
		}
	}

	return updated, nil
}

// Cancel cancels a subscription
func (s *subscriptionService) Cancel(ctx context.Context, id string, immediate bool, reason string) (*structs.Subscription, error) {
	if id == "" {
		return nil, fmt.Errorf(ecode.FieldIsRequired("id"))
	}

	// Get existing subscription
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, handleEntError(ctx, "Subscription", err)
	}

	// Cannot cancel an already cancelled or expired subscription
	if existing.Status == structs.SubscriptionStatusCancelled || existing.Status == structs.SubscriptionStatusExpired {
		return nil, fmt.Errorf("subscription is already cancelled or expired")
	}

	// Get the channel for event publishing
	channel, err := s.channelRepo.GetByID(ctx, existing.ChannelID)
	if err != nil {
		logger.Warnf(ctx, "Failed to get payment channel for subscription cancellation: %v", err)
		// Continue even if channel fetch fails
	}

	now := time.Now()

	// If immediate cancellation, set status to cancelled and set cancelled_at
	// If not immediate, set cancel_at to the end of the current period
	if immediate {
		existing.Status = structs.SubscriptionStatusCancelled
		existing.CancelledAt = &now
	} else {
		existing.CancelAt = &existing.CurrentPeriodEnd
	}

	// Store reason in metadata
	if existing.Metadata == nil {
		existing.Metadata = make(map[string]any)
	}
	existing.Metadata["cancellation_reason"] = reason

	existing.UpdatedAt = now.UnixMilli()

	// If the subscription has a provider reference, cancel with the provider
	if existing.ProviderRef != "" && channel != nil {
		// Get payment provider
		paymentProvider, err := s.providerSvc.GetProvider(channel.Provider, channel.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to get payment provider: %w", err)
		}

		// Cancel subscription with provider
		err = paymentProvider.CancelSubscription(ctx, existing.ProviderRef, immediate)
		if err != nil {
			return nil, fmt.Errorf("failed to cancel subscription with provider: %w", err)
		}
	}

	// Save to database
	updated, err := s.repo.Update(ctx, existing)
	if err != nil {
		return nil, handleEntError(ctx, "Subscription", err)
	}

	// Publish event
	if s.publisher != nil && channel != nil {
		// Create event data
		eventData := event.NewSubscriptionEventData(
			updated.ID,
			updated.UserID,
			updated.SpaceID,
			updated.ProductID,
			updated.ChannelID,
			channel.Provider,
			updated.Status,
			updated.CurrentPeriodStart,
			updated.CurrentPeriodEnd,
			updated.Metadata,
		)

		// Publish event
		s.publisher.PublishSubscriptionCancelled(ctx, eventData)
	}

	return s.Serialize(updated), nil
}

// List lists subscriptions
func (s *subscriptionService) List(ctx context.Context, query *structs.SubscriptionQuery) (paging.Result[*structs.Subscription], error) {
	pp := paging.Params{
		Cursor:    query.Cursor,
		Limit:     query.PageSize,
		Direction: "forward", // Default direction
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.Subscription, int, error) {
		lq := *query
		lq.Cursor = cursor
		lq.PageSize = limit

		subscriptions, err := s.repo.List(ctx, &lq)
		if err != nil {
			logger.Errorf(ctx, "Error listing subscriptions: %v", err)
			return nil, 0, err
		}

		total, err := s.repo.Count(ctx, query)
		if err != nil {
			logger.Errorf(ctx, "Error counting subscriptions: %v", err)
			return nil, 0, err
		}

		return s.Serializes(subscriptions), int(total), nil
	})
}

// GetByUser gets subscriptions for a user
func (s *subscriptionService) GetByUser(ctx context.Context, query *structs.SubscriptionQuery) (paging.Result[*structs.Subscription], error) {
	if query.UserID == "" {
		return paging.Result[*structs.Subscription]{}, fmt.Errorf(ecode.FieldIsRequired("user_id"))
	}

	return s.List(ctx, query)
}

// ProcessRenewals processes subscriptions due for renewal (continued)
func (s *subscriptionService) ProcessRenewals(ctx context.Context) error {
	// Get subscriptions due for renewal in the next 24 hours
	cutoff := time.Now().Add(24 * time.Hour)
	subscriptions, err := s.repo.GetActiveSubscriptionsForRenewal(ctx, cutoff)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions for renewal: %w", err)
	}

	logger.Infof(ctx, "Processing %d subscriptions for renewal", len(subscriptions))

	for _, subscription := range subscriptions {
		// Get the product
		product, err := s.productRepo.GetByID(ctx, subscription.ProductID)
		if err != nil {
			logger.Errorf(ctx, "Failed to get product for subscription %s: %v", subscription.ID, err)
			continue
		}

		// Calculate next billing period
		nextPeriodStart := subscription.CurrentPeriodEnd
		nextPeriodEnd := s.calculateNextBillingDate(nextPeriodStart, product.BillingInterval)

		// Create payment order for the next billing period
		orderInput := &structs.CreateOrderInput{
			OrderNumber:    time.Now().Format("20060102150405") + nanoid.String(6),
			Amount:         product.Price,
			Currency:       product.Currency,
			Status:         structs.PaymentStatusPending,
			Type:           structs.PaymentTypeSubscription,
			ChannelID:      subscription.ChannelID,
			UserID:         subscription.UserID,
			SpaceID:        subscription.SpaceID,
			ProductID:      subscription.ProductID,
			SubscriptionID: subscription.ID,
			ExpiresAt:      subscription.CurrentPeriodEnd.Add(24 * time.Hour), // 24 hours after current period ends
			Description:    fmt.Sprintf("Renewal for subscription to %s", product.Name),
			Metadata: map[string]any{
				"subscription_id":      subscription.ID,
				"billing_period_start": nextPeriodStart.Format(time.RFC3339),
				"billing_period_end":   nextPeriodEnd.Format(time.RFC3339),
				"is_renewal":           true,
			},
		}

		_, err = s.orderRepo.Create(ctx, orderInput)
		if err != nil {
			logger.Errorf(ctx, "Failed to create renewal order for subscription %s: %v", subscription.ID, err)
			continue
		}

		// Update subscription with next billing period
		// This will happen immediately, but the status will be updated based on payment result
		subscription.CurrentPeriodStart = nextPeriodStart
		subscription.CurrentPeriodEnd = nextPeriodEnd
		subscription.UpdatedAt = time.Now().UnixMilli()

		_, err = s.repo.Update(ctx, subscription)
		if err != nil {
			logger.Errorf(ctx, "Failed to update subscription %s with next billing period: %v", subscription.ID, err)
			continue
		}

		// Publish event
		if s.publisher != nil {
			// Get the channel for event data
			channel, err := s.channelRepo.GetByID(ctx, subscription.ChannelID)
			if err == nil {
				eventData := &event.SubscriptionEventData{
					SubscriptionID:     subscription.ID,
					UserID:             subscription.UserID,
					SpaceID:            subscription.SpaceID,
					ProductID:          subscription.ProductID,
					ChannelID:          subscription.ChannelID,
					Provider:           channel.Provider,
					Status:             subscription.Status,
					CurrentPeriodStart: subscription.CurrentPeriodStart,
					CurrentPeriodEnd:   subscription.CurrentPeriodEnd,
					Metadata:           subscription.Metadata,
				}

				s.publisher.PublishSubscriptionRenewed(ctx, eventData)
			}
		}

		logger.Infof(ctx, "Processed renewal for subscription %s", subscription.ID)
	}

	return nil
}

// calculateNextBillingDate calculates the next billing date based on the interval
func (s *subscriptionService) calculateNextBillingDate(start time.Time, interval structs.BillingInterval) time.Time {
	switch interval {
	case structs.BillingIntervalDaily:
		return start.AddDate(0, 0, 1)
	case structs.BillingIntervalWeekly:
		return start.AddDate(0, 0, 7)
	case structs.BillingIntervalMonthly:
		return start.AddDate(0, 1, 0)
	case structs.BillingIntervalYearly:
		return start.AddDate(1, 0, 0)
	default:
		// Default to monthly if interval is not recognized
		return start.AddDate(0, 1, 0)
	}
}

// ProcessExpiredSubscriptions processes expired subscriptions
func (s *subscriptionService) ProcessExpiredSubscriptions(ctx context.Context) error {
	// Get active subscriptions that have passed their current period end
	now := time.Now()
	query := &structs.SubscriptionQuery{
		Status: structs.SubscriptionStatusActive,
	}
	subscriptions, err := s.repo.List(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to get active subscriptions: %w", err)
	}

	var expiredSubscriptions []*structs.Subscription
	for _, sub := range subscriptions {
		if now.After(sub.CurrentPeriodEnd) {
			expiredSubscriptions = append(expiredSubscriptions, sub)
		}
	}

	logger.Infof(ctx, "Processing %d expired subscriptions", len(expiredSubscriptions))

	for _, sub := range expiredSubscriptions {
		// Update subscription status to expired
		sub.Status = structs.SubscriptionStatusExpired
		sub.UpdatedAt = now.UnixMilli()

		// Check if there's a pending payment for this subscription
		// If yes, don't expire it yet as it may be paid soon
		query := &structs.OrderQuery{
			SubscriptionID: sub.ID,
			Status:         structs.PaymentStatusPending,
		}
		pendingOrders, err := s.orderRepo.List(ctx, query)
		if err != nil {
			logger.Errorf(ctx, "Failed to check pending payments for subscription %s: %v", sub.ID, err)
			continue
		}

		if len(pendingOrders) > 0 {
			logger.Infof(ctx, "Subscription %s has pending payments, skipping expiration", sub.ID)
			continue
		}

		_, err = s.repo.Update(ctx, sub)
		if err != nil {
			logger.Errorf(ctx, "Failed to expire subscription %s: %v", sub.ID, err)
			continue
		}

		// Publish event
		if s.publisher != nil {
			// Get the channel for event data
			channel, err := s.channelRepo.GetByID(ctx, sub.ChannelID)
			if err == nil {
				eventData := &event.SubscriptionEventData{
					SubscriptionID:     sub.ID,
					UserID:             sub.UserID,
					SpaceID:            sub.SpaceID,
					ProductID:          sub.ProductID,
					ChannelID:          sub.ChannelID,
					Provider:           channel.Provider,
					Status:             sub.Status,
					CurrentPeriodStart: sub.CurrentPeriodStart,
					CurrentPeriodEnd:   sub.CurrentPeriodEnd,
					Metadata:           sub.Metadata,
				}

				s.publisher.PublishSubscriptionExpired(ctx, eventData)
			}
		}

		logger.Infof(ctx, "Expired subscription %s", sub.ID)
	}

	return nil
}

// GetSubscriptionStats gets subscription statistics
func (s *subscriptionService) GetSubscriptionStats(ctx context.Context) (*structs.SubscriptionSummary, error) {
	return s.repo.GetSubscriptionSummary(ctx)
}

// CheckUserHasActiveSubscription checks if a user has an active subscription to a product
func (s *subscriptionService) CheckUserHasActiveSubscription(ctx context.Context, userID, productID string) (bool, error) {
	if userID == "" {
		return false, fmt.Errorf("user ID is required")
	}

	if productID == "" {
		return false, fmt.Errorf("product ID is required")
	}

	// Try to get active subscription
	_, err := s.repo.GetActiveSubscription(ctx, userID, productID)
	if err != nil {
		return false, nil // No active subscription found
	}

	return true, nil
}

// Serialize serializes a subscription entity to a response format
func (s *subscriptionService) Serialize(subscription *structs.Subscription) *structs.Subscription {
	if subscription == nil {
		return nil
	}

	return &structs.Subscription{
		ID:                 subscription.ID,
		Status:             subscription.Status,
		UserID:             subscription.UserID,
		SpaceID:            subscription.SpaceID,
		ProductID:          subscription.ProductID,
		ChannelID:          subscription.ChannelID,
		CurrentPeriodStart: subscription.CurrentPeriodStart,
		CurrentPeriodEnd:   subscription.CurrentPeriodEnd,
		CancelAt:           subscription.CancelAt,
		CancelledAt:        subscription.CancelledAt,
		TrialStart:         subscription.TrialStart,
		TrialEnd:           subscription.TrialEnd,
		ProviderRef:        subscription.ProviderRef,
		Metadata:           subscription.Metadata,
		CreatedAt:          subscription.CreatedAt,
		UpdatedAt:          subscription.UpdatedAt,
	}
}

// Serializes serializes multiple subscription entities to response format
func (s *subscriptionService) Serializes(subscriptions []*structs.Subscription) []*structs.Subscription {
	result := make([]*structs.Subscription, len(subscriptions))
	for i, subscription := range subscriptions {
		result[i] = s.Serialize(subscription)
	}
	return result
}
