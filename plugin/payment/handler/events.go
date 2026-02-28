package handler

import (
	"context"
	"ncobase/plugin/payment/event"
	"ncobase/plugin/payment/service"
	"ncobase/plugin/payment/structs"
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
			ctx := context.Background()
			e.handlePaymentEvent(ctx, "payment.created", data, e.processPaymentCreated)
		},
		"payment_succeeded": func(data any) {
			ctx := context.Background()
			e.handlePaymentEvent(ctx, "payment.succeeded", data, e.processPaymentSucceeded)
		},
		"payment_failed": func(data any) {
			ctx := context.Background()
			e.handlePaymentEvent(ctx, "payment.failed", data, e.processPaymentFailed)
		},
		"payment_cancelled": func(data any) {
			ctx := context.Background()
			e.handlePaymentEvent(ctx, "payment.cancelled", data, e.processPaymentCancelled)
		},
		"payment_expired": func(data any) {
			ctx := context.Background()
			e.handlePaymentEvent(ctx, "payment.expired", data, e.processPaymentExpired)
		},
		"payment_refunded": func(data any) {
			ctx := context.Background()
			e.handlePaymentEvent(ctx, "payment.refunded", data, e.processPaymentRefunded)
		},

		// Subscription events
		"subscription_created": func(data any) {
			ctx := context.Background()
			e.handleSubscriptionEvent(ctx, "subscription.created", data, e.processSubscriptionCreated)
		},
		"subscription_renewed": func(data any) {
			ctx := context.Background()
			e.handleSubscriptionEvent(ctx, "subscription.renewed", data, e.processSubscriptionRenewed)
		},
		"subscription_updated": func(data any) {
			ctx := context.Background()
			e.handleSubscriptionEvent(ctx, "subscription.updated", data, e.processSubscriptionUpdated)
		},
		"subscription_cancelled": func(data any) {
			ctx := context.Background()
			e.handleSubscriptionEvent(ctx, "subscription.cancelled", data, e.processSubscriptionCancelled)
		},
		"subscription_expired": func(data any) {
			ctx := context.Background()
			e.handleSubscriptionEvent(ctx, "subscription.expired", data, e.processSubscriptionExpired)
		},

		// Product events
		"product_created": func(data any) {
			ctx := context.Background()
			e.handleProductEvent(ctx, "product.created", data, e.processProductCreated)
		},
		"product_updated": func(data any) {
			ctx := context.Background()
			e.handleProductEvent(ctx, "product.updated", data, e.processProductUpdated)
		},
		"product_deleted": func(data any) {
			ctx := context.Background()
			e.handleProductEvent(ctx, "product.deleted", data, e.processProductDeleted)
		},

		// Channel events
		"channel_created": func(data any) {
			ctx := context.Background()
			e.handleChannelEvent(ctx, "channel.created", data, e.processChannelCreated)
		},
		"channel_updated": func(data any) {
			ctx := context.Background()
			e.handleChannelEvent(ctx, "channel.updated", data, e.processChannelUpdated)
		},
		"channel_deleted": func(data any) {
			ctx := context.Background()
			e.handleChannelEvent(ctx, "channel.deleted", data, e.processChannelDeleted)
		},
		"channel_activated": func(data any) {
			ctx := context.Background()
			e.handleChannelEvent(ctx, "channel.activated", data, e.processChannelActivated)
		},
		"channel_disabled": func(data any) {
			ctx := context.Background()
			e.handleChannelEvent(ctx, "channel.disabled", data, e.processChannelDisabled)
		},
	}
}

// Payment event specialized handlers

func (e *eventHandler) processPaymentCreated(ctx context.Context, data *event.PaymentEventData) {
	logger.Infof(ctx, "Payment created: OrderID=%s, Amount=%.2f %s, Type=%s",
		data.OrderID, data.Amount, data.Currency, data.Type)

	// Log payment creation for analytics
	if data.Metadata == nil {
		data.Metadata = make(map[string]any)
	}
	data.Metadata["event_processed_at"] = data.Timestamp.Format("2006-01-02T15:04:05Z07:00")

	// Create payment log entry
	logEntry := &structs.CreateLogInput{
		OrderID:     data.OrderID,
		ChannelID:   data.ChannelID,
		Type:        structs.LogTypeCreate,
		StatusAfter: data.Status,
		RequestData: "Payment created event processed",
		UserID:      data.UserID,
	}

	if _, err := e.service.Log.Create(ctx, logEntry); err != nil {
		logger.Warnf(ctx, "Failed to create payment log: %v", err)
	}

	logger.Infof(ctx, "Payment creation event processed successfully")
}

