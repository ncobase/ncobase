package service

import (
	"context"
	"fmt"
	"ncobase/payment/data/repository"
	"ncobase/payment/event"
	"ncobase/payment/structs"
	"time"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils/nanoid"
)

// OrderServiceInterface defines the interface for order service operations
type OrderServiceInterface interface {
	Create(ctx context.Context, input *structs.CreateOrderInput) (*structs.Order, error)
	GetByID(ctx context.Context, id string) (*structs.Order, error)
	GetByOrderNumber(ctx context.Context, orderNumber string) (*structs.Order, error)
	GeneratePaymentURL(ctx context.Context, orderID string) (string, map[string]any, error)
	VerifyPayment(ctx context.Context, orderID string, data map[string]any) error
	RefundPayment(ctx context.Context, orderID string, amount float64, reason string) error
	List(ctx context.Context, query *structs.OrderQuery) (paging.Result[*structs.Order], error)
	ProcessWebhook(ctx context.Context, channelID string, payload []byte, headers map[string]string) error
	Serialize(order *structs.Order) *structs.Order
	Serializes(orders []*structs.Order) []*structs.Order
}

// orderService provides operations for payment orders
type orderService struct {
	repo        repository.OrderRepositoryInterface
	channelRepo repository.ChannelRepositoryInterface
	logRepo     repository.LogRepositoryInterface
	publisher   event.PublisherInterface
	providerSvc ProviderServiceInterface
}

// NewOrderService creates a new order service
func NewOrderService(
	repo repository.OrderRepositoryInterface,
	channelRepo repository.ChannelRepositoryInterface,
	logRepo repository.LogRepositoryInterface,
	publisher event.PublisherInterface,
	providerSvc ProviderServiceInterface,
) OrderServiceInterface {
	return &orderService{
		repo:        repo,
		channelRepo: channelRepo,
		logRepo:     logRepo,
		publisher:   publisher,
		providerSvc: providerSvc,
	}
}

// Create creates a new payment order
func (s *orderService) Create(ctx context.Context, input *structs.CreateOrderInput) (*structs.Order, error) {
	// Validate input
	if input.Amount <= 0 {
		return nil, fmt.Errorf(ecode.FieldIsInvalid("amount must be greater than zero"))
	}

	if input.ChannelID == "" {
		return nil, fmt.Errorf(ecode.FieldIsRequired("channel_id"))
	}

	if input.UserID == "" {
		return nil, fmt.Errorf(ecode.FieldIsRequired("user_id"))
	}

	// Validate currency (new validation)
	supportedCurrencies := map[structs.CurrencyCode]bool{
		structs.CurrencyUSD: true,
		structs.CurrencyEUR: true,
		structs.CurrencyGBP: true,
		structs.CurrencyCNY: true,
		structs.CurrencyJPY: true,
	}

	if input.Currency == "" {
		input.Currency = structs.CurrencyUSD // Set default currency
	} else if !supportedCurrencies[input.Currency] {
		return nil, fmt.Errorf(ecode.FieldIsInvalid("unsupported currency"))
	}

	// Get the payment channel
	channel, err := s.channelRepo.GetByID(ctx, input.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment channel: %w", err)
	}

	// Check if the channel is active
	if channel.Status != structs.ChannelStatusActive {
		return nil, fmt.Errorf("payment channel is not active")
	}

	// Check if the channel supports the payment type
	supportsType := false
	for _, t := range channel.SupportedType {
		if t == input.Type {
			supportsType = true
			break
		}
	}

	if !supportsType {
		return nil, fmt.Errorf("payment channel does not support payment type %s", input.Type)
	}

	// Generate unique order number
	orderNumber := s.generateOrderNumber()

	// Set default expiration time if not provided
	expiresAt := input.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = time.Now().Add(24 * time.Hour) // Default: 24 hours
	}

	// Create order entity
	order := &structs.CreateOrderInput{
		OrderNumber:    orderNumber,
		Amount:         input.Amount,
		Currency:       input.Currency,
		Status:         structs.PaymentStatusPending,
		Type:           input.Type,
		ChannelID:      input.ChannelID,
		UserID:         input.UserID,
		TenantID:       input.TenantID,
		ProductID:      input.ProductID,
		SubscriptionID: input.SubscriptionID,
		ExpiresAt:      expiresAt,
		Description:    input.Description,
		Metadata:       input.Metadata,
	}

	// Save to database
	created, err := s.repo.Create(ctx, order)
	if err != nil {
		return nil, err
	}

	// Create payment log
	logEntry := &structs.CreateLogInput{
		OrderID:     created.ID,
		ChannelID:   created.ChannelID,
		Type:        structs.LogTypeCreate,
		StatusAfter: created.Status,
		RequestData: fmt.Sprintf("%+v", input),
		UserID:      created.UserID,
	}

	_, err = s.logRepo.Create(ctx, logEntry)
	if err != nil {
		logger.Warnf(ctx, "Failed to create payment log: %v", err)
	}

	// Publish event
	if s.publisher != nil {
		// Create event data
		eventData := event.NewPaymentEventData(
			created.ID,
			created.OrderNumber,
			created.ChannelID,
			channel.Provider,
			created.UserID,
			created.TenantID,
			created.Amount,
			created.Currency,
			created.Status,
			created.Type,
			created.ProductID,
			created.SubscriptionID,
			created.Metadata,
		)

		// Publish event
		s.publisher.PublishPaymentCreated(ctx, eventData)
	}

	return s.Serialize(created), nil
}

