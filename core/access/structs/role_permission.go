package structs

// RolePermission represents the role permission relationship.
type RolePermission struct {
	RoleID       string `json:"role_id,omitempty"`
	PermissionID string `json:"permission_id,omitempty"`
}