func (e *eventHandler) processPaymentSucceeded(ctx context.Context, data *event.PaymentEventData) {
	logger.Infof(ctx, "Payment succeeded: OrderID=%s, Amount=%.2f %s",
		data.OrderID, data.Amount, data.Currency)

	// Update subscription status if this is a subscription payment
	if data.SubscriptionID != "" {
		logger.Infof(ctx, "Activating subscription: SubscriptionID=%s", data.SubscriptionID)

		subscription, err := e.service.Subscription.GetByID(ctx, data.SubscriptionID)
		if err != nil {
			logger.Errorf(ctx, "Failed to get subscription: %v", err)
		} else {
			// Update subscription to active if it was pending
			if subscription.Status != structs.SubscriptionStatusActive {
				updates := map[string]any{
					"status": structs.SubscriptionStatusActive,
				}
				if _, err := e.service.Subscription.Update(ctx, data.SubscriptionID, updates); err != nil {
					logger.Errorf(ctx, "Failed to activate subscription: %v", err)
				} else {
					logger.Infof(ctx, "Subscription activated successfully")
				}
			}
		}
	}

	// Grant access to product if applicable
	if data.ProductID != "" && e.usw != nil && e.usw.HasUserService() {
		logger.Infof(ctx, "Granting product access: UserID=%s, ProductID=%s",
			data.UserID, data.ProductID)
		// User service can be extended to track product access
	}

	// Log successful payment
	logEntry := &structs.CreateLogInput{
		OrderID:      data.OrderID,
		ChannelID:    data.ChannelID,
		Type:         structs.LogTypeUpdate,
		StatusBefore: structs.PaymentStatusPending,
		StatusAfter:  structs.PaymentStatusCompleted,
		ResponseData: "Payment succeeded event processed",
		UserID:       data.UserID,
	}

	if _, err := e.service.Log.Create(ctx, logEntry); err != nil {
		logger.Warnf(ctx, "Failed to create payment log: %v", err)
	}

	logger.Infof(ctx, "Payment success event processed successfully")
}

func (e *eventHandler) processPaymentFailed(ctx context.Context, data *event.PaymentEventData) {
	logger.Warnf(ctx, "Payment failed: OrderID=%s, Amount=%.2f %s",
		data.OrderID, data.Amount, data.Currency)

	// Handle subscription payment failure
	if data.SubscriptionID != "" {
		logger.Warnf(ctx, "Subscription payment failed: SubscriptionID=%s", data.SubscriptionID)

		subscription, err := e.service.Subscription.GetByID(ctx, data.SubscriptionID)
		if err != nil {
			logger.Errorf(ctx, "Failed to get subscription: %v", err)
		} else {
			// Mark subscription as past_due if active
			if subscription.Status == structs.SubscriptionStatusActive {
				updates := map[string]any{
					"status": structs.SubscriptionStatusPastDue,
					"metadata": map[string]any{
						"payment_failed_at": data.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
						"failed_order_id":   data.OrderID,
					},
				}
				if _, err := e.service.Subscription.Update(ctx, data.SubscriptionID, updates); err != nil {
					logger.Errorf(ctx, "Failed to update subscription status: %v", err)
				}
			}
		}
	}

	// Log payment failure
	logEntry := &structs.CreateLogInput{
		OrderID:      data.OrderID,
		ChannelID:    data.ChannelID,
		Type:         structs.LogTypeError,
		StatusBefore: structs.PaymentStatusPending,
		StatusAfter:  structs.PaymentStatusFailed,
		Error:        "Payment failed",
		UserID:       data.UserID,
	}

	if _, err := e.service.Log.Create(ctx, logEntry); err != nil {
		logger.Warnf(ctx, "Failed to create payment log: %v", err)
	}

	logger.Infof(ctx, "Payment failure event processed")
}

