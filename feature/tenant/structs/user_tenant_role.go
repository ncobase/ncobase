package structs

// UserTenantRole represents the user tenant role.
type UserTenantRole struct {
	UserID   string `json:"user_id,omitempty"`
	TenantID string `json:"tenant_id,omitempty"`
	RoleID   string `json:"role_id,omitempty"`
}