// GeneratePaymentURL generates a payment URL for a payment order
func (s *orderService) GeneratePaymentURL(ctx context.Context, orderID string) (string, map[string]any, error) {
	if orderID == "" {
		return "", nil, fmt.Errorf(ecode.FieldIsRequired("order_id"))
	}

	// Get the order
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Check if the order is pending
	if order.Status != structs.PaymentStatusPending {
		return "", nil, fmt.Errorf("order is not in pending status")
	}

	// Check if the order has expired
	if order.ExpiresAt.Before(time.Now()) {
		// Update order status to expired
		_, err = s.repo.Update(ctx, &structs.UpdateOrderInput{
			Status: structs.PaymentStatusCancelled,
		})
		if err != nil {
			logger.Warnf(ctx, "Failed to update expired order status: %v", err)
		}

		return "", nil, fmt.Errorf("order has expired")
	}

	// Get the payment channel
	channel, err := s.channelRepo.GetByID(ctx, order.ChannelID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get payment channel: %w", err)
	}

	// Get payment provider
	paymentProvider, err := s.providerSvc.GetProvider(channel.Provider, channel.Config)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get payment provider: %w", err)
	}

	// Create payment with provider
	providerRef, paymentData, err := paymentProvider.CreatePayment(ctx, order)
	if err != nil {
		// Log the error
		errorLog := &structs.CreateLogInput{
			OrderID:      order.ID,
			ChannelID:    order.ChannelID,
			Type:         structs.LogTypeError,
			StatusBefore: order.Status,
			StatusAfter:  order.Status,
			Error:        err.Error(),
			UserID:       order.UserID,
		}

		_, logErr := s.logRepo.Create(ctx, errorLog)
		if logErr != nil {
			logger.Warnf(ctx, "Failed to create error log: %v", logErr)
		}

		return "", nil, fmt.Errorf("failed to create payment with provider: %w", err)
	}

	_, err = s.repo.Update(ctx, &structs.UpdateOrderInput{
		ProviderRef: providerRef,
	})
	if err != nil {
		logger.Warnf(ctx, "Failed to update order with provider reference: %v", err)
	}

	// Log the payment creation
	logEntry := &structs.CreateLogInput{
		OrderID:      order.ID,
		ChannelID:    order.ChannelID,
		Type:         structs.LogTypeUpdate,
		StatusBefore: order.Status,
		StatusAfter:  order.Status,
		RequestData:  fmt.Sprintf("Generate payment URL for order %s", order.ID),
		ResponseData: fmt.Sprintf("Provider reference: %s", providerRef),
		UserID:       order.UserID,
	}

	_, err = s.logRepo.Create(ctx, logEntry)
	if err != nil {
		logger.Warnf(ctx, "Failed to create payment log: %v", err)
	}

	// Return the payment URL
	return providerRef, paymentData, nil
}