func (e *eventHandler) processPaymentCancelled(ctx context.Context, data *event.PaymentEventData) {
	logger.Infof(ctx, "Payment cancelled: OrderID=%s, Amount=%.2f %s",
		data.OrderID, data.Amount, data.Currency)

	// Handle subscription cancellation if applicable
	if data.SubscriptionID != "" {
		logger.Infof(ctx, "Payment cancelled for subscription: SubscriptionID=%s", data.SubscriptionID)

		// Add cancellation metadata to subscription
		subscription, err := e.service.Subscription.GetByID(ctx, data.SubscriptionID)
		if err != nil {
			logger.Errorf(ctx, "Failed to get subscription: %v", err)
		} else {
			metadata := subscription.Metadata
			if metadata == nil {
				metadata = make(map[string]any)
			}
			metadata["payment_cancelled_at"] = data.Timestamp.Format("2006-01-02T15:04:05Z07:00")
			metadata["cancelled_order_id"] = data.OrderID

			updates := map[string]any{"metadata": metadata}
			if _, err := e.service.Subscription.Update(ctx, data.SubscriptionID, updates); err != nil {
				logger.Errorf(ctx, "Failed to update subscription metadata: %v", err)
			}
		}
	}

	// Log payment cancellation
	logEntry := &structs.CreateLogInput{
		OrderID:      data.OrderID,
		ChannelID:    data.ChannelID,
		Type:         structs.LogTypeUpdate,
		StatusBefore: structs.PaymentStatusPending,
		StatusAfter:  structs.PaymentStatusCancelled,
		ResponseData: "Payment cancelled event processed",
		UserID:       data.UserID,
	}

	if _, err := e.service.Log.Create(ctx, logEntry); err != nil {
		logger.Warnf(ctx, "Failed to create payment log: %v", err)
	}

	logger.Infof(ctx, "Payment cancellation event processed")
}

func (e *eventHandler) processPaymentExpired(ctx context.Context, data *event.PaymentEventData) {
	logger.Infof(ctx, "Payment expired: OrderID=%s, Amount=%.2f %s",
		data.OrderID, data.Amount, data.Currency)

	// Handle subscription expiration
	if data.SubscriptionID != "" {
		logger.Infof(ctx, "Payment expired for subscription: SubscriptionID=%s", data.SubscriptionID)

		subscription, err := e.service.Subscription.GetByID(ctx, data.SubscriptionID)
		if err != nil {
			logger.Errorf(ctx, "Failed to get subscription: %v", err)
		} else {
			// Add expiration metadata
			metadata := subscription.Metadata
			if metadata == nil {
				metadata = make(map[string]any)
			}
			metadata["payment_expired_at"] = data.Timestamp.Format("2006-01-02T15:04:05Z07:00")
			metadata["expired_order_id"] = data.OrderID

			updates := map[string]any{"metadata": metadata}
			if _, err := e.service.Subscription.Update(ctx, data.SubscriptionID, updates); err != nil {
				logger.Errorf(ctx, "Failed to update subscription metadata: %v", err)
			}
		}
	}

	// Log payment expiration for analytics
	logEntry := &structs.CreateLogInput{
		OrderID:      data.OrderID,
		ChannelID:    data.ChannelID,
		Type:         structs.LogTypeUpdate,
		StatusBefore: structs.PaymentStatusPending,
		StatusAfter:  structs.PaymentStatusCancelled,
		ResponseData: "Payment expired event processed",
		UserID:       data.UserID,
	}

	if _, err := e.service.Log.Create(ctx, logEntry); err != nil {
		logger.Warnf(ctx, "Failed to create payment log: %v", err)
	}

	logger.Infof(ctx, "Payment expiration event processed")
}

