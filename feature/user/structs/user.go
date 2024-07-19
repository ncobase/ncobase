package structs

import (
	"time"

	"ncobase/common/types"
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
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
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
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// ListUserParams represents the query parameters for listing users.
type ListUserParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Offset    int    `form:"offset,omitempty" json:"offset,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
}
