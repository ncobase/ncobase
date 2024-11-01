package structs

// UserTenant represents the user tenant.
type UserTenant struct {
	UserID   string `json:"user_id,omitempty"`
	TenantID string `json:"tenant_id,omitempty"`
}
