package structs

// UserSpaceRole represents the user space role.
type UserSpaceRole struct {
	UserID  string `json:"user_id,omitempty"`
	SpaceID string `json:"space_id,omitempty"`
	RoleID  string `json:"role_id,omitempty"`
}

// AddUserToSpaceRoleRequest represents the request to add a user to a space role
type AddUserToSpaceRoleRequest struct {
	UserID string `json:"user_id" binding:"required"`
	RoleID string `json:"role_id" binding:"required"`
}

// UpdateUserSpaceRoleRequest represents the request to update a user's space role
type UpdateUserSpaceRoleRequest struct {
	OldRoleID string `json:"old_role_id" binding:"required"`
	NewRoleID string `json:"new_role_id" binding:"required"`
}

// BulkUpdateUserSpaceRolesRequest represents the request to bulk update user space roles
type BulkUpdateUserSpaceRolesRequest struct {
	Updates []UserSpaceRoleUpdate `json:"updates" binding:"required"`
}

// UserSpaceRoleUpdate represents a single user space role update
type UserSpaceRoleUpdate struct {
	UserID    string `json:"user_id" binding:"required"`
	RoleID    string `json:"role_id" binding:"required"`
	Operation string `json:"operation" binding:"required"` // "add", "remove", "update"
	OldRoleID string `json:"old_role_id,omitempty"`        // Required for "update" operation
}

// UserSpaceRoleResponse represents the response for user space role operations
type UserSpaceRoleResponse struct {
	UserID  string `json:"user_id"`
	SpaceID string `json:"space_id"`
	RoleID  string `json:"role_id"`
	Status  string `json:"status"`
}

// UserSpaceRolesResponse represents the response for getting user roles in space
type UserSpaceRolesResponse struct {
	UserID  string   `json:"user_id"`
	SpaceID string   `json:"space_id"`
	RoleIDs []string `json:"role_ids"`
	Count   int      `json:"count"`
}

// SpaceRoleUsersResponse represents the response for getting users by role in space
type SpaceRoleUsersResponse struct {
	SpaceID string   `json:"space_id"`
	RoleID  string   `json:"role_id"`
	UserIDs []string `json:"user_ids"`
	Count   int      `json:"count"`
}

// ListSpaceUsersParams represents the parameters for listing space users
type ListSpaceUsersParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	RoleID    string `form:"role_id,omitempty" json:"role_id,omitempty"`
	SortBy    string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}

// SpaceUsersListResponse represents the response for listing space users
type SpaceUsersListResponse struct {
	Users  []SpaceUserInfo `json:"users"`
	Total  int             `json:"total"`
	Cursor string          `json:"cursor,omitempty"`
}

// SpaceUserInfo represents user information with roles in space
type SpaceUserInfo struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username,omitempty"`
	Email    string   `json:"email,omitempty"`
	RoleIDs  []string `json:"role_ids"`
	JoinedAt int64    `json:"joined_at,omitempty"`
}

// BulkUpdateResponse represents the response for bulk operations
type BulkUpdateResponse struct {
	Total   int                     `json:"total"`
	Success int                     `json:"success"`
	Failed  int                     `json:"failed"`
	Errors  []BulkUpdateError       `json:"errors,omitempty"`
	Results []UserSpaceRoleResponse `json:"results,omitempty"`
}

// BulkUpdateError represents an error in bulk operations
type BulkUpdateError struct {
	UserID string `json:"user_id"`
	RoleID string `json:"role_id"`
	Error  string `json:"error"`
}