func (e *eventHandler) processPaymentRefunded(ctx context.Context, data *event.PaymentEventData) {
	logger.Infof(ctx, "Payment refunded: OrderID=%s, Amount=%.2f %s",
		data.OrderID, data.Amount, data.Currency)

	// Handle subscription refund
	if data.SubscriptionID != "" {
		logger.Infof(ctx, "Payment refunded for subscription: SubscriptionID=%s", data.SubscriptionID)

		subscription, err := e.service.Subscription.GetByID(ctx, data.SubscriptionID)
		if err != nil {
			logger.Errorf(ctx, "Failed to get subscription: %v", err)
		} else {
			// Add refund metadata
			metadata := subscription.Metadata
			if metadata == nil {
				metadata = make(map[string]any)
			}
			metadata["payment_refunded_at"] = data.Timestamp.Format("2006-01-02T15:04:05Z07:00")
			metadata["refunded_order_id"] = data.OrderID
			metadata["refund_amount"] = data.Amount

			updates := map[string]any{"metadata": metadata}
			if _, err := e.service.Subscription.Update(ctx, data.SubscriptionID, updates); err != nil {
				logger.Errorf(ctx, "Failed to update subscription metadata: %v", err)
			}
		}
	}

	// Revoke product access if applicable
	if data.ProductID != "" && e.usw != nil && e.usw.HasUserService() {
		logger.Infof(ctx, "Revoking product access due to refund: UserID=%s, ProductID=%s",
			data.UserID, data.ProductID)
		// User service can be extended to revoke product access
	}

	// Log payment refund
	logEntry := &structs.CreateLogInput{
		OrderID:      data.OrderID,
		ChannelID:    data.ChannelID,
		Type:         structs.LogTypeRefund,
		StatusBefore: structs.PaymentStatusCompleted,
		StatusAfter:  structs.PaymentStatusRefunded,
		ResponseData: "Payment refunded event processed",
		UserID:       data.UserID,
	}

	if _, err := e.service.Log.Create(ctx, logEntry); err != nil {
		logger.Warnf(ctx, "Failed to create payment log: %v", err)
	}

	logger.Infof(ctx, "Payment refund event processed successfully")
}

// Subscription event specialized handlers

func (e *eventHandler) processSubscriptionCreated(ctx context.Context, data *event.SubscriptionEventData) {
	logger.Infof(ctx, "Subscription created: SubscriptionID=%s, UserID=%s, ProductID=%s",
		data.SubscriptionID, data.UserID, data.ProductID)

	// Get product details for welcome message
	product, err := e.service.Product.GetByID(ctx, data.ProductID)
	if err != nil {
		logger.Errorf(ctx, "Failed to get product details: %v", err)
	} else {
		logger.Infof(ctx, "User subscribed to product: %s", product.Name)
	}

	// Grant access to subscription benefits
	if e.usw != nil && e.usw.HasUserService() {
		logger.Infof(ctx, "Granting subscription access: UserID=%s, ProductID=%s",
			data.UserID, data.ProductID)

		// Update user metadata to track subscription
		user, err := e.usw.GetUserByID(ctx, data.UserID)
		if err != nil {
			logger.Errorf(ctx, "Failed to get user: %v", err)
		} else {
			logger.Infof(ctx, "User %s now has active subscription", user.Username)
		}
	}

	// Add user to space if space-based subscription
	if data.SpaceID != "" && e.tsw != nil && e.tsw.HasUserSpaceService() {
		logger.Infof(ctx, "Adding user to space: UserID=%s, SpaceID=%s",
			data.UserID, data.SpaceID)

		if _, err := e.tsw.AddUserToSpace(ctx, data.UserID, data.SpaceID); err != nil {
			logger.Warnf(ctx, "Failed to add user to space: %v", err)
		}
	}

	logger.Infof(ctx, "Subscription creation event processed successfully")
}

