package middleware

import "ncobase/internal/config"

var signingKey string

// Init initializes the middleware with the given signing key.
func Init(conf *config.Config) {
	signingKey = conf.JWTSecret
}

// GetSigningKey returns the signing key.
func GetSigningKey() string {
	return signingKey
}
