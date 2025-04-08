package structs

import (
	accessStructs "ncobase/core/access/structs"
	spaceStructs "ncobase/core/space/structs"
	tenantStructs "ncobase/core/tenant/structs"
	userStructs "ncobase/core/user/structs"
	"ncore/pkg/oauth"
)

// SendCodeBody Send verify code body
type SendCodeBody struct {
	Email string `json:"email,omitempty" validate:"required_if=Phone empty"`
	Phone string `json:"phone,omitempty" validate:"required_if=Email empty"`
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
	Tenant      string `json:"tenant,omitempty"`
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

// Captcha contains the fields for captcha validation.
type Captcha struct {
	ID       string `json:"id" validate:"required"`
	Solution string `json:"solution" validate:"required"`
}

// LoginBody Login body
type LoginBody struct {
	Username string   `json:"username" validate:"required"`
	Password string   `json:"password" validate:"required"`
	Captcha  *Captcha `json:"captcha,omitempty"`
}

// AccountMeshes represents the account meshes.
type AccountMeshes struct {
	User    *userStructs.ReadUser        `json:"user,omitempty"`
	Profile *userStructs.ReadUserProfile `json:"profile,omitempty"`
	Roles   []*accessStructs.ReadRole    `json:"roles,omitempty"`
	Tenants []*tenantStructs.ReadTenant  `json:"tenants,omitempty"`
	Groups  []*spaceStructs.ReadGroup    `json:"groups,omitempty"`
}