// VerifyPayment verifies a payment with the provider
func (s *orderService) VerifyPayment(ctx context.Context, orderID string, verificationData map[string]any) error {
	if orderID == "" {
		return fmt.Errorf("order ID is required")
	}

	// Get the order
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Get the payment channel
	channel, err := s.channelRepo.GetByID(ctx, order.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to get payment channel: %w", err)
	}

	// Get payment provider
	paymentProvider, err := s.providerSvc.GetProvider(channel.Provider, channel.Config)
	if err != nil {
		return fmt.Errorf("failed to get payment provider: %w", err)
	}

	// Verify payment with provider
	result, err := paymentProvider.VerifyPayment(ctx, order.ID, verificationData)
	if err != nil {
		// Log the error
		errorLog := &structs.CreateLogInput{
			OrderID:      order.ID,
			ChannelID:    order.ChannelID,
			Type:         structs.LogTypeVerify,
			StatusBefore: order.Status,
			StatusAfter:  order.Status,
			RequestData:  fmt.Sprintf("%+v", verificationData),
			Error:        err.Error(),
			UserID:       order.UserID,
		}

		_, logErr := s.logRepo.Create(ctx, errorLog)
		if logErr != nil {
			logger.Warnf(ctx, "Failed to create verification error log: %v", logErr)
		}

		return fmt.Errorf("failed to verify payment with provider: %w", err)
	}

	// Get the old status for logging
	oldStatus := order.Status

	// Update order status based on verification result
	if result.Success {
		order.Status = result.Status

		// If payment is completed, update paid time
		if order.Status == structs.PaymentStatusCompleted {
			now := time.Now()
			order.PaidAt = now
		}

		// Update provider reference if provided
		if result.ProviderRef != "" {
			order.ProviderRef = result.ProviderRef
		}

		// Update metadata
		if result.Metadata != nil {
			if order.Metadata == nil {
				order.Metadata = make(map[string]any)
			}

			// Merge metadata
			for k, v := range result.Metadata {
				order.Metadata[k] = v
			}
		}
		// Save updated order
		_, err = s.repo.Update(ctx, &structs.UpdateOrderInput{
			Status:      order.Status,
			PaidAt:      &order.PaidAt,
			ProviderRef: order.ProviderRef,
			Metadata:    order.Metadata,
		})
		if err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}
	}

	// Log the verification
	logEntry := &structs.CreateLogInput{
		OrderID:      order.ID,
		ChannelID:    order.ChannelID,
		Type:         structs.LogTypeVerify,
		StatusBefore: oldStatus,
		StatusAfter:  order.Status,
		RequestData:  fmt.Sprintf("%+v", verificationData),
		ResponseData: fmt.Sprintf("%+v", result),
		UserID:       order.UserID,
	}

	_, err = s.logRepo.Create(ctx, logEntry)
	if err != nil {
		logger.Warnf(ctx, "Failed to create verification log: %v", err)
	}

	// Publish event if status changed
	if oldStatus != order.Status && s.publisher != nil {
		// Create event data
		eventData := event.NewPaymentEventData(
			order.ID,
			order.OrderNumber,
			order.ChannelID,
			channel.Provider,
			order.UserID,
			order.TenantID,
			order.Amount,
			order.Currency,
			order.Status,
			order.Type,
			order.ProductID,
			order.SubscriptionID,
			order.Metadata,
		)

		// Publish appropriate event based on new status
		switch order.Status {
		case structs.PaymentStatusCompleted:
			s.publisher.PublishPaymentSucceeded(ctx, eventData)
		case structs.PaymentStatusFailed:
			s.publisher.PublishPaymentFailed(ctx, eventData)
		case structs.PaymentStatusCancelled:
			s.publisher.PublishPaymentCancelled(ctx, eventData)
		case structs.PaymentStatusRefunded:
			s.publisher.PublishPaymentRefunded(ctx, eventData)
		}
	}

	return nil
}

