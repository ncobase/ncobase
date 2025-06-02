package structs

// TenantOption represents the tenant option relationship.
type TenantOption struct {
	TenantID string `json:"tenant_id,omitempty"`
	OptionID string `json:"option_id,omitempty"`
}

// AddOptionsToTenantRequest represents the request to add options to a tenant
type AddOptionsToTenantRequest struct {
	OptionID string `json:"option_id" binding:"required"`
}
