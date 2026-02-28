// Package access provides access control and authorization functionality for the Ncobase platform.
//
// This package implements Role-Based Access Control (RBAC) using Casbin as the underlying
// policy engine. It manages permissions, roles, and access policies across the system.
//
// Key Features:
//   - Role-Based Access Control (RBAC) with Casbin integration
//   - Permission management and assignment
//   - Role hierarchy and inheritance
//   - Policy enforcement and validation
//   - Dynamic permission checking
//
// Main Components:
//   - Handler: HTTP endpoints for access control operations
//   - Service: Business logic for permission and role management
//   - Repository: Data access layer for access control entities
//
// Usage Example:
//
//	// Check if user has permission
//	hasAccess, err := accessService.CheckPermission(ctx, userID, resource, action)
//	if err != nil {
//	    return err
//	}
//	if !hasAccess {
//	    return errors.New("access denied")
//	}
//
// The package integrates with the Casbin library to provide flexible and powerful
// access control mechanisms suitable for enterprise applications.
package access
