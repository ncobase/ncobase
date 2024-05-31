package structs

import "time"

// CreateUserBody Create user body
type CreateUserBody struct {
	RegisterToken string `json:"register_token" binding:"required"`
	DisplayName   string `json:"display_name" binding:"required"`
	Username      string `json:"username" binding:"required"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	ShortBio      string `json:"short_bio"`
}

// CreateProfileBody Create user profile body
type CreateProfileBody struct {
	UserID       string    `json:"user_id" binding:"required"`
	DisplayName  string    `json:"display_name"`
	ShortBio     string    `json:"short_bio"`
	About        *string   `json:"about,omitempty"`
	Thumbnail    *string   `json:"thumbnail,omitempty"`
	ProfileLinks *[]string `json:"profile_links,omitempty"`
}

// CreateUserMetaBody Create user meta body
type CreateUserMetaBody struct {
	UserID            string `json:"user_id" binding:"required"`
	EmailNotification bool   `json:"email_notification,omitempty"`
	EmailPromotions   bool   `json:"email_promotions,omitempty"`
	PhoneNotification bool   `json:"phone_notification,omitempty"`
	PhonePromotions   bool   `json:"phone_promotions,omitempty"`
}

// UserKey Find user param
type UserKey struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

// ReadUser Output user schema
type ReadUser struct {
	ID          string             `json:"id"`
	Username    string             `json:"username"`
	Email       string             `json:"email"`
	Phone       string             `json:"phone"`
	IsCertified bool               `json:"is_certified"`
	IsAdmin     bool               `json:"is_admin"`
	Profile     *UserProfileSchema `json:"profile"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// UserProfileSchema Output user profile schema
type UserProfileSchema struct {
	ID           string    `json:"id,omitempty"`
	DisplayName  string    `json:"display_name"`
	ShortBio     string    `json:"short_bio"`
	About        *string   `json:"about,omitempty"`
	Thumbnail    *string   `json:"thumbnail,omitempty"`
	ProfileLinks *[]string `json:"profile_links,omitempty"`
}

// ChangeUserPasswordBody Change user password
type ChangeUserPasswordBody struct {
	UserID      string `json:"user_id" binding:"required"`
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// // UserMetaSchema Output user meta schema
// type UserMetaSchema struct {
// 	ID                string `json:"id,omitempty"`
// 	EmailNotification bool   `json:"email_notification"`
// 	EmailPromotions   bool   `json:"email_promotions"`
// 	PhoneNotification bool   `json:"phone_notification"`
// 	PhonePromotions   bool   `json:"phone_promotions"`
// }
