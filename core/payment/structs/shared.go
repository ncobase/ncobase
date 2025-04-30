package structs

// PaymentStatus represents the status of a payment
type PaymentStatus string

// Payment statuses
const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
	PaymentStatusCancelled PaymentStatus = "cancelled"
)

// PaymentType represents the type of payment
type PaymentType string

// Payment types
const (
	PaymentTypeOneTime      PaymentType = "one_time"
	PaymentTypeSubscription PaymentType = "subscription"
	PaymentTypeRecurring    PaymentType = "recurring"
)

// CurrencyCode represents a currency code
type CurrencyCode string

// Common currency codes
const (
	CurrencyUSD CurrencyCode = "USD"
	CurrencyEUR CurrencyCode = "EUR"
	CurrencyGBP CurrencyCode = "GBP"
	CurrencyCNY CurrencyCode = "CNY"
	CurrencyJPY CurrencyCode = "JPY"
)

// PaymentProvider represents a payment provider type
type PaymentProvider string

// Payment providers
const (
	PaymentProviderStripe    PaymentProvider = "stripe"
	PaymentProviderPayPal    PaymentProvider = "paypal"
	PaymentProviderAlipay    PaymentProvider = "alipay"
	PaymentProviderWeChatPay PaymentProvider = "wechatpay"
)

// PaymentEventType represents the type of payment event
type PaymentEventType string

// Payment event types
const (
	PaymentEventCreated          PaymentEventType = "payment.created"
	PaymentEventSucceeded        PaymentEventType = "payment.succeeded"
	PaymentEventFailed           PaymentEventType = "payment.failed"
	PaymentEventRefunded         PaymentEventType = "payment.refunded"
	PaymentEventCancelled        PaymentEventType = "payment.cancelled"
	SubscriptionEventCreated     PaymentEventType = "subscription.created"
	SubscriptionEventRenewed     PaymentEventType = "subscription.renewed"
	SubscriptionEventExpired     PaymentEventType = "subscription.expired"
	SubscriptionEventCancelled   PaymentEventType = "subscription.cancelled"
	SubscriptionEventUpdated     PaymentEventType = "subscription.updated"
	ProductEventCreated          PaymentEventType = "product.created"
	ProductEventUpdated          PaymentEventType = "product.updated"
	ProductEventDeleted          PaymentEventType = "product.deleted"
	PaymentChannelEventCreated   PaymentEventType = "payment_channel.created"
	PaymentChannelEventUpdated   PaymentEventType = "payment_channel.updated"
	PaymentChannelEventDeleted   PaymentEventType = "payment_channel.deleted"
	PaymentChannelEventActivated PaymentEventType = "payment_channel.activated"
	PaymentChannelEventDisabled  PaymentEventType = "payment_channel.disabled"
)

// PaginationQuery defines common pagination fields
type PaginationQuery struct {
	Cursor    string `form:"cursor" json:"cursor,omitempty"`
	PageSize  int    `form:"page_size,default=20" json:"page_size,omitempty"`
	Direction string `form:"direction,default=forward" json:"direction,omitempty"`
}

type BaseModel struct {
	ID        string `json:"id"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
	DeletedAt int64  `json:"deleted_at,omitempty"`
}

// ProviderConfig represents the configuration for a payment provider
type ProviderConfig map[string]any
