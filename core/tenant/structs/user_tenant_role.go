package structs

// UserTenantRole represents the user tenant role.
type UserTenantRole struct {
	UserID   string `json:"user_id,omitempty"`
	TenantID string `json:"tenant_id,omitempty"`
	RoleID   string `json:"role_id,omitempty"`
}

// AddUserToTenantRoleRequest represents the request to add a user to a tenant role
type AddUserToTenantRoleRequest struct {
	UserID string `json:"user_id" binding:"required"`
	RoleID string `json:"role_id" binding:"required"`
}

// UpdateUserTenantRoleRequest represents the request to update a user's tenant role
type UpdateUserTenantRoleRequest struct {
	OldRoleID string `json:"old_role_id" binding:"required"`
	NewRoleID string `json:"new_role_id" binding:"required"`
}

// BulkUpdateUserTenantRolesRequest represents the request to bulk update user tenant roles
type BulkUpdateUserTenantRolesRequest struct {
	Updates []UserTenantRoleUpdate `json:"updates" binding:"required"`
}

// UserTenantRoleUpdate represents a single user tenant role update
type UserTenantRoleUpdate struct {
	UserID    string `json:"user_id" binding:"required"`
	RoleID    string `json:"role_id" binding:"required"`
	Operation string `json:"operation" binding:"required"` // "add", "remove", "update"
	OldRoleID string `json:"old_role_id,omitempty"`        // Required for "update" operation
}

// UserTenantRoleResponse represents the response for user tenant role operations
type UserTenantRoleResponse struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	RoleID   string `json:"role_id"`
	Status   string `json:"status"`
}

// UserTenantRolesResponse represents the response for getting user roles in tenant
type UserTenantRolesResponse struct {
	UserID   string   `json:"user_id"`
	TenantID string   `json:"tenant_id"`
	RoleIDs  []string `json:"role_ids"`
	Count    int      `json:"count"`
}

// TenantRoleUsersResponse represents the response for getting users by role in tenant
type TenantRoleUsersResponse struct {
	TenantID string   `json:"tenant_id"`
	RoleID   string   `json:"role_id"`
	UserIDs  []string `json:"user_ids"`
	Count    int      `json:"count"`
}

// ListTenantUsersParams represents the parameters for listing tenant users
type ListTenantUsersParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	RoleID    string `form:"role_id,omitempty" json:"role_id,omitempty"`
	SortBy    string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}

// TenantUsersListResponse represents the response for listing tenant users
type TenantUsersListResponse struct {
	Users  []TenantUserInfo `json:"users"`
	Total  int              `json:"total"`
	Cursor string           `json:"cursor,omitempty"`
}

// TenantUserInfo represents user information with roles in tenant
type TenantUserInfo struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username,omitempty"`
	Email    string   `json:"email,omitempty"`
	RoleIDs  []string `json:"role_ids"`
	JoinedAt int64    `json:"joined_at,omitempty"`
}

// BulkUpdateResponse represents the response for bulk operations
type BulkUpdateResponse struct {
	Total   int                      `json:"total"`
	Success int                      `json:"success"`
	Failed  int                      `json:"failed"`
	Errors  []BulkUpdateError        `json:"errors,omitempty"`
	Results []UserTenantRoleResponse `json:"results,omitempty"`
}

// BulkUpdateError represents an error in bulk operations
type BulkUpdateError struct {
	UserID string `json:"user_id"`
	RoleID string `json:"role_id"`
	Error  string `json:"error"`
}
