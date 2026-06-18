package service

import (
	"context"
	"errors"
	"fmt"
	systemWrapper "ncobase/core/system/wrapper"
	"unicode"
)

// ValidatePasswordPolicy enforces the non-secret password requirements from system options.
func ValidatePasswordPolicy(ctx context.Context, options *systemWrapper.OptionServiceWrapper, password string) error {
	if password == "" {
		return errors.New("new password cannot be empty")
	}

	policy := systemWrapper.DefaultPasswordPolicy()
	if options != nil {
		policy = options.PasswordPolicy(ctx)
	}

	if len([]rune(password)) < policy.MinLength {
		return fmt.Errorf("new password must be at least %d characters long", policy.MinLength)
	}

	var hasUpper, hasLower, hasNumber, hasSymbol bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasNumber = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbol = true
		}
	}

	if policy.RequireUppercase && !hasUpper {
		return errors.New("new password must contain at least one uppercase letter")
	}
	if policy.RequireLowercase && !hasLower {
		return errors.New("new password must contain at least one lowercase letter")
	}
	if policy.RequireNumbers && !hasNumber {
		return errors.New("new password must contain at least one number")
	}
	if policy.RequireSymbols && !hasSymbol {
		return errors.New("new password must contain at least one symbol")
	}

	return nil
}
