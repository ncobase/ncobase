package event

import (
	"ncobase/plugin/payment/structs"
	"time"
)

// PaymentEventData represents data included in payment events
type PaymentEventData struct {
	Timestamp      time.Time               `json:"timestamp"`
	OrderID        string                  `json:"order_id"`
	OrderNumber    string                  `json:"order_number"`
	ChannelID      string                  `json:"channel_id"`
	Provider       structs.PaymentProvider `json:"provider"`
	UserID         string                  `json:"user_id"`
	SpaceID        string                  `json:"space_id,omitempty"`
	Amount         float64                 `json:"amount"`
	Currency       structs.CurrencyCode    `json:"currency"`
	Status         structs.PaymentStatus   `json:"status"`
	Type           structs.PaymentType     `json:"type"`
	ProductID      string                  `json:"product_id,omitempty"`
	SubscriptionID string                  `json:"subscription_id,omitempty"`
	Metadata       map[string]any          `json:"metadata,omitempty"`
}

// NewPaymentEventData creates a new payment event data instance
func NewPaymentEventData(
	orderID, orderNumber, channelID string,
	provider structs.PaymentProvider,
	userID, spaceID string,
	amount float64,
	currency structs.CurrencyCode,
	status structs.PaymentStatus,
	paymentType structs.PaymentType,
	productID, subscriptionID string,
	metadata map[string]any,
) *PaymentEventData {
	return &PaymentEventData{
		Timestamp:      time.Now(),
		OrderID:        orderID,
		OrderNumber:    orderNumber,
		ChannelID:      channelID,
		Provider:       provider,
		UserID:         userID,
		SpaceID:        spaceID,
		Amount:         amount,
		Currency:       currency,
		Status:         status,
		Type:           paymentType,
		ProductID:      productID,
		SubscriptionID: subscriptionID,
		Metadata:       metadata,
	}
}

// SubscriptionEventData represents data included in subscription events
type SubscriptionEventData struct {
	Timestamp          time.Time                  `json:"timestamp"`
	SubscriptionID     string                     `json:"subscription_id"`
	UserID             string                     `json:"user_id"`
	SpaceID            string                     `json:"space_id,omitempty"`
	ProductID          string                     `json:"product_id"`
	ChannelID          string                     `json:"channel_id"`
	Provider           structs.PaymentProvider    `json:"provider"`
	Status             structs.SubscriptionStatus `json:"status"`
	CurrentPeriodStart time.Time                  `json:"current_period_start"`
	CurrentPeriodEnd   time.Time                  `json:"current_period_end"`
	Metadata           map[string]any             `json:"metadata,omitempty"`
}

// NewSubscriptionEventData creates a new subscription event data instance
func NewSubscriptionEventData(
	subscriptionID, userID, spaceID, productID, channelID string,
	provider structs.PaymentProvider,
	status structs.SubscriptionStatus,
	currentPeriodStart, currentPeriodEnd time.Time,
	metadata map[string]any,
) *SubscriptionEventData {
	return &SubscriptionEventData{
		Timestamp:          time.Now(),
		SubscriptionID:     subscriptionID,
		UserID:             userID,
		SpaceID:            spaceID,
		ProductID:          productID,
		ChannelID:          channelID,
		Provider:           provider,
		Status:             status,
		CurrentPeriodStart: currentPeriodStart,
		CurrentPeriodEnd:   currentPeriodEnd,
		Metadata:           metadata,
	}
}

// ChannelEventData represents data included in payment channel events
type ChannelEventData struct {
	Timestamp time.Time               `json:"timestamp"`
	ChannelID string                  `json:"channel_id"`
	Name      string                  `json:"name"`
	Provider  structs.PaymentProvider `json:"provider"`
	Status    structs.ChannelStatus   `json:"status"`
	IsDefault bool                    `json:"is_default"`
	SpaceID   string                  `json:"space_id,omitempty"`
	Metadata  map[string]any          `json:"metadata,omitempty"`
}

// NewChannelEventData creates a new channel event data instance
func NewChannelEventData(
	channelID, name string,
	provider structs.PaymentProvider,
	status structs.ChannelStatus,
	isDefault bool,
	spaceID string,
	metadata map[string]any,
) *ChannelEventData {
	return &ChannelEventData{
		Timestamp: time.Now(),
		ChannelID: channelID,
		Name:      name,
		Provider:  provider,
		Status:    status,
		IsDefault: isDefault,
		SpaceID:   spaceID,
		Metadata:  metadata,
	}
}

// ProductEventData represents data included in product events
type ProductEventData struct {
	Timestamp       time.Time               `json:"timestamp"`
	ProductID       string                  `json:"product_id"`
	Name            string                  `json:"name"`
	Status          structs.ProductStatus   `json:"status"`
	PricingType     structs.PricingType     `json:"pricing_type"`
	Price           float64                 `json:"price"`
	Currency        structs.CurrencyCode    `json:"currency"`
	BillingInterval structs.BillingInterval `json:"billing_interval,omitempty"`
	SpaceID         string                  `json:"space_id,omitempty"`
	Metadata        map[string]any          `json:"metadata,omitempty"`
}

// NewProductEventData creates a new product event data instance
func NewProductEventData(
	productID, name string,
	status structs.ProductStatus,
	pricingType structs.PricingType,
	price float64,
	currency structs.CurrencyCode,
	billingInterval structs.BillingInterval,
	spaceID string,
	metadata map[string]any,
) *ProductEventData {
	return &ProductEventData{
		Timestamp:       time.Now(),
		ProductID:       productID,
		Name:            name,
		Status:          status,
		PricingType:     pricingType,
		Price:           price,
		Currency:        currency,
		BillingInterval: billingInterval,
		SpaceID:         spaceID,
		Metadata:        metadata,
	}
}
