package wrapper

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	ext "github.com/ncobase/ncore/extension/types"
)

const (
	optionSystemFrontend = "system.frontend"
	optionSystemEmail    = "system.email_policy"
	optionSystemSecurity = "system.security"
	optionAuthToken      = "auth.token"
	optionAuthSession    = "auth.session"
)

// OptionObjectService defines the system option methods used by runtime config consumers.
type OptionObjectService interface {
	GetObjectByName(ctx context.Context, name string) (map[string]any, error)
}

// FrontendRuntimeOptions contains public frontend entry points used by email links.
type FrontendRuntimeOptions struct {
	SignInURL string
	SignUpURL string
}

// AuthTokenRuntimeOptions contains token expiry policy managed through system options.
type AuthTokenRuntimeOptions struct {
	AccessTokenExpiry   time.Duration
	RefreshTokenExpiry  time.Duration
	RegisterTokenExpiry time.Duration
	MFATokenExpiry      time.Duration
}

// AuthSessionRuntimeOptions contains session policy managed through system options.
type AuthSessionRuntimeOptions struct {
	MaxSessions     int
	SessionExpiry   time.Duration
	CleanupInterval time.Duration
}

// EmailRuntimePolicy contains non-secret email behavior policy managed through system options.
type EmailRuntimePolicy struct {
	Enabled            bool
	SenderName         string
	AllowAuthEmail     bool
	AllowPasswordReset bool
	DigestFrequency    string
}

// PasswordRuntimePolicy contains non-secret password requirements managed through system options.
type PasswordRuntimePolicy struct {
	MinLength        int  `json:"min_length"`
	RequireUppercase bool `json:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase"`
	RequireNumbers   bool `json:"require_numbers"`
	RequireSymbols   bool `json:"require_symbols"`
}

// OptionServiceWrapper wraps system option access for modules that cannot import system services directly.
type OptionServiceWrapper struct {
	em      ext.ManagerInterface
	options OptionObjectService
}

// NewOptionServiceWrapper creates a new option service wrapper.
func NewOptionServiceWrapper(em ext.ManagerInterface) *OptionServiceWrapper {
	wrapper := &OptionServiceWrapper{em: em}
	wrapper.loadServices()
	return wrapper
}

func (w *OptionServiceWrapper) loadServices() {
	if w == nil || w.em == nil {
		return
	}
	if optionSvc, err := w.em.GetCrossService("system", "Option"); err == nil {
		if service, ok := optionSvc.(OptionObjectService); ok {
			w.options = service
		}
	}
}

// RefreshServices refreshes the option service reference.
func (w *OptionServiceWrapper) RefreshServices() {
	w.loadServices()
}

// GetObjectByName returns an object option by name.
func (w *OptionServiceWrapper) GetObjectByName(ctx context.Context, name string) (map[string]any, error) {
	if w == nil {
		return nil, fmt.Errorf("system option wrapper is not configured")
	}
	if w.options == nil {
		w.loadServices()
	}
	if w.options == nil {
		return nil, fmt.Errorf("system option service is not available")
	}
	return w.options.GetObjectByName(ctx, name)
}

// Frontend returns frontend runtime options with code defaults for bootstrap safety.
func (w *OptionServiceWrapper) Frontend(ctx context.Context) FrontendRuntimeOptions {
	options := FrontendRuntimeOptions{
		SignInURL: "http://localhost:3000/login",
		SignUpURL: "http://localhost:3000/register",
	}

	values, err := w.GetObjectByName(ctx, optionSystemFrontend)
	if err != nil {
		return options
	}
	if value, ok := stringFromAny(values["sign_in_url"]); ok && value != "" {
		options.SignInURL = value
	}
	if value, ok := stringFromAny(values["sign_up_url"]); ok && value != "" {
		options.SignUpURL = value
	}
	return options
}

// AuthToken returns token runtime policy with code defaults for bootstrap safety.
func (w *OptionServiceWrapper) AuthToken(ctx context.Context) AuthTokenRuntimeOptions {
	options := AuthTokenRuntimeOptions{
		AccessTokenExpiry:   2 * time.Hour,
		RefreshTokenExpiry:  7 * 24 * time.Hour,
		RegisterTokenExpiry: 30 * time.Minute,
		MFATokenExpiry:      5 * time.Minute,
	}

	values, err := w.GetObjectByName(ctx, optionAuthToken)
	if err != nil {
		return options
	}
	if value, ok := durationFromAny(values["access_token_expiry"]); ok && value > 0 {
		options.AccessTokenExpiry = value
	}
	if value, ok := durationFromAny(values["refresh_token_expiry"]); ok && value > 0 {
		options.RefreshTokenExpiry = value
	}
	if value, ok := durationFromAny(values["register_token_expiry"]); ok && value > 0 {
		options.RegisterTokenExpiry = value
	}
	if value, ok := durationFromAny(values["mfa_token_expiry"]); ok && value > 0 {
		options.MFATokenExpiry = value
	}
	return options
}

