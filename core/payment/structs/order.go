package structs

import (
	"fmt"
	"time"
)

// Order represents a payment order
type Order struct {
	ID             string         `json:"id,omitempty"`
	OrderNumber    string         `json:"order_number"`
	Amount         float64        `json:"amount"`
	Currency       CurrencyCode   `json:"currency"`
	Status         PaymentStatus  `json:"status"`
	Type           PaymentType    `json:"type"`
	ChannelID      string         `json:"channel_id"`
	UserID         string         `json:"user_id"`
	TenantID       string         `json:"tenant_id,omitempty"`
	ProductID      string         `json:"product_id,omitempty"`
	SubscriptionID string         `json:"subscription_id,omitempty"`
	ExpiresAt      time.Time      `json:"expires_at"`
	PaidAt         time.Time      `json:"paid_at,omitempty"`
	ProviderRef    string         `json:"provider_ref,omitempty"`
	Description    string         `json:"description,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	CreatedAt      int64          `json:"created_at,omitempty"`
	UpdatedAt      int64          `json:"updated_at,omitempty"`
}

// GetCursorValue returns the cursor value for pagination
func (o *Order) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", o.ID, o.CreatedAt)
}

// CreateOrderInput represents input for creating a payment order
type CreateOrderInput struct {
	OrderNumber    string         `json:"order_number"`
	Amount         float64        `json:"amount"`
	Currency       CurrencyCode   `json:"currency"`
	Status         PaymentStatus  `json:"status"`
	Type           PaymentType    `json:"type"`
	ChannelID      string         `json:"channel_id"`
	UserID         string         `json:"user_id"`
	TenantID       string         `json:"tenant_id,omitempty"`
	ProductID      string         `json:"product_id,omitempty"`
	SubscriptionID string         `json:"subscription_id,omitempty"`
	ExpiresAt      time.Time      `json:"expires_at"`
	PaidAt         time.Time      `json:"paid_at,omitempty"`
	ProviderRef    string         `json:"provider_ref,omitempty"`
	Description    string         `json:"description,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

// UpdateOrderInput represents input for updating an order
type UpdateOrderInput struct {
	ID             string         `json:"id,omitempty"`
	OrderNumber    string         `json:"order_number,omitempty"`
	Amount         *float64       `json:"amount,omitempty"`
	Currency       CurrencyCode   `json:"currency,omitempty"`
	Status         PaymentStatus  `json:"status,omitempty"`
	Type           PaymentType    `json:"type,omitempty"`
	ChannelID      string         `json:"channel_id,omitempty"`
	UserID         string         `json:"user_id,omitempty"`
	TenantID       string         `json:"tenant_id,omitempty"`
	ProductID      string         `json:"product_id,omitempty"`
	SubscriptionID string         `json:"subscription_id,omitempty"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
	PaidAt         *time.Time     `json:"paid_at,omitempty"`
	ProviderRef    string         `json:"provider_ref,omitempty"`
	Description    string         `json:"description,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

// RefundOrderInput represents input for refunding a payment
type RefundOrderInput struct {
	Amount float64 `json:"amount" binding:"min=0"`
	Reason string  `json:"reason"`
}

// OrderQuery represents query parameters for listing orders
type OrderQuery struct {
	Status         PaymentStatus `form:"status" json:"status,omitempty"`
	Type           PaymentType   `form:"type" json:"type,omitempty"`
	ChannelID      string        `form:"channel_id" json:"channel_id,omitempty"`
	UserID         string        `form:"user_id" json:"user_id,omitempty"`
	TenantID       string        `form:"tenant_id" json:"tenant_id,omitempty"`
	ProductID      string        `form:"product_id" json:"product_id,omitempty"`
	SubscriptionID string        `form:"subscription_id" json:"subscription_id,omitempty"`
	StartDate      int64         `form:"start_date" json:"start_date,omitempty"`
	EndDate        int64         `form:"end_date" json:"end_date,omitempty"`
	PaginationQuery
}

// OrderSummary represents a summary of order stats
type OrderSummary struct {
	TotalCount    int64   `json:"total_count"`
	SuccessCount  int64   `json:"success_count"`
	FailedCount   int64   `json:"failed_count"`
	TotalAmount   float64 `json:"total_amount"`
	SuccessAmount float64 `json:"success_amount"`
	Currency      string  `json:"currency"`
	PeriodStart   string  `json:"period_start"`
	PeriodEnd     string  `json:"period_end"`
}
