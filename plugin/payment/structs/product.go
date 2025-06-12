package structs

import "fmt"

// ProductStatus represents the status of a product
type ProductStatus string

// Product statuses
const (
	ProductStatusActive   ProductStatus = "active"
	ProductStatusDisabled ProductStatus = "disabled"
	ProductStatusDraft    ProductStatus = "draft"
)

// PricingType represents the pricing model for a product
type PricingType string

// Pricing types
const (
	PricingTypeOneTime     PricingType = "one_time"
	PricingTypeRecurring   PricingType = "recurring"
	PricingTypeUsageBased  PricingType = "usage_based"
	PricingTypeTieredUsage PricingType = "tiered_usage"
)

// BillingInterval represents the billing interval for recurring payments
type BillingInterval string

// Billing intervals
const (
	BillingIntervalDaily   BillingInterval = "daily"
	BillingIntervalWeekly  BillingInterval = "weekly"
	BillingIntervalMonthly BillingInterval = "monthly"
	BillingIntervalYearly  BillingInterval = "yearly"
)

// Product represents a product or service that can be purchased
type Product struct {
	ID              string          `json:"id,omitempty"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	Status          ProductStatus   `json:"status"`
	PricingType     PricingType     `json:"pricing_type"`
	Price           float64         `json:"price"`
	Currency        CurrencyCode    `json:"currency"`
	BillingInterval BillingInterval `json:"billing_interval,omitempty"`
	TrialDays       int             `json:"trial_days,omitempty"`
	Features        []string        `json:"features,omitempty"`
	SpaceID         string          `json:"space_id,omitempty"`
	Metadata        map[string]any  `json:"metadata,omitempty"`
	CreatedAt       int64           `json:"created_at,omitempty"`
	UpdatedAt       int64           `json:"updated_at,omitempty"`
}

// GetCursorValue returns the cursor value for pagination
func (p *Product) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", p.ID, p.CreatedAt)
}

// CreateProductInput represents input for creating a product
type CreateProductInput struct {
	Name            string          `json:"name" binding:"required"`
	Description     string          `json:"description" binding:"required"`
	Status          ProductStatus   `json:"status" binding:"required"`
	PricingType     PricingType     `json:"pricing_type" binding:"required"`
	Price           float64         `json:"price" binding:"required"`
	Currency        CurrencyCode    `json:"currency" binding:"required"`
	BillingInterval BillingInterval `json:"billing_interval,omitempty"`
	TrialDays       int             `json:"trial_days,omitempty"`
	Features        []string        `json:"features,omitempty"`
	SpaceID         string          `json:"space_id,omitempty"`
	Metadata        map[string]any  `json:"metadata,omitempty"`
}

// UpdateProductInput represents input for updating a product
type UpdateProductInput struct {
	ID              string          `json:"id,omitempty"`
	Name            string          `json:"name,omitempty"`
	Description     string          `json:"description,omitempty"`
	Status          ProductStatus   `json:"status,omitempty"`
	Price           *float64        `json:"price,omitempty"`
	BillingInterval BillingInterval `json:"billing_interval,omitempty"`
	TrialDays       *int            `json:"trial_days,omitempty"`
	Features        []string        `json:"features,omitempty"`
	Metadata        map[string]any  `json:"metadata,omitempty"`
}

// ProductQuery represents query parameters for listing products
type ProductQuery struct {
	Status      ProductStatus `form:"status" json:"status,omitempty"`
	PricingType PricingType   `form:"pricing_type" json:"pricing_type,omitempty"`
	SpaceID     string        `form:"space_id" json:"space_id,omitempty"`
	PaginationQuery
}