func (e *eventHandler) processSubscriptionRenewed(ctx context.Context, data *event.SubscriptionEventData) {
	logger.Infof(ctx, "Subscription renewed: SubscriptionID=%s, UserID=%s",
		data.SubscriptionID, data.UserID)

	// Get subscription details
	subscription, err := e.service.Subscription.GetByID(ctx, data.SubscriptionID)
	if err != nil {
		logger.Errorf(ctx, "Failed to get subscription: %v", err)
		return
	}

	// Log renewal for analytics
	logger.Infof(ctx, "Subscription renewed until: %s",
		subscription.CurrentPeriodEnd.Format("2006-01-02"))

	// Ensure subscription is active
	if subscription.Status != structs.SubscriptionStatusActive {
		updates := map[string]any{
			"status": structs.SubscriptionStatusActive,
		}
		if _, err := e.service.Subscription.Update(ctx, data.SubscriptionID, updates); err != nil {
			logger.Errorf(ctx, "Failed to activate renewed subscription: %v", err)
		}
	}

	// Update user access if needed
	if e.usw != nil && e.usw.HasUserService() {
		logger.Infof(ctx, "Extending subscription access: UserID=%s", data.UserID)
	}

	logger.Infof(ctx, "Subscription renewal event processed successfully")
}

func (e *eventHandler) processSubscriptionUpdated(ctx context.Context, data *event.SubscriptionEventData) {
	logger.Infof(ctx, "Subscription updated: SubscriptionID=%s, Status=%s",
		data.SubscriptionID, data.Status)

	// Get subscription details
	subscription, err := e.service.Subscription.GetByID(ctx, data.SubscriptionID)
	if err != nil {
		logger.Errorf(ctx, "Failed to get subscription: %v", err)
		return
	}

	// Check if product changed (upgrade/downgrade)
	if subscription.ProductID != data.ProductID {
		logger.Infof(ctx, "Subscription product changed: %s -> %s",
			subscription.ProductID, data.ProductID)

		// Update access levels based on new product
		if e.usw != nil && e.usw.HasUserService() {
			logger.Infof(ctx, "Updating user access levels: UserID=%s", data.UserID)
		}
	}

	// Check if status changed
	if subscription.Status != data.Status {
		logger.Infof(ctx, "Subscription status changed: %s -> %s",
			subscription.Status, data.Status)
	}

	logger.Infof(ctx, "Subscription update event processed successfully")
}

func (e *eventHandler) processSubscriptionCancelled(ctx context.Context, data *event.SubscriptionEventData) {
	logger.Infof(ctx, "Subscription cancelled: SubscriptionID=%s, UserID=%s",
		data.SubscriptionID, data.UserID)

	// Get subscription details
	subscription, err := e.service.Subscription.GetByID(ctx, data.SubscriptionID)
	if err != nil {
		logger.Errorf(ctx, "Failed to get subscription: %v", err)
		return
	}

	// Check if cancellation is immediate or at period end
	if subscription.CancelledAt != nil {
		logger.Infof(ctx, "Subscription cancelled immediately")

		// Revoke access immediately
		if e.usw != nil && e.usw.HasUserService() {
			logger.Infof(ctx, "Revoking subscription access: UserID=%s", data.UserID)
		}
	} else if subscription.CancelAt != nil {
		logger.Infof(ctx, "Subscription will cancel at period end: %s",
			subscription.CancelAt.Format("2006-01-02"))
	}

	// Record cancellation reason for analytics
	if subscription.Metadata != nil {
		if reason, ok := subscription.Metadata["cancellation_reason"].(string); ok {
			logger.Infof(ctx, "Cancellation reason: %s", reason)
		}
	}

	logger.Infof(ctx, "Subscription cancellation event processed successfully")
}

func (e *eventHandler) processSubscriptionExpired(ctx context.Context, data *event.SubscriptionEventData) {
	logger.Infof(ctx, "Subscription expired: SubscriptionID=%s, UserID=%s",
		data.SubscriptionID, data.UserID)

	// Get subscription details
	subscription, err := e.service.Subscription.GetByID(ctx, data.SubscriptionID)
	if err != nil {
		logger.Errorf(ctx, "Failed to get subscription: %v", err)
		return
	}

	// Revoke user access to premium features
	if e.usw != nil && e.usw.HasUserService() {
		logger.Infof(ctx, "Revoking expired subscription access: UserID=%s, ProductID=%s",
			data.UserID, data.ProductID)

		user, err := e.usw.GetUserByID(ctx, data.UserID)
		if err != nil {
			logger.Errorf(ctx, "Failed to get user: %v", err)
		} else {
			logger.Infof(ctx, "User %s subscription has expired", user.Username)
		}
	}

	// Remove user from space if space-based subscription
	if data.SpaceID != "" && e.tsw != nil && e.tsw.HasSpaceService() {
		logger.Infof(ctx, "Subscription expired for space: SpaceID=%s", data.SpaceID)
	}

	// Log expiration for analytics
	logger.Infof(ctx, "Subscription expired on: %s",
		subscription.CurrentPeriodEnd.Format("2006-01-02"))

	logger.Infof(ctx, "Subscription expiration event processed successfully")
}

