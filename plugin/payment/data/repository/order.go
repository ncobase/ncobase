package repository

import (
	"context"
	"fmt"
	"ncobase/plugin/payment/data"
	"ncobase/plugin/payment/data/ent"
	paymentOrderEnt "ncobase/plugin/payment/data/ent/paymentorder"
	"ncobase/plugin/payment/structs"
	"time"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/utils/convert"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"
)

// OrderRepositoryInterface defines the interface for order repository operations
type OrderRepositoryInterface interface {
	Create(ctx context.Context, order *structs.CreateOrderInput) (*structs.Order, error)
	GetByID(ctx context.Context, id string) (*structs.Order, error)
	GetByOrderNumber(ctx context.Context, orderNumber string) (*structs.Order, error)
	Update(ctx context.Context, order *structs.UpdateOrderInput) (*structs.Order, error)
	List(ctx context.Context, query *structs.OrderQuery) ([]*structs.Order, error)
	Count(ctx context.Context, query *structs.OrderQuery) (int64, error)
	GetOrderSummary(ctx context.Context, startDate, endDate int64, currency string) (*structs.OrderSummary, error)
}

// orderRepository handles payment order persistence
type orderRepository struct {
	data *data.Data
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(d *data.Data) OrderRepositoryInterface {
	return &orderRepository{data: d}
}

// Create creates a new payment order
func (r *orderRepository) Create(ctx context.Context, order *structs.CreateOrderInput) (*structs.Order, error) {

	builder := r.data.EC.PaymentOrder.Create().
		SetOrderNumber(order.OrderNumber).
		SetAmount(order.Amount).
		SetCurrency(string(order.Currency)).
		SetStatus(string(order.Status)).
		SetType(string(order.Type)).
		SetChannelID(order.ChannelID).
		SetUserID(order.UserID).
		SetExpiresAt(order.ExpiresAt)

	// Set optional fields
	if order.SpaceID != "" {
		builder.SetSpaceID(order.SpaceID)
	}

	if order.ProductID != "" {
		builder.SetProductID(order.ProductID)
	}

	if order.SubscriptionID != "" {
		builder.SetSubscriptionID(order.SubscriptionID)
	}

	if !order.PaidAt.IsZero() {
		builder.SetPaidAt(order.PaidAt)
	}

	if order.ProviderRef != "" {
		builder.SetProviderRef(order.ProviderRef)
	}

	if order.Description != "" {
		builder.SetDescription(order.Description)
	}

	if validator.IsNotEmpty(order.Metadata) {
		builder.SetExtras(order.Metadata)
	}

	// Create order
	row, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment order: %w", err)
	}

	// Convert to struct
	result, err := r.entToStruct(row)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetByID gets a payment order by ID
func (r *orderRepository) GetByID(ctx context.Context, id string) (*structs.Order, error) {
	order, err := r.data.EC.PaymentOrder.Query().
		Where(paymentOrderEnt.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment order: %w", err)
	}

	return r.entToStruct(order)
}

// GetByOrderNumber gets a payment order by order number
func (r *orderRepository) GetByOrderNumber(ctx context.Context, orderNumber string) (*structs.Order, error) {
	order, err := r.data.EC.PaymentOrder.Query().
		Where(paymentOrderEnt.OrderNumber(orderNumber)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment order by number: %w", err)
	}

	return r.entToStruct(order)
}

// Update updates a payment order
func (r *orderRepository) Update(ctx context.Context, order *structs.UpdateOrderInput) (*structs.Order, error) {

	builder := r.data.EC.PaymentOrder.UpdateOneID(order.ID).
		SetStatus(string(order.Status))

	// Update fields that might have changed
	if order.ProviderRef != "" {
		builder.SetProviderRef(order.ProviderRef)
	}

	if !order.PaidAt.IsZero() {
		builder.SetPaidAt(convert.ToValue(order.PaidAt))
	}

	if validator.IsNotEmpty(order.Metadata) {
		builder.SetExtras(order.Metadata)
	}

	// Update order
	row, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update payment order: %w", err)
	}

	// Convert to struct
	result, err := r.entToStruct(row)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// List lists payment orders
func (r *orderRepository) List(ctx context.Context, query *structs.OrderQuery) ([]*structs.Order, error) {
	// Build query with cursor support
	q := r.data.EC.PaymentOrder.Query()

	// Apply filters
	if query.Status != "" {
		q = q.Where(paymentOrderEnt.Status(string(query.Status)))
	}

	if query.Type != "" {
		q = q.Where(paymentOrderEnt.Type(string(query.Type)))
	}

	if query.ChannelID != "" {
		q = q.Where(paymentOrderEnt.ChannelID(query.ChannelID))
	}

	if query.UserID != "" {
		q = q.Where(paymentOrderEnt.UserID(query.UserID))
	}

	if query.SpaceID != "" {
		q = q.Where(paymentOrderEnt.SpaceID(query.SpaceID))
	}

	if query.ProductID != "" {
		q = q.Where(paymentOrderEnt.ProductID(query.ProductID))
	}

	if query.SubscriptionID != "" {
		q = q.Where(paymentOrderEnt.SubscriptionID(query.SubscriptionID))
	}

	if query.StartDate > 0 {
		q = q.Where(paymentOrderEnt.CreatedAtGTE(query.StartDate))
	}

	if query.EndDate > 0 {
		q = q.Where(paymentOrderEnt.CreatedAtLTE(query.EndDate))
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
			paymentOrderEnt.Or(
				paymentOrderEnt.CreatedAtLT(timestamp),
				paymentOrderEnt.And(
					paymentOrderEnt.CreatedAtEQ(timestamp),
					paymentOrderEnt.IDLT(id),
				),
			),
		)
	}

	// Set order - most recent first by default
	q.Order(ent.Desc(paymentOrderEnt.FieldCreatedAt), ent.Desc(paymentOrderEnt.FieldID))

	// Set limit
	if query.PageSize > 0 {
		q.Limit(query.PageSize)
	} else {
		q.Limit(20) // Default page size
	}

	// Execute the query
	rows, err := q.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list payment orders: %w", err)
	}

	// Convert to structs
	var orders []*structs.Order
	for _, row := range rows {
		order, err := r.entToStruct(row)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// Count counts payment orders
func (r *orderRepository) Count(ctx context.Context, query *structs.OrderQuery) (int64, error) {
	// Build query without cursor, limit, and order
	q := r.data.EC.PaymentOrder.Query()

	// Apply filters
	if query.Status != "" {
		q = q.Where(paymentOrderEnt.Status(string(query.Status)))
	}

	if query.Type != "" {
		q = q.Where(paymentOrderEnt.Type(string(query.Type)))
	}

	if query.ChannelID != "" {
		q = q.Where(paymentOrderEnt.ChannelID(query.ChannelID))
	}

	if query.UserID != "" {
		q = q.Where(paymentOrderEnt.UserID(query.UserID))
	}

	if query.SpaceID != "" {
		q = q.Where(paymentOrderEnt.SpaceID(query.SpaceID))
	}

	if query.ProductID != "" {
		q = q.Where(paymentOrderEnt.ProductID(query.ProductID))
	}

	if query.SubscriptionID != "" {
		q = q.Where(paymentOrderEnt.SubscriptionID(query.SubscriptionID))
	}

	if query.StartDate > 0 {
		q = q.Where(paymentOrderEnt.CreatedAtGTE(query.StartDate))
	}

	if query.EndDate > 0 {
		q = q.Where(paymentOrderEnt.CreatedAtLTE(query.EndDate))
	}

	// Execute the count
	count, err := q.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count payment orders: %w", err)
	}

	return int64(count), nil
}

