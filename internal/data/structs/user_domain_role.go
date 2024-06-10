package structs

// UserDomainRole represents the user domain role.
type UserDomainRole struct {
	UserID   string `json:"user_id,omitempty"`
	DomainID string `json:"domain_id,omitempty"`
	RoleID   string `json:"role_id,omitempty"`
}
