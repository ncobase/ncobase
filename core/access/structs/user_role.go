package structs

// UserRole represents the user role.
type UserRole struct {
	UserID string `json:"user_id,omitempty"`
	RoleID string `json:"role_id,omitempty"`
}
