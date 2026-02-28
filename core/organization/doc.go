// Package organization provides organization and tenant management for the Ncobase platform.
//
// This package manages organizational structures, hierarchies, and relationships within
// the system. It supports multi-tenant architectures where each organization can have
// its own settings, users, and resources.
//
// Key Features:
//   - Organization CRUD operations
//   - Organizational hierarchy management
//   - Organization settings and configuration
//   - Member management within organizations
//   - Organization-level resource isolation
//
// Main Components:
//   - Handler: HTTP endpoints for organization operations
//   - Service: Business logic for organization management
//   - Repository: Data access layer for organization entities
//
// Organization Structure:
//   - Organizations can have parent-child relationships
//   - Each organization has its own settings and configuration
//   - Organizations can have multiple members with different roles
//   - Resources can be scoped to specific organizations
//
// Usage Example:
//
//	// Create a new organization
//	org, err := orgService.Create(ctx, &structs.CreateOrganizationInput{
//	    Name:        "Acme Corp",
//	    Description: "Main organization",
//	    Type:        "enterprise",
//	})
//	if err != nil {
//	    return err
//	}
//
// The package supports multi-tenancy patterns where each organization operates
// independently with its own data and configuration.
package organization