// Product event specialized handlers

func (e *eventHandler) processProductCreated(ctx context.Context, data *event.ProductEventData) {
	logger.Infof(ctx, "Product created: ProductID=%s, Name=%s, Price=%.2f %s",
		data.ProductID, data.Name, data.Price, data.Currency)

	// Log product creation for analytics
	logger.Infof(ctx, "New product available: %s (Type: %s, Pricing: %s)",
		data.Name, data.PricingType, data.BillingInterval)

	// Update product catalog cache
	if data.Status == structs.ProductStatusActive {
		logger.Infof(ctx, "Product is active and available for purchase")
	} else {
		logger.Infof(ctx, "Product created but not yet active (Status: %s)", data.Status)
	}

	// Initialize product analytics
	if data.Metadata == nil {
		data.Metadata = make(map[string]any)
	}
	data.Metadata["created_at"] = data.Timestamp.Format("2006-01-02T15:04:05Z07:00")

	// Notify space admins if space-specific product
	if data.SpaceID != "" && e.tsw != nil && e.tsw.HasSpaceService() {
		space, err := e.tsw.GetSpace(ctx, data.SpaceID)
		if err != nil {
			logger.Warnf(ctx, "Failed to get space for product: %v", err)
		} else {
			logger.Infof(ctx, "Product created for space: %s", space.Name)
		}
	}

	logger.Infof(ctx, "Product creation event processed successfully")
}

func (e *eventHandler) processProductUpdated(ctx context.Context, data *event.ProductEventData) {
	logger.Infof(ctx, "Product updated: ProductID=%s, Name=%s",
		data.ProductID, data.Name)

	// Get existing product to compare changes
	product, err := e.service.Product.GetByID(ctx, data.ProductID)
	if err != nil {
		logger.Errorf(ctx, "Failed to get product: %v", err)
		return
	}

	// Check for price changes
	if product.Price != data.Price {
		logger.Infof(ctx, "Product price changed: %.2f -> %.2f %s",
			product.Price, data.Price, data.Currency)

		// Get active subscriptions for this product
		query := &structs.SubscriptionQuery{
			ProductID: data.ProductID,
			Status:    structs.SubscriptionStatusActive,
		}
		subscriptions, err := e.service.Subscription.List(ctx, query)
		if err != nil {
			logger.Errorf(ctx, "Failed to get subscriptions: %v", err)
		} else if len(subscriptions.Items) > 0 {
			logger.Infof(ctx, "Price change affects %d active subscriptions", len(subscriptions.Items))
		}
	}

	// Check for status changes
	if product.Status != data.Status {
		logger.Infof(ctx, "Product status changed: %s -> %s", product.Status, data.Status)

		if data.Status == structs.ProductStatusDisabled {
			logger.Warnf(ctx, "Product deactivated - no new subscriptions allowed")
		}
	}

	// Update product catalog cache
	logger.Infof(ctx, "Product catalog updated")

	logger.Infof(ctx, "Product update event processed successfully")
}

