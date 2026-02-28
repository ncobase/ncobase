// Package space provides workspace and multi-tenant space management for the Ncobase platform.
//
// This package implements a comprehensive space management system that enables multi-tenancy,
// workspace isolation, and resource organization. Spaces are logical containers that group
// users, resources, and configurations together.
//
// Key Features:
//   - Space CRUD operations and lifecycle management
//   - Multi-tenant workspace isolation
//   - Space-level settings and configuration
//   - Billing and quota management per space
//   - Space member management and roles
//   - Custom menus and navigation per space
//   - Space templates and initialization
//
// Main Components:
//   - Handler: HTTP endpoints for space operations
//   - Service: Business logic for space management
//   - Repository: Data access layer for space entities
//
// Space Capabilities:
//   - Billing: Track usage and manage subscriptions
//   - Quotas: Enforce resource limits per space
//   - Settings: Space-specific configuration
//   - Menus: Custom navigation and UI per space
//   - Members: User access and role management
//
// Usage Example:
//
//	// Create a new space
//	space, err := spaceService.Create(ctx, &structs.CreateSpaceInput{
//	    Name:        "My Workspace",
//	    Description: "Team collaboration space",
//	    Type:        "team",
//	})
//	if err != nil {
//	    return err
//	}
//
//	// Set space quota
//	err = spaceService.SetQuota(ctx, space.ID, &structs.QuotaInput{
//	    Storage:  10737418240, // 10GB
//	    Users:    50,
//	    Projects: 100,
//	})
//
// The package provides complete multi-tenancy support with resource isolation,
// billing integration, and flexible configuration options.
package space
