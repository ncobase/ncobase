package structs

import (
	"ncobase/common/types"
	"time"
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
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email,omitempty"`
	Phone       string    `json:"phone,omitempty"`
	IsCertified bool      `json:"is_certified"`
	IsAdmin     bool      `json:"is_admin"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserPassword represents the user password schema.
type UserPassword struct {
	User        string `json:"user,omitempty" validate:"required"`
	OldPassword string `json:"old_password,omitempty"`
	NewPassword string `json:"new_password,omitempty" validate:"required"`
	Confirm     string `json:"confirm,omitempty" validate:"required,eqfield=NewPassword"`
}

// UserProfileBody represents the user profile schema.
type UserProfileBody struct {
	ID          string        `json:"id,omitempty"`
	DisplayName string        `json:"display_name,omitempty"`
	ShortBio    string        `json:"short_bio,omitempty"`
	About       *string       `json:"about,omitempty"`
	Thumbnail   *string       `json:"thumbnail,omitempty"`
	Links       *[]types.JSON `json:"links,omitempty"`
}

// UserTenant represents the user tenant.
type UserTenant struct {
	UserID   string `json:"user_id,omitempty"`
	TenantID string `json:"tenant_id,omitempty"`
}

// UserMeshes represents the user meshes.
type UserMeshes struct {
	User    *UserBody        `json:"user"`
	Profile *UserProfileBody `json:"profile,omitempty"`
	Roles   []*ReadRole      `json:"roles,omitempty"`
	Tenants []*ReadTenant    `json:"tenants,omitempty"`
	Groups  []*ReadGroup     `json:"groups,omitempty"`
}
