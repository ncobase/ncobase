// Package initialize provides system initialization and setup for the Ncobase platform.
//
// This package handles the initial setup and configuration of the Ncobase system,
// including database schema creation, default data seeding, and multi-mode initialization
// for different deployment scenarios (website, company, enterprise).
//
// Key Features:
//   - Multi-mode initialization (website, company, enterprise)
//   - Database schema creation and migration
//   - Default data seeding
//   - Admin user creation
//   - System configuration setup
//   - Initialization status tracking
//   - Idempotent initialization operations
//
// Main Components:
//   - Handler: HTTP endpoints for initialization operations
//   - Service: Business logic for system setup
//   - Repository: Data access for initialization state
//
// Initialization Modes:
//   - Website: Basic setup for public websites
//   - Company: Setup for company/organization use
//   - Enterprise: Full enterprise setup with advanced features
//
// Initialization Steps:
//   1. Database schema creation
//   2. Default roles and permissions
//   3. System settings and configuration
//   4. Admin user creation
//   5. Sample data (optional)
//   6. Plugin initialization
//
// Usage Example:
//
//	// Check initialization status
//	status, err := initService.GetStatus(ctx)
//	if err != nil {
//	    return err
//	}
//	if status.Initialized {
//	    return errors.New("system already initialized")
//	}
//
//	// Initialize system
//	result, err := initService.Initialize(ctx, &structs.InitializeInput{
//	    Mode:          "enterprise",
//	    AdminEmail:    "admin@example.com",
//	    AdminPassword: "securepassword",
//	    CompanyName:   "Acme Corp",
//	    SeedData:      true,
//	})
//	if err != nil {
//	    return err
//	}
//
// The package ensures safe and repeatable system initialization with support for
// different deployment scenarios and configuration options.
package initialize