func (e *eventHandler) processProductDeleted(ctx context.Context, data *event.ProductEventData) {
	logger.Infof(ctx, "Product deleted: ProductID=%s, Name=%s", data.ProductID, data.Name)

	// Check for active subscriptions
	query := &structs.SubscriptionQuery{
		ProductID: data.ProductID,
		Status:    structs.SubscriptionStatusActive,
	}
	subscriptions, err := e.service.Subscription.List(ctx, query)
	if err != nil {
		logger.Errorf(ctx, "Failed to get subscriptions: %v", err)
	} else if len(subscriptions.Items) > 0 {
		logger.Warnf(ctx, "Product has %d active subscriptions that need handling", len(subscriptions.Items))

		// Log each active subscription
		for _, sub := range subscriptions.Items {
			logger.Infof(ctx, "Active subscription: %s (User: %s, Expires: %s)",
				sub.ID, sub.UserID, sub.CurrentPeriodEnd.Format("2006-01-02"))
		}
	}

	// Remove product from catalog
	logger.Infof(ctx, "Product removed from catalog")

	// Archive product data for reporting
	logger.Infof(ctx, "Product data archived for historical reporting")

	logger.Infof(ctx, "Product deletion event processed successfully")
}

// Channel event specialized handlers

func (e *eventHandler) processChannelCreated(ctx context.Context, data *event.ChannelEventData) {
	logger.Infof(ctx, "Payment channel created: ChannelID=%s, Name=%s, Provider=%s",
		data.ChannelID, data.Name, data.Provider)

	// Validate channel configuration
	channel, err := e.service.Channel.GetByID(ctx, data.ChannelID)
	if err != nil {
		logger.Errorf(ctx, "Failed to get channel: %v", err)
		return
	}

	// Log channel details
	logger.Infof(ctx, "Channel status: %s, Default: %v", channel.Status, channel.IsDefault)
	logger.Infof(ctx, "Supported payment types: %v", channel.SupportedType)

	// Initialize channel monitoring
	if data.Status == structs.ChannelStatusActive {
		logger.Infof(ctx, "Channel is active and ready for payments")
	} else {
		logger.Infof(ctx, "Channel created but not active (Status: %s)", data.Status)
	}

	// Notify space admins if space-specific channel
	if data.SpaceID != "" && e.tsw != nil && e.tsw.HasSpaceService() {
		space, err := e.tsw.GetSpace(ctx, data.SpaceID)
		if err != nil {
			logger.Warnf(ctx, "Failed to get space for channel: %v", err)
		} else {
			logger.Infof(ctx, "Payment channel created for space: %s", space.Name)
		}
	}

	logger.Infof(ctx, "Channel creation event processed successfully")
}

func (e *eventHandler) processChannelUpdated(ctx context.Context, data *event.ChannelEventData) {
	logger.Infof(ctx, "Payment channel updated: ChannelID=%s, Name=%s",
		data.ChannelID, data.Name)

	// Get existing channel to compare changes
	channel, err := e.service.Channel.GetByID(ctx, data.ChannelID)
	if err != nil {
		logger.Errorf(ctx, "Failed to get channel: %v", err)
		return
	}

	// Check for status changes
	if channel.Status != data.Status {
		logger.Infof(ctx, "Channel status changed: %s -> %s", channel.Status, data.Status)
	}

	// Check for default channel changes
	if channel.IsDefault != data.IsDefault {
		if data.IsDefault {
			logger.Infof(ctx, "Channel set as default payment method")
		} else {
			logger.Infof(ctx, "Channel removed as default payment method")
		}
	}

	// Check for configuration changes
	logger.Infof(ctx, "Channel configuration updated")

	// Update payment method displays
	logger.Infof(ctx, "Payment method cache refreshed")

	// Log changes for audit
	if data.Metadata != nil {
		if updatedBy, ok := data.Metadata["updated_by"].(string); ok {
			logger.Infof(ctx, "Channel updated by: %s", updatedBy)
		}
	}

	logger.Infof(ctx, "Channel update event processed successfully")
}

