package handler

import (
	"context"
	"ncobase/plugin/payment/event"
	"ncobase/plugin/payment/service"
	"ncobase/plugin/payment/wrapper"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// EventHandlerInterface defines the interface for event handler operations
type EventHandlerInterface interface {
	GetHandlers() map[string]event.Handler
	RefreshDependencies()
}

// eventHandler provides event handlers for the payment module
type eventHandler struct {
	service *service.Service

	usw *wrapper.UserServiceWrapper
	tsw *wrapper.SpaceServiceWrapper
}

// NewEventProvider creates a new event handler provider
func NewEventProvider(
	em ext.ManagerInterface,
	service *service.Service,
) EventHandlerInterface {
	usw := wrapper.NewUserServiceWrapper(em)
	tsw := wrapper.NewSpaceServiceWrapper(em)
	return &eventHandler{
		service: service,
		usw:     usw,
		tsw:     tsw,
	}
}

// RefreshDependencies refreshes the dependencies of the subscriber
func (e *eventHandler) RefreshDependencies() {
	e.usw.RefreshServices()
	e.tsw.RefreshServices()
}

// Payment event logging and type assertion
func (e *eventHandler) handlePaymentEvent(ctx context.Context, eventName string, data any, handler func(ctx context.Context, data *event.PaymentEventData)) {
	logger.Infof(ctx, "Processing %s event", eventName)

	eventData, ok := data.(*event.PaymentEventData)
	if !ok {
		logger.Error(ctx, "Invalid payment event data format")
		return
	}

	logger.Infof(ctx, "Payment event details: OrderID=%s, Amount=%.2f %s",
		eventData.OrderID, eventData.Amount, eventData.Currency)

	// Call the specific handler function
	handler(ctx, eventData)
}

// Subscription events
func (e *eventHandler) handleSubscriptionEvent(ctx context.Context, eventName string, data any, handler func(ctx context.Context, data *event.SubscriptionEventData)) {
	logger.Infof(ctx, "Processing %s event", eventName)

	eventData, ok := data.(*event.SubscriptionEventData)
	if !ok {
		logger.Error(ctx, "Invalid subscription event data format")
		return
	}

	logger.Infof(ctx, "Subscription event details: SubscriptionID=%s, Status=%s",
		eventData.SubscriptionID, eventData.Status)

	// Call the specific handler function
	handler(ctx, eventData)
}

// Product events
func (e *eventHandler) handleProductEvent(ctx context.Context, eventName string, data any, handler func(ctx context.Context, data *event.ProductEventData)) {
	logger.Infof(ctx, "Processing %s event", eventName)

	eventData, ok := data.(*event.ProductEventData)
	if !ok {
		logger.Error(ctx, "Invalid product event data format")
		return
	}

	logger.Infof(ctx, "Product event details: ProductID=%s, Name=%s",
		eventData.ProductID, eventData.Name)

	// Call the specific handler function
	handler(ctx, eventData)
}

// Channel events
func (e *eventHandler) handleChannelEvent(ctx context.Context, eventName string, data any, handler func(ctx context.Context, data *event.ChannelEventData)) {
	logger.Infof(ctx, "Processing %s event", eventName)

	eventData, ok := data.(*event.ChannelEventData)
	if !ok {
		logger.Error(ctx, "Invalid channel event data format")
		return
	}

	logger.Infof(ctx, "Channel event details: ChannelID=%s, Provider=%s",
		eventData.ChannelID, eventData.Provider)

	// Call the specific handler function
	handler(ctx, eventData)
}

// GetHandlers returns a map of event handlers
func (e *eventHandler) GetHandlers() map[string]event.Handler {
	return map[string]event.Handler{
		// Payment events
		"payment_created": func(data any) {
			e.handlePaymentEvent(context.TODO(), "payment.created", data, e.processPaymentCreated)
		},
		"payment_succeeded": func(data any) {
			e.handlePaymentEvent(context.TODO(), "payment.succeeded", data, e.processPaymentSucceeded)
		},
		"payment_failed": func(data any) {
			e.handlePaymentEvent(context.TODO(), "payment.failed", data, e.processPaymentFailed)
		},
		"payment_cancelled": func(data any) {
			e.handlePaymentEvent(context.TODO(), "payment.cancelled", data, e.processPaymentCancelled)
		},
		"payment_expired": func(data any) {
			e.handlePaymentEvent(context.TODO(), "payment.expired", data, e.processPaymentExpired)
		},
		"payment_refunded": func(data any) {
			e.handlePaymentEvent(context.TODO(), "payment.refunded", data, e.processPaymentRefunded)
		},

		// Subscription events
		"subscription_created": func(data any) {
			e.handleSubscriptionEvent(context.TODO(), "subscription.created", data, e.processSubscriptionCreated)
		},
		"subscription_renewed": func(data any) {
			e.handleSubscriptionEvent(context.TODO(), "subscription.renewed", data, e.processSubscriptionRenewed)
		},
		"subscription_updated": func(data any) {
			e.handleSubscriptionEvent(context.TODO(), "subscription.updated", data, e.processSubscriptionUpdated)
		},
		"subscription_cancelled": func(data any) {
			e.handleSubscriptionEvent(context.TODO(), "subscription.cancelled", data, e.processSubscriptionCancelled)
		},
		"subscription_expired": func(data any) {
			e.handleSubscriptionEvent(context.TODO(), "subscription.expired", data, e.processSubscriptionExpired)
		},

		// Product events
		"product_created": func(data any) {
			e.handleProductEvent(context.TODO(), "product.created", data, e.processProductCreated)
		},
		"product_updated": func(data any) {
			e.handleProductEvent(context.TODO(), "product.updated", data, e.processProductUpdated)
		},
		"product_deleted": func(data any) {
			e.handleProductEvent(context.TODO(), "product.deleted", data, e.processProductDeleted)
		},

		// Channel events
		"channel_created": func(data any) {
			e.handleChannelEvent(context.TODO(), "channel.created", data, e.processChannelCreated)
		},
		"channel_updated": func(data any) {
			e.handleChannelEvent(context.TODO(), "channel.updated", data, e.processChannelUpdated)
		},
		"channel_deleted": func(data any) {
			e.handleChannelEvent(context.TODO(), "channel.deleted", data, e.processChannelDeleted)
		},
		"channel_activated": func(data any) {
			e.handleChannelEvent(context.TODO(), "channel.activated", data, e.processChannelActivated)
		},
		"channel_disabled": func(data any) {
			e.handleChannelEvent(context.TODO(), "channel.disabled", data, e.processChannelDisabled)
		},
	}
}

// Payment event specialized handlers

func (e *eventHandler) processPaymentCreated(_ context.Context, _ *event.PaymentEventData) {
	// Business logic specific to payment created
	// Send notifications
	// Update statistics
	// Trigger integrations
}

func (e *eventHandler) processPaymentSucceeded(ctx context.Context, data *event.PaymentEventData) {
	// Process subscription if this payment is for a subscription
	if data.SubscriptionID != "" {
		logger.Infof(ctx, "Updating subscription status: SubscriptionID=%s", data.SubscriptionID)
		// Call subscription service to update status
		// e.service.Subscription.UpdateStatus(...)
	}

	// Additional business logic
	// Send receipt to customer
	// Update access permissions
	// Trigger fulfillment processes
}

func (e *eventHandler) processPaymentFailed(ctx context.Context, data *event.PaymentEventData) {
	// Process subscription if this payment is for a subscription
	if data.SubscriptionID != "" {
		logger.Infof(ctx, "Updating subscription status: SubscriptionID=%s", data.SubscriptionID)
		// Call subscription service to update status
		// e.service.Subscription.UpdateStatus(...)
	}

	// Additional business logic
	// Send failure notification to customer
	// Schedule retry attempts
	// Update customer risk profile
}

func (e *eventHandler) processPaymentCancelled(ctx context.Context, data *event.PaymentEventData) {
	// Process subscription if this payment is for a subscription
	if data.SubscriptionID != "" {
		logger.Infof(ctx, "Updating subscription status: SubscriptionID=%s", data.SubscriptionID)
		// Call subscription service to update status
		// e.service.Subscription.UpdateStatus(...)
	}

	// Additional business logic
	// Send cancellation notification
	// Release reserved inventory
	// Update analytics
}

func (e *eventHandler) processPaymentExpired(ctx context.Context, data *event.PaymentEventData) {
	// Process subscription if this payment is for a subscription
	if data.SubscriptionID != "" {
		logger.Infof(ctx, "Updating subscription status: SubscriptionID=%s", data.SubscriptionID)
		// Call subscription service to update status
		// e.service.Subscription.UpdateStatus(...)
	}

	// Additional business logic
	// Send expiration notification
	// Release reserved inventory
	// Create abandoned cart analytics
}

func (e *eventHandler) processPaymentRefunded(ctx context.Context, data *event.PaymentEventData) {
	// Process subscription if this payment is for a subscription
	if data.SubscriptionID != "" {
		logger.Infof(ctx, "Updating subscription status: SubscriptionID=%s", data.SubscriptionID)
		// Call subscription service to update status
		// e.service.Subscription.UpdateStatus(...)
	}

	// Additional business logic
	// Send refund confirmation
	// Update inventory if needed
	// Adjust financial records
	// Track refund analytics
}

// Subscription event specialized handlers

func (e *eventHandler) processSubscriptionCreated(ctx context.Context, data *event.SubscriptionEventData) {
	// Business logic specific to subscription created
	// Send welcome email to subscriber
	// Grant access to subscription benefits
	// Add user to relevant orgs/roles
	// Initialize usage metrics

	// If user service is available, update user's subscriptions
	if e.usw != nil {
		logger.Infof(ctx, "Updating user subscriptions: UserID=%s", data.UserID)
		// Example: e.userService.UpdateUserSubscriptions(ctx, data.UserID, data.SubscriptionID)
	}
}

func (e *eventHandler) processSubscriptionRenewed(_ context.Context, _ *event.SubscriptionEventData) {
	// Business logic specific to subscription renewed
	// Send renewal confirmation email
	// Update subscription period in other systems
	// Record renewal for analytics/reporting
	// Apply any renewal bonuses/rewards
}

func (e *eventHandler) processSubscriptionUpdated(_ context.Context, _ *event.SubscriptionEventData) {
	// Business logic specific to subscription updated
	// Send update notification email
	// Update permissions if plan changed
	// Adjust usage limits if applicable
	// Sync subscription details with other systems
}

func (e *eventHandler) processSubscriptionCancelled(_ context.Context, _ *event.SubscriptionEventData) {
	// Business logic specific to subscription cancelled
	// Send cancellation confirmation
	// Record cancellation reason for analytics
	// Schedule access removal at period end
	// Attempt win-back action (e.g., special offer)
}

func (e *eventHandler) processSubscriptionExpired(ctx context.Context, data *event.SubscriptionEventData) {
	// Business logic specific to subscription expired
	// Remove user access to premium features
	// Send expiration notification and reactivation offer
	// Update user status in other systems
	// Clean up related resources if needed

	// If user service is available, update user's permissions
	if e.usw != nil {
		logger.Infof(ctx, "Updating user permissions: UserID=%s", data.UserID)
		// Example: e.userService.RemoveSubscriptionAccess(ctx, data.UserID, data.ProductID)
	}
}

// Product event specialized handlers

func (e *eventHandler) processProductCreated(_ context.Context, _ *event.ProductEventData) {
	// Business logic specific to product created
	// Update product catalog cache
	// Notify marketing team of new product
	// Syndicate product data to external systems
	// Initialize analytics for new product
}

func (e *eventHandler) processProductUpdated(_ context.Context, _ *event.ProductEventData) {
	// Business logic specific to product updated
	// Update product cache
	// Notify users of price changes if applicable
	// Update product information in external systems
	// Log price history for analytics
}

func (e *eventHandler) processProductDeleted(_ context.Context, _ *event.ProductEventData) {
	// Business logic specific to product deleted
	// Remove product from catalogs and caches
	// Notify users who have purchased the product
	// Handle existing subscriptions to this product
	// Archive product data for reporting
}

// Channel event specialized handlers

func (e *eventHandler) processChannelCreated(_ context.Context, _ *event.ChannelEventData) {
	// Business logic specific to channel created
	// Initialize channel in monitoring systems
	// Update available payment methods in UI
	// Notify admin team of new payment channel
	// Register channel with external analytics systems
}

func (e *eventHandler) processChannelUpdated(_ context.Context, _ *event.ChannelEventData) {
	// Business logic specific to channel updated
	// Update channel information in caches
	// Refresh payment method displays
	// Log changes for audit purposes
	// Update external systems with new configuration
}

func (e *eventHandler) processChannelDeleted(_ context.Context, _ *event.ChannelEventData) {
	// Business logic specific to channel deleted
	// Remove channel from available payment methods
	// Notify affected merchants/admins
	// Update default channel fallbacks
	// Archive channel configuration for reporting
}

func (e *eventHandler) processChannelActivated(_ context.Context, _ *event.ChannelEventData) {
	// Business logic specific to channel activated
	// Add channel to active payment methods
	// Send notification to admins
	// Update payment method display in checkout
	// Initialize monitoring for the channel
}

func (e *eventHandler) processChannelDisabled(_ context.Context, _ *event.ChannelEventData) {
	// Business logic specific to channel disabled
	// Remove channel from active payment methods
	// Notify users who have used this channel recently
	// Redirect affected recurring payments if needed
	// Update analytics and monitoring systems
}