// ProcessWebhook processes a webhook from the payment provider
func (s *orderService) ProcessWebhook(ctx context.Context, channelID string, payload []byte, headers map[string]string) error {
	// Get the payment channel
	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to get payment channel: %w", err)
	}

	// Get payment provider
	paymentProvider, err := s.providerSvc.GetProvider(channel.Provider, channel.Config)
	if err != nil {
		return fmt.Errorf("failed to get payment provider: %w", err)
	}

	// Process webhook with provider
	result, err := paymentProvider.ProcessWebhook(ctx, payload, headers)
	if err != nil {
		// Log the error
		errorLog := &structs.CreateLogInput{
			ChannelID:   channelID,
			Type:        structs.LogTypeCallback,
			RequestData: string(payload),
			Error:       err.Error(),
		}

		_, logErr := s.logRepo.Create(ctx, errorLog)
		if logErr != nil {
			logger.Warnf(ctx, "Failed to create webhook error log: %v", logErr)
		}

		return fmt.Errorf("failed to process webhook with provider: %w", err)
	}

	// If no order ID is found, just log the webhook
	if result.OrderID == "" {
		// Log the webhook
		logEntry := &structs.CreateLogInput{
			ChannelID:    channelID,
			Type:         structs.LogTypeCallback,
			RequestData:  string(payload),
			ResponseData: fmt.Sprintf("%+v", result),
		}

		_, err = s.logRepo.Create(ctx, logEntry)
		if err != nil {
			logger.Warnf(ctx, "Failed to create webhook log: %v", err)
		}

		return nil
	}

	// Get the order
	order, err := s.repo.GetByOrderNumber(ctx, result.OrderID)
	if err != nil {
		// Try to get by ID if not found by order number
		order, err = s.repo.GetByID(ctx, result.OrderID)
		if err != nil {
			// Log that we received a webhook for an unknown order
			logEntry := &structs.CreateLogInput{
				ChannelID:    channelID,
				Type:         structs.LogTypeCallback,
				RequestData:  string(payload),
				ResponseData: fmt.Sprintf("%+v", result),
				Error:        fmt.Sprintf("Order not found: %s", result.OrderID),
			}

			_, logErr := s.logRepo.Create(ctx, logEntry)
			if logErr != nil {
				logger.Warnf(ctx, "Failed to create webhook log: %v", logErr)
			}

			return fmt.Errorf("order not found: %s", result.OrderID)
		}
	}

	// Get the old status for logging
	oldStatus := order.Status

	// Update order status based on webhook result
	if result.Status != "" {
		order.Status = result.Status

		// If payment is completed, update paid time
		if order.Status == structs.PaymentStatusCompleted && order.PaidAt.IsZero() {
			now := time.Now()
			order.PaidAt = now
		}

		// Update provider reference if provided
		if result.ProviderRef != "" {
			order.ProviderRef = result.ProviderRef
		}

		// Save updated order
		_, err = s.repo.Update(ctx, &structs.UpdateOrderInput{
			Status:      order.Status,
			PaidAt:      &order.PaidAt,
			ProviderRef: order.ProviderRef,
		})
		if err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}
	}

	// Log the webhook
	logEntry := &structs.CreateLogInput{
		OrderID:      order.ID,
		ChannelID:    channelID,
		Type:         structs.LogTypeCallback,
		StatusBefore: oldStatus,
		StatusAfter:  order.Status,
		RequestData:  string(payload),
		ResponseData: fmt.Sprintf("%+v", result),
	}

	_, err = s.logRepo.Create(ctx, logEntry)
	if err != nil {
		logger.Warnf(ctx, "Failed to create webhook log: %v", err)
	}

	// Publish event if status changed
	if oldStatus != order.Status && s.publisher != nil {
		eventData := &event.PaymentEventData{
			OrderID:        order.ID,
			OrderNumber:    order.OrderNumber,
			ChannelID:      order.ChannelID,
			Provider:       channel.Provider,
			UserID:         order.UserID,
			TenantID:       order.TenantID,
			Amount:         order.Amount,
			Currency:       order.Currency,
			Status:         order.Status,
			Type:           order.Type,
			ProductID:      order.ProductID,
			SubscriptionID: order.SubscriptionID,
			Metadata:       order.Metadata,
		}

		switch order.Status {
		case structs.PaymentStatusCompleted:
			s.publisher.PublishPaymentSucceeded(ctx, eventData)
		case structs.PaymentStatusFailed:
			s.publisher.PublishPaymentFailed(ctx, eventData)
		case structs.PaymentStatusCancelled:
			s.publisher.PublishPaymentCancelled(ctx, eventData)
		case structs.PaymentStatusRefunded:
			s.publisher.PublishPaymentRefunded(ctx, eventData)
		}
	}

	return nil
}