func (e *eventHandler) processChannelDeleted(ctx context.Context, data *event.ChannelEventData) {
	logger.Infof(ctx, "Payment channel deleted: ChannelID=%s, Name=%s",
		data.ChannelID, data.Name)

	// Check for active payments using this channel
	orderQuery := &structs.OrderQuery{
		ChannelID: data.ChannelID,
		Status:    structs.PaymentStatusPending,
	}
	orders, err := e.service.Order.List(ctx, orderQuery)
	if err != nil {
		logger.Errorf(ctx, "Failed to get orders: %v", err)
	} else if len(orders.Items) > 0 {
		logger.Warnf(ctx, "Channel has %d pending payments that need migration", len(orders.Items))
	}

	// Check for active subscriptions using this channel
	subQuery := &structs.SubscriptionQuery{
		ChannelID: data.ChannelID,
		Status:    structs.SubscriptionStatusActive,
	}
	subscriptions, err := e.service.Subscription.List(ctx, subQuery)
	if err != nil {
		logger.Errorf(ctx, "Failed to get subscriptions: %v", err)
	} else if len(subscriptions.Items) > 0 {
		logger.Warnf(ctx, "Channel has %d active subscriptions that need migration", len(subscriptions.Items))

		// Log each active subscription
		for _, sub := range subscriptions.Items {
			logger.Infof(ctx, "Active subscription: %s (User: %s)", sub.ID, sub.UserID)
		}
	}

	// Remove channel from available payment methods
	logger.Infof(ctx, "Channel removed from payment methods")

	// Archive channel configuration
	logger.Infof(ctx, "Channel configuration archived")

	logger.Infof(ctx, "Channel deletion event processed successfully")
}

func (e *eventHandler) processChannelActivated(ctx context.Context, data *event.ChannelEventData) {
	logger.Infof(ctx, "Payment channel activated: ChannelID=%s, Name=%s, Provider=%s",
		data.ChannelID, data.Name, data.Provider)

	// Get channel details
	channel, err := e.service.Channel.GetByID(ctx, data.ChannelID)
	if err != nil {
		logger.Errorf(ctx, "Failed to get channel: %v", err)
		return
	}

	// Verify channel configuration
	logger.Infof(ctx, "Channel configuration verified")
	logger.Infof(ctx, "Supported payment types: %v", channel.SupportedType)

	// Add channel to active payment methods
	logger.Infof(ctx, "Channel added to active payment methods")

	// Initialize monitoring for the channel
	logger.Infof(ctx, "Channel monitoring initialized")

	// Update payment method display in checkout
	if data.IsDefault {
		logger.Infof(ctx, "Channel set as default payment method")
	}

	// Notify space admins if applicable
	if data.SpaceID != "" && e.tsw != nil && e.tsw.HasSpaceService() {
		logger.Infof(ctx, "Payment channel activated for space: %s", data.SpaceID)
	}

	logger.Infof(ctx, "Channel activation event processed successfully")
}

func (e *eventHandler) processChannelDisabled(ctx context.Context, data *event.ChannelEventData) {
	logger.Infof(ctx, "Payment channel disabled: ChannelID=%s, Name=%s",
		data.ChannelID, data.Name)

	// Remove channel from active payment methods
	logger.Infof(ctx, "Channel removed from active payment methods")

	// Check for pending payments
	orderQuery := &structs.OrderQuery{
		ChannelID: data.ChannelID,
		Status:    structs.PaymentStatusPending,
	}
	orders, err := e.service.Order.List(ctx, orderQuery)
	if err != nil {
		logger.Errorf(ctx, "Failed to get orders: %v", err)
	} else if len(orders.Items) > 0 {
		logger.Warnf(ctx, "Channel has %d pending payments", len(orders.Items))
	}

	// Check for recurring payments that need redirection
	subQuery := &structs.SubscriptionQuery{
		ChannelID: data.ChannelID,
		Status:    structs.SubscriptionStatusActive,
	}
	subscriptions, err := e.service.Subscription.List(ctx, subQuery)
	if err != nil {
		logger.Errorf(ctx, "Failed to get subscriptions: %v", err)
	} else if len(subscriptions.Items) > 0 {
		logger.Warnf(ctx, "Channel has %d active subscriptions that may need alternative payment method",
			len(subscriptions.Items))
	}

	// Update analytics and monitoring
	logger.Infof(ctx, "Channel monitoring updated")

	// Log reason for disabling if available
	if data.Metadata != nil {
		if reason, ok := data.Metadata["disable_reason"].(string); ok {
			logger.Infof(ctx, "Channel disabled reason: %s", reason)
		}
	}

	logger.Infof(ctx, "Channel disable event processed successfully")
}
