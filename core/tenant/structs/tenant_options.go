package structs

// TenantOptions represents the tenant options relationship.
type TenantOptions struct {
	TenantID  string `json:"tenant_id,omitempty"`
	OptionsID string `json:"options_id,omitempty"`
}

// AddOptionsToTenantRequest represents the request to add options to a tenant
type AddOptionsToTenantRequest struct {
	OptionsID string `json:"options_id" binding:"required"`
}