// RefundPayment requests a refund for a payment
func (s *orderService) RefundPayment(ctx context.Context, orderID string, amount float64, reason string) error {
	if orderID == "" {
		return fmt.Errorf("order ID is required")
	}

	// Get the order
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Check if the order is completed
	if order.Status != structs.PaymentStatusCompleted {
		return fmt.Errorf("only completed payments can be refunded")
	}

	// If amount is not specified, use the full amount
	if amount <= 0 {
		amount = order.Amount
	}

	// Check if the amount is valid
	if amount > order.Amount {
		return fmt.Errorf("refund amount cannot exceed payment amount")
	}

	// Get the payment channel
	channel, err := s.channelRepo.GetByID(ctx, order.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to get payment channel: %w", err)
	}

	// Get payment provider
	paymentProvider, err := s.providerSvc.GetProvider(channel.Provider, channel.Config)
	if err != nil {
		return fmt.Errorf("failed to get payment provider: %w", err)
	}

	// Request refund with provider
	result, err := paymentProvider.RefundPayment(ctx, order.ID, amount, reason)
	if err != nil {
		// Log the error
		errorLog := &structs.CreateLogInput{
			OrderID:      order.ID,
			ChannelID:    order.ChannelID,
			Type:         structs.LogTypeRefund,
			StatusBefore: order.Status,
			StatusAfter:  order.Status,
			RequestData:  fmt.Sprintf("Refund request: amount=%.2f, reason=%s", amount, reason),
			Error:        err.Error(),
			UserID:       order.UserID,
		}

		_, logErr := s.logRepo.Create(ctx, errorLog)
		if logErr != nil {
			logger.Warnf(ctx, "Failed to create refund error log: %v", logErr)
		}

		return fmt.Errorf("failed to refund payment with provider: %w", err)
	}

	// Only update order if refund was successful
	if result.Success {
		// Get the old status for logging
		oldStatus := order.Status

		// Update order status
		order.Status = structs.PaymentStatusRefunded

		// Add refund information to metadata
		if order.Metadata == nil {
			order.Metadata = make(map[string]any)
		}

		refundInfo := map[string]any{
			"refund_id":     result.RefundID,
			"refund_amount": result.Amount,
			"refund_time":   time.Now().Format(time.RFC3339),
			"refund_reason": reason,
		}

		// Store refund information
		if existingRefunds, ok := order.Metadata["refunds"].([]map[string]any); ok {
			existingRefunds = append(existingRefunds, refundInfo)
			order.Metadata["refunds"] = existingRefunds
		} else {
			order.Metadata["refunds"] = []map[string]any{refundInfo}
		}

		// Save updated order
		_, err = s.repo.Update(ctx, &structs.UpdateOrderInput{
			Status:   order.Status,
			Metadata: order.Metadata,
		})
		if err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		// Log the refund
		logEntry := &structs.CreateLogInput{
			OrderID:      order.ID,
			ChannelID:    order.ChannelID,
			Type:         structs.LogTypeRefund,
			StatusBefore: oldStatus,
			StatusAfter:  order.Status,
			RequestData:  fmt.Sprintf("Refund request: amount=%.2f, reason=%s", amount, reason),
			ResponseData: fmt.Sprintf("%+v", result),
			UserID:       order.UserID,
		}

		_, err = s.logRepo.Create(ctx, logEntry)
		if err != nil {
			logger.Warnf(ctx, "Failed to create refund log: %v", err)
		}

		// Publish refund event
		if s.publisher != nil {
			eventData := &event.PaymentEventData{
				OrderID:        order.ID,
				OrderNumber:    order.OrderNumber,
				ChannelID:      order.ChannelID,
				Provider:       channel.Provider,
				UserID:         order.UserID,
				TenantID:       order.TenantID,
				Amount:         order.Amount,
				Currency:       order.Currency,
				Status:         order.Status,
				Type:           order.Type,
				ProductID:      order.ProductID,
				SubscriptionID: order.SubscriptionID,
				Metadata:       order.Metadata,
			}

			s.publisher.PublishPaymentRefunded(ctx, eventData)
		}
	}

	return nil
}

