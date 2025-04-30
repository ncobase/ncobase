package provider

import (
	"context"
	"ncobase/core/payment/structs"
)

// Provider defines the interface for payment providers
type Provider interface {
	// Initialize sets up the provider with configuration
	Initialize(config structs.ProviderConfig) error

	// GetName returns the name of the provider
	GetName() string

	// CreatePayment creates a payment with the provider
	CreatePayment(ctx context.Context, order *structs.Order) (string, map[string]any, error)

	// VerifyPayment verifies a payment with the provider
	VerifyPayment(ctx context.Context, orderID string, data map[string]any) (*structs.PaymentVerificationResult, error)

	// ProcessWebhook processes a webhook event from the provider
	ProcessWebhook(ctx context.Context, payload []byte, headers map[string]string) (*structs.WebhookResult, error)

	// RefundPayment requests a refund for a payment
	RefundPayment(ctx context.Context, orderID string, amount float64, reason string) (*structs.RefundResult, error)

	// CreateSubscription creates a subscription with the provider
	CreateSubscription(ctx context.Context, subscription *structs.CreateSubscriptionInput) (string, map[string]any, error)

	// UpdateSubscription updates a subscription with the provider
	UpdateSubscription(ctx context.Context, subscriptionID string, updates map[string]any) error

	// CancelSubscription cancels a subscription with the provider
	CancelSubscription(ctx context.Context, subscriptionID string, immediate bool) error

	// GetPaymentStatus gets the current status of a payment
	GetPaymentStatus(ctx context.Context, orderID string) (structs.PaymentStatus, error)
}

// ProviderFactory is a function that creates a new payment provider
type ProviderFactory func(config structs.ProviderConfig) (Provider, error)
