// Package auth provides authentication and session management for the Ncobase platform.
//
// This package handles user authentication through multiple methods including traditional
// username/password, OAuth providers, and multi-factor authentication (MFA). It manages
// user sessions, tokens, and authentication state.
//
// Key Features:
//   - Username/password authentication
//   - OAuth 2.0 integration (Google, GitHub, etc.)
//   - Multi-Factor Authentication (MFA) with TOTP
//   - Session management and token generation
//   - Captcha verification for security
//   - Password reset and recovery
//   - Login/logout tracking and audit
//
// Main Components:
//   - Handler: HTTP endpoints for authentication operations
//   - Service: Business logic for authentication flows
//   - Repository: Data access for auth-related entities
//
// Supported Authentication Methods:
//   - Local: Username and password
//   - OAuth: Third-party provider authentication
//   - MFA: Time-based One-Time Password (TOTP)
//   - API Key: Programmatic access authentication
//
// Usage Example:
//
//	// Authenticate user
//	session, err := authService.Login(ctx, &structs.LoginInput{
//	    Username: "user@example.com",
//	    Password: "securepassword",
//	})
//	if err != nil {
//	    return err
//	}
//
// Security Features:
//   - Password hashing with bcrypt
//   - Session token encryption
//   - Rate limiting on authentication attempts
//   - Captcha protection against bots
//   - MFA for enhanced security
package auth