// GetByID gets a payment order by ID
func (s *orderService) GetByID(ctx context.Context, id string) (*structs.Order, error) {
	if id == "" {
		return nil, fmt.Errorf(ecode.FieldIsRequired("id"))
	}

	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment order: %w", err)
	}

	return s.Serialize(order), nil
}

// GetByOrderNumber gets a payment order by order number
func (s *orderService) GetByOrderNumber(ctx context.Context, orderNumber string) (*structs.Order, error) {
	if orderNumber == "" {
		return nil, fmt.Errorf(ecode.FieldIsRequired("order_number"))
	}

	order, err := s.repo.GetByOrderNumber(ctx, orderNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment order by number: %w", err)
	}

	return s.Serialize(order), nil
}

// List lists payment orders with pagination
func (s *orderService) List(ctx context.Context, query *structs.OrderQuery) (paging.Result[*structs.Order], error) {
	pp := paging.Params{
		Cursor:    query.Cursor,
		Limit:     query.PageSize,
		Direction: "forward", // Default direction
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.Order, int, error) {
		lq := *query
		lq.Cursor = cursor
		lq.PageSize = limit

		orders, err := s.repo.List(ctx, &lq)
		if err != nil {
			logger.Errorf(ctx, "Error listing orders: %v", err)
			return nil, 0, err
		}

		total, err := s.repo.Count(ctx, query)
		if err != nil {
			logger.Errorf(ctx, "Error counting orders: %v", err)
			return nil, 0, err
		}

		return s.Serializes(orders), int(total), nil
	})
}

// Serialize serializes a payment order to response format
func (s *orderService) Serialize(order *structs.Order) *structs.Order {
	if order == nil {
		return nil
	}

	return &structs.Order{
		ID:             order.ID,
		OrderNumber:    order.OrderNumber,
		Amount:         order.Amount,
		Currency:       order.Currency,
		Status:         order.Status,
		Type:           order.Type,
		ChannelID:      order.ChannelID,
		UserID:         order.UserID,
		TenantID:       order.TenantID,
		ProductID:      order.ProductID,
		SubscriptionID: order.SubscriptionID,
		ExpiresAt:      order.ExpiresAt,
		PaidAt:         order.PaidAt,
		ProviderRef:    order.ProviderRef,
		Description:    order.Description,
		Metadata:       order.Metadata,
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
	}
}

// Serializes serializes multiple payment orders to response format
func (s *orderService) Serializes(orders []*structs.Order) []*structs.Order {
	result := make([]*structs.Order, len(orders))
	for i, order := range orders {
		result[i] = s.Serialize(order)
	}
	return result
}

// generateOrderNumber generates a unique order number
func (s *orderService) generateOrderNumber() string {
	// Format: YYYYMMDDHHmmss + 6 random characters
	return time.Now().Format("20060102150405") + nanoid.String(6)
}
