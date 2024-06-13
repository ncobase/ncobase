package structs

import "stocms/pkg/oauth"

// SendCodeBody Send verify code body
type SendCodeBody struct {
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
}

// CodeParams Verify code param
type CodeParams struct {
	Code string `json:"code" validate:"required"`
}

// CommonRegisterBody Common fields for register body
type CommonRegisterBody struct {
	DisplayName string `json:"display_name" validate:"required"`
	Username    string `json:"username" validate:"required"`
	Phone       string `json:"phone,omitempty"`
	ShortBio    string `json:"short_bio,omitempty"`
	Domain      string `json:"domain,omitempty"`
}

// RegisterBody Register body
type RegisterBody struct {
	CommonRegisterBody
	RegisterToken string `json:"register_token" validate:"required"`
}

// OAuthRegisterBody OAuth register body
type OAuthRegisterBody struct {
	CommonRegisterBody
	RegisterToken string `json:"register_token,omitempty"`
}

// RegisterTokenBody Register token body
type RegisterTokenBody struct {
	Profile  oauth.Profile `json:"profile"`
	Token    string        `json:"authorize"`
	Provider string        `json:"provider"`
}

// LoginBody Login body
type LoginBody struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}
