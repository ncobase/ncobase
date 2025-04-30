package structs

// PaymentVerificationResult represents the result of payment verification
type PaymentVerificationResult struct {
	Success     bool           `json:"success"`
	Status      PaymentStatus  `json:"status"`
	Amount      float64        `json:"amount"`
	Currency    CurrencyCode   `json:"currency"`
	ProviderRef string         `json:"provider_ref,omitempty"`
	Error       string         `json:"error,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// WebhookResult represents the result of webhook processing
type WebhookResult struct {
	EventType   PaymentEventType `json:"event_type"`
	OrderID     string           `json:"order_id,omitempty"`
	Status      PaymentStatus    `json:"status,omitempty"`
	ProviderRef string           `json:"provider_ref,omitempty"`
	RawData     map[string]any   `json:"raw_data,omitempty"`
	Error       string           `json:"error,omitempty"`
}

// RefundResult represents the result of a refund request
type RefundResult struct {
	Success     bool          `json:"success"`
	RefundID    string        `json:"refund_id,omitempty"`
	Amount      float64       `json:"amount"`
	Currency    CurrencyCode  `json:"currency"`
	Status      PaymentStatus `json:"status"`
	ProviderRef string        `json:"provider_ref,omitempty"`
	Error       string        `json:"error,omitempty"`
}
