// Package payment provides payment processing and subscription management for the Ncobase platform.
//
// This package implements a comprehensive payment system with support for multiple
// payment providers (Stripe, Alipay, WeChat Pay), subscription management, and
// payment event handling. It provides a unified API for payment operations across
// different payment gateways.
//
// Key Features:
//   - Multi-provider payment processing (Stripe, Alipay, WeChat Pay)
//   - Subscription management and billing
//   - Product and pricing management
//   - Payment channel configuration
//   - Webhook handling for payment events
//   - Payment history and reporting
//   - Refund and cancellation support
//   - Invoice generation
//
// Main Components:
//   - Handler: HTTP endpoints for payment operations
//   - Service: Business logic for payment processing
//   - Repository: Data access layer for payment entities
//   - Events: Payment event publishers and handlers
//
// Supported Payment Providers:
//   - Stripe: Credit cards, subscriptions, invoices
//   - Alipay: Chinese payment gateway
//   - WeChat Pay: Chinese mobile payment
//
// Payment Entities:
//   - Payment: Individual payment transactions
//   - Subscription: Recurring billing subscriptions
//   - Product: Items available for purchase
//   - Channel: Payment provider configurations
//   - Invoice: Payment invoices and receipts
//
// Usage Example:
//
//	// Create a payment
//	payment, err := paymentService.CreatePayment(ctx, &structs.CreatePaymentInput{
//	    Amount:      9999, // $99.99 in cents
//	    Currency:    "USD",
//	    Provider:    "stripe",
//	    Description: "Premium subscription",
//	    CustomerID:  customerID,
//	})
//	if err != nil {
//	    return err
//	}
//
//	// Create subscription
//	subscription, err := paymentService.CreateSubscription(ctx, &structs.CreateSubscriptionInput{
//	    CustomerID: customerID,
//	    ProductID:  productID,
//	    PriceID:    priceID,
//	    Provider:   "stripe",
//	})
//
//	// Handle webhook event
//	err = paymentService.HandleWebhook(ctx, &structs.WebhookInput{
//	    Provider: "stripe",
//	    Event:    "payment.succeeded",
//	    Data:     webhookData,
//	})
//
// Event System:
//   - payment.created, payment.succeeded, payment.failed
//   - subscription.created, subscription.updated, subscription.cancelled
//   - product.created, product.updated, product.deleted
//   - channel.created, channel.updated, channel.deleted
//
// The package provides enterprise-grade payment processing with support for
// multiple providers, comprehensive event handling, and flexible subscription management.
package payment
