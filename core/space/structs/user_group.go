package structs

// UserRole represents a user role within a group.
type UserRole string

const (
	RoleOwner  UserRole = "owner"
	RoleAdmin  UserRole = "admin"
	RoleEditor UserRole = "editor"
	RoleViewer UserRole = "viewer"
	RoleMember UserRole = "member"
)

// IsValidUserRole checks if a role is valid
func IsValidUserRole(role UserRole) bool {
	return role == RoleOwner || role == RoleAdmin || role == RoleEditor || role == RoleViewer || role == RoleMember
}

// UserGroup represents the user group.
type UserGroup struct {
	UserID  string   `json:"user_id,omitempty"`
	GroupID string   `json:"group_id,omitempty"`
	Role    UserRole `json:"role,omitempty"`
}

// AddMemberRequest represents the request to add a member to a group
type AddMemberRequest struct {
	UserID string   `json:"user_id" binding:"required"`
	Role   UserRole `json:"role" binding:"required"`
}

// UpdateMemberRequest represents the request to update a member's role
type UpdateMemberRequest struct {
	Role UserRole `json:"role" binding:"required"`
}

// GroupMember represents a user membership in a group.
type GroupMember struct {
	ID        string   `json:"id,omitempty"`
	UserID    string   `json:"user_id,omitempty"`
	Name      string   `json:"name,omitempty"`
	Email     string   `json:"email,omitempty"`
	Role      UserRole `json:"role,omitempty"`
	AddedAt   int64    `json:"added_at,omitempty"`
	Avatar    string   `json:"avatar,omitempty"`
	LastLogin *int64   `json:"last_login,omitempty"`
}
