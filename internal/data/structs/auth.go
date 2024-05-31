package structs

import (
	"stocms/pkg/oauth"
)

// SendCodeBody Send verify code body
type SendCodeBody struct {
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
}

// CodeParams Verify code param
type CodeParams struct {
	Code string `json:"code" binging:"required"`
}

// RegisterBody Register body
type RegisterBody struct {
	RegisterToken string `json:"register_token" binding:"required"`
	DisplayName   string `json:"display_name" binding:"required"`
	Username      string `json:"username" binding:"required"`
	Phone         string `json:"phone"`
	ShortBio      string `json:"short_bio"`
}

// OAuthRegisterBody OAuth register body
type OAuthRegisterBody struct {
	RegisterToken string `json:"register_token"`
	DisplayName   string `json:"display_name" binding:"required"`
	Username      string `json:"username" binding:"required"`
	ShortBio      string `json:"short_bio"`
	Phone         string `json:"phone"`
}

// RegisterTokenBody Register token body
type RegisterTokenBody struct {
	Profile  oauth.Profile `json:"profile"`
	Token    string        `json:"authorize"`
	Provider string        `json:"provider"`
}

// LoginBody Login body
type LoginBody struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
