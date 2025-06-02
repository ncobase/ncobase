package structs

// TenantMenu represents the tenant menu relationship.
type TenantMenu struct {
	TenantID string `json:"tenant_id,omitempty"`
	MenuID   string `json:"menu_id,omitempty"`
}

// AddMenuToTenantRequest represents the request to add a menu to a tenant
type AddMenuToTenantRequest struct {
	MenuID string `json:"menu_id" binding:"required"`
}
