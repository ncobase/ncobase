package structs

// MFALoginVerifyBody verifies an MFA login challenge.
type MFALoginVerifyBody struct {
	MFAToken     string `json:"mfa_token" validate:"required"`
	Code         string `json:"code,omitempty"`
	RecoveryCode string `json:"recovery_code,omitempty"`
}

// TwoFactorSetupBody starts 2FA setup.
type TwoFactorSetupBody struct {
	Method string `json:"method" validate:"required"` // "app"
	Phone  string `json:"phone,omitempty"`
}

// TwoFactorSetupResponse returns TOTP provisioning data.
type TwoFactorSetupResponse struct {
	Method     string `json:"method"`
	Secret     string `json:"secret"`
	OTPAuthURI string `json:"otpauth_uri"`
	QRPNG      string `json:"qr_png"` // base64 encoded PNG
}

// TwoFactorVerifyBody enables 2FA after validating the OTP code.
type TwoFactorVerifyBody struct {
	Code   string `json:"code" validate:"required"`
	Method string `json:"method" validate:"required"` // "app"
}

// TwoFactorDisableBody disables 2FA.
type TwoFactorDisableBody struct {
	Password     string `json:"password" validate:"required"`
	Code         string `json:"code,omitempty"`
	RecoveryCode string `json:"recovery_code,omitempty"`
}

// RecoveryCodesResponse returns newly generated recovery codes.
type RecoveryCodesResponse struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

// TwoFactorStatusResponse describes current 2FA state.
type TwoFactorStatusResponse struct {
	Enabled                bool   `json:"enabled"`
	Method                 string `json:"method,omitempty"`
	RecoveryCodesRemaining int    `json:"recovery_codes_remaining,omitempty"`
}
