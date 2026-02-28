// Package user provides user management and profile functionality for the Ncobase platform.
//
// This package handles all user-related operations including user accounts, profiles,
// employee records, and API key management. It provides comprehensive user lifecycle
// management from registration to deletion.
//
// Key Features:
//   - User account management (CRUD operations)
//   - User profile and personal information
//   - Employee records and organizational data
//   - API key generation and management
//   - User preferences and settings
//   - User status and lifecycle management
//   - Password management and security
//
// Main Components:
//   - Handler: HTTP endpoints for user operations
//   - Service: Business logic for user management
//   - Repository: Data access layer for user entities
//
// User Entities:
//   - User: Core user account and authentication
//   - Profile: Personal information and preferences
//   - Employee: Organizational and employment data
//   - APIKey: Programmatic access credentials
//
// Usage Example:
//
//	// Create a new user
//	user, err := userService.Create(ctx, &structs.CreateUserInput{
//	    Username: "john.doe",
//	    Email:    "john@example.com",
//	    Password: "securepassword",
//	    Name:     "John Doe",
//	})
//	if err != nil {
//	    return err
//	}
//
//	// Update user profile
//	profile, err := userService.UpdateProfile(ctx, user.ID, &structs.UpdateProfileInput{
//	    Avatar:   "https://example.com/avatar.jpg",
//	    Bio:      "Software Engineer",
//	    Location: "San Francisco, CA",
//	})
//
//	// Generate API key
//	apiKey, err := userService.CreateAPIKey(ctx, user.ID, &structs.CreateAPIKeyInput{
//	    Name:        "Production API",
//	    Description: "API key for production access",
//	    ExpiresAt:   time.Now().AddDate(1, 0, 0), // 1 year
//	})
//
// The package provides complete user lifecycle management with support for
// profiles, organizational data, and secure API access.
package user
