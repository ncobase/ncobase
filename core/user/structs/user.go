package structs

import (
	"fmt"

	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
)

// FindUser represents the parameters for finding a user.
type FindUser struct {
	ID       string `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
}

// UserBody represents the user schema.
type UserBody struct {
	ID          string      `json:"id"`
	Username    string      `json:"username"`
	Email       string      `json:"email,omitempty"`
	Phone       string      `json:"phone,omitempty"`
	IsCertified bool        `json:"is_certified"`
	IsAdmin     bool        `json:"is_admin"`
	Status      int         `json:"status"`
	Extras      *types.JSON `json:"extras"`
	CreatedAt   *int64      `json:"created_at"`
	UpdatedAt   *int64      `json:"updated_at"`
}

// UserPassword represents the user password schema.
type UserPassword struct {
	User        string `json:"user,omitempty" validate:"required"`
	OldPassword string `json:"old_password,omitempty"`
	NewPassword string `json:"new_password,omitempty" validate:"required"`
	Confirm     string `json:"confirm,omitempty" validate:"required,eqfield=NewPassword"`
}

// ReadUser represents the user schema.
type ReadUser struct {
	ID          string      `json:"id"`
	Username    string      `json:"username"`
	Email       string      `json:"email,omitempty"`
	Phone       string      `json:"phone,omitempty"`
	IsCertified bool        `json:"is_certified"`
	IsAdmin     bool        `json:"is_admin"`
	Status      int         `json:"status"`
	Extras      *types.JSON `json:"extras"`
	CreatedAt   *int64      `json:"created_at"`
	UpdatedAt   *int64      `json:"updated_at"`
}

// GetCursorValue returns the cursor value.
func (r *ReadUser) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, convert.ToValue(r.CreatedAt))
}

// ListUserParams represents the query parameters for listing users.
type ListUserParams struct {
	Cursor       string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit        int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction    string `form:"direction,omitempty" json:"direction,omitempty"`
	SearchQuery  string `form:"search,omitempty" json:"search,omitempty"`
	RoleFilter   string `form:"role,omitempty" json:"role,omitempty"`
	StatusFilter string `form:"status_filter,omitempty" json:"status_filter,omitempty"`
	SortBy       string `form:"sortBy,omitempty" json:"sortBy,omitempty"`
	Status       int    `form:"status,omitempty" json:"status,omitempty"`
}

// PasswordResetRequest represents the request to reset a user's password
type PasswordResetRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}
