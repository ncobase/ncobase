package middleware

import "ncobase/common/config"

var signingKey string

// Init initializes the middleware with the given signing key.
func Init(conf *config.Config) {
	signingKey = conf.Auth.JWT.Secret
}

// GetSigningKey returns the signing key.
func GetSigningKey() string {
	return signingKey
}
