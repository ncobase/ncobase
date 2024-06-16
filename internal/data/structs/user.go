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

// User represents the user schema.
type User struct {
	ID          string       `json:"id"`
	Username    string       `json:"username"`
	Email       string       `json:"email,omitempty"`
	Phone       string       `json:"phone,omitempty"`
	IsCertified bool         `json:"is_certified"`
	IsAdmin     bool         `json:"is_admin"`
	Profile     *UserProfile `json:"profile,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// UserRequestBody is a unified structure for user-related request bodies.
type UserRequestBody struct {
	RegisterToken string    `json:"register_token,omitempty" validate:"required_if=Action create"`
	DisplayName   string    `json:"display_name,omitempty"`
	Username      string    `json:"username,omitempty"`
	Email         string    `json:"email,omitempty"`
	Phone         string    `json:"phone,omitempty"`
	ShortBio      string    `json:"short_bio,omitempty"`
	About         *string   `json:"about,omitempty"`
	Thumbnail     *string   `json:"thumbnail,omitempty"`
	ProfileLinks  *[]string `json:"profile_links,omitempty"`
	UserID        string    `json:"user_id,omitempty" validate:"required_if=Action profile"`
	OldPassword   string    `json:"old_password,omitempty" validate:"required_if=Action password"`
	NewPassword   string    `json:"new_password,omitempty" validate:"required_if=Action password"`
	Action        string    `json:"-"`
}

// UserProfile represents the user profile schema.
type UserProfile struct {
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