// GetOrderSummary gets a summary of order stats
func (r *orderRepository) GetOrderSummary(ctx context.Context, startDate, endDate int64, currency string) (*structs.OrderSummary, error) {
	// Get total count
	totalCount, err := r.data.EC.PaymentOrder.Query().
		Where(
			paymentOrderEnt.CreatedAtGTE(startDate),
			paymentOrderEnt.CreatedAtLTE(endDate),
			paymentOrderEnt.Currency(currency),
		).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get order count: %w", err)
	}

	// Get success count
	successCount, err := r.data.EC.PaymentOrder.Query().
		Where(
			paymentOrderEnt.Status(string(structs.PaymentStatusCompleted)),
			paymentOrderEnt.CreatedAtGTE(startDate),
			paymentOrderEnt.CreatedAtLTE(endDate),
			paymentOrderEnt.Currency(currency),
		).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get success count: %w", err)
	}

	// Get failed count
	failedCount, err := r.data.EC.PaymentOrder.Query().
		Where(
			paymentOrderEnt.Status(string(structs.PaymentStatusFailed)),
			paymentOrderEnt.CreatedAtGTE(startDate),
			paymentOrderEnt.CreatedAtLTE(endDate),
			paymentOrderEnt.Currency(currency),
		).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed count: %w", err)
	}

	// Get total amount
	var totalAmount float64
	orders, err := r.data.EC.PaymentOrder.Query().
		Where(
			paymentOrderEnt.CreatedAtGTE(startDate),
			paymentOrderEnt.CreatedAtLTE(endDate),
			paymentOrderEnt.Currency(currency),
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders for total amount: %w", err)
	}
	for _, order := range orders {
		totalAmount += order.Amount
	}

	// Get success amount
	var successAmount float64
	successOrders, err := r.data.EC.PaymentOrder.Query().
		Where(
			paymentOrderEnt.Status(string(structs.PaymentStatusCompleted)),
			paymentOrderEnt.CreatedAtGTE(startDate),
			paymentOrderEnt.CreatedAtLTE(endDate),
			paymentOrderEnt.Currency(currency),
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get successful orders for amount: %w", err)
	}
	for _, order := range successOrders {
		successAmount += order.Amount
	}

	return &structs.OrderSummary{
		TotalCount:    int64(totalCount),
		SuccessCount:  int64(successCount),
		FailedCount:   int64(failedCount),
		TotalAmount:   totalAmount,
		SuccessAmount: successAmount,
		Currency:      currency,
		PeriodStart:   *convert.UnixMilliToString(&startDate, time.RFC3339),
		PeriodEnd:     *convert.UnixMilliToString(&endDate, time.RFC3339),
	}, nil
}

// entToStruct converts an Ent PaymentOrder to a structs.Order
func (r *orderRepository) entToStruct(order *ent.PaymentOrder) (*structs.Order, error) {
	return &structs.Order{
		ID:             order.ID,
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
		OrderNumber:    order.OrderNumber,
		Amount:         order.Amount,
		Currency:       structs.CurrencyCode(order.Currency),
		Status:         structs.PaymentStatus(order.Status),
		Type:           structs.PaymentType(order.Type),
		ChannelID:      order.ChannelID,
		UserID:         order.UserID,
		SpaceID:        order.SpaceID,
		ProductID:      order.ProductID,
		SubscriptionID: order.SubscriptionID,
		ExpiresAt:      order.ExpiresAt,
		PaidAt:         *order.PaidAt,
		ProviderRef:    order.ProviderRef,
		Description:    order.Description,
		Metadata:       order.Extras,
	}, nil
}