// AuthSession returns session runtime policy with code defaults for bootstrap safety.
func (w *OptionServiceWrapper) AuthSession(ctx context.Context) AuthSessionRuntimeOptions {
	options := AuthSessionRuntimeOptions{
		MaxSessions:     0,
		SessionExpiry:   7 * 24 * time.Hour,
		CleanupInterval: time.Hour,
	}

	values, err := w.GetObjectByName(ctx, optionAuthSession)
	if err != nil {
		return options
	}
	if value, ok := intFromAny(values["max_sessions"]); ok && value >= 0 {
		options.MaxSessions = value
	}
	if value, ok := durationFromAny(values["session_expiry"]); ok && value > 0 {
		options.SessionExpiry = value
	}
	if value, ok := durationFromAny(values["cleanup_interval"]); ok && value > 0 {
		options.CleanupInterval = value
	}
	return options
}

// EmailPolicy returns non-secret email runtime policy with code defaults for bootstrap safety.
func (w *OptionServiceWrapper) EmailPolicy(ctx context.Context) EmailRuntimePolicy {
	options := EmailRuntimePolicy{
		Enabled:            true,
		SenderName:         "System Admin",
		AllowAuthEmail:     true,
		AllowPasswordReset: true,
		DigestFrequency:    "daily",
	}

	values, err := w.GetObjectByName(ctx, optionSystemEmail)
	if err != nil {
		return options
	}
	if value, ok := boolFromAny(values["enabled"]); ok {
		options.Enabled = value
	}
	if value, ok := stringFromAny(values["sender_name"]); ok && value != "" {
		options.SenderName = value
	}
	if value, ok := boolFromAny(values["allow_auth_email"]); ok {
		options.AllowAuthEmail = value
	}
	if value, ok := boolFromAny(values["allow_password_reset"]); ok {
		options.AllowPasswordReset = value
	}
	if value, ok := stringFromAny(values["digest_frequency"]); ok && value != "" {
		options.DigestFrequency = value
	}
	return options
}

// PasswordPolicy returns the public, non-secret password complexity policy.
func (w *OptionServiceWrapper) PasswordPolicy(ctx context.Context) PasswordRuntimePolicy {
	options := DefaultPasswordPolicy()

	values, err := w.GetObjectByName(ctx, optionSystemSecurity)
	if err != nil {
		return options
	}

	if value, ok := intFromAny(firstPresent(values, "passwordMinLength", "min_length")); ok && value > 0 {
		options.MinLength = value
	}

	if complexity, ok := boolFromAny(firstPresent(values, "passwordComplexity", "password_complexity")); ok {
		options.RequireUppercase = complexity
		options.RequireLowercase = complexity
		options.RequireNumbers = complexity
	}

	if value, ok := boolFromAny(firstPresent(values, "requireUppercase", "require_uppercase")); ok {
		options.RequireUppercase = value
	}
	if value, ok := boolFromAny(firstPresent(values, "requireLowercase", "require_lowercase")); ok {
		options.RequireLowercase = value
	}
	if value, ok := boolFromAny(firstPresent(values, "requireNumbers", "require_numbers", "requireDigits", "require_digits")); ok {
		options.RequireNumbers = value
	}
	if value, ok := boolFromAny(firstPresent(values, "requireSymbols", "require_symbols", "requireSpecial", "require_special")); ok {
		options.RequireSymbols = value
	}

	return options
}

// DefaultPasswordPolicy returns the code-level fallback used when system options are unavailable.
func DefaultPasswordPolicy() PasswordRuntimePolicy {
	return PasswordRuntimePolicy{
		MinLength:        8,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumbers:   true,
		RequireSymbols:   false,
	}
}

func firstPresent(values map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := values[key]; ok {
			return value
		}
	}
	return nil
}

func stringFromAny(value any) (string, bool) {
	v, ok := value.(string)
	return strings.TrimSpace(v), ok
}

func boolFromAny(value any) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(v))
		return parsed, err == nil
	default:
		return false, false
	}
}

func intFromAny(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(v))
		return parsed, err == nil
	default:
		return 0, false
	}
}

func durationFromAny(value any) (time.Duration, bool) {
	switch v := value.(type) {
	case string:
		duration, err := parseRuntimeDuration(v)
		return duration, err == nil
	case int:
		return time.Duration(v) * time.Second, true
	case int64:
		return time.Duration(v) * time.Second, true
	case float64:
		return time.Duration(v * float64(time.Second)), true
	default:
		return 0, false
	}
}

func parseRuntimeDuration(value string) (time.Duration, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("empty duration")
	}
	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil {
		return time.Duration(seconds) * time.Second, nil
	}
	if strings.HasSuffix(value, "d") {
		days, err := strconv.ParseInt(strings.TrimSuffix(value, "d"), 10, 64)
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	if strings.HasSuffix(value, "w") {
		weeks, err := strconv.ParseInt(strings.TrimSuffix(value, "w"), 10, 64)
		if err != nil {
			return 0, err
		}
		return time.Duration(weeks) * 7 * 24 * time.Hour, nil
	}
	return time.ParseDuration(value)
}
