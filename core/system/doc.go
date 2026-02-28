// Package system provides system-wide configuration and management for the Ncobase platform.
//
// This package handles system-level settings, dictionaries, menus, and administrative
// functions that apply across the entire platform. It provides the foundation for
// configurable and customizable system behavior.
//
// Key Features:
//   - System settings and configuration management
//   - Data dictionaries for key-value storage
//   - Menu and navigation management
//   - System options and preferences
//   - Administrative tools and utilities
//   - System health and status monitoring
//
// Main Components:
//   - Handler: HTTP endpoints for system operations
//   - Service: Business logic for system management
//   - Repository: Data access layer for system entities
//
// System Entities:
//   - Dictionaries: Key-value pairs for system-wide data
//   - Menus: Navigation structure and routing
//   - Options: System preferences and settings
//   - Settings: Configuration parameters
//
// Usage Example:
//
//	// Get system setting
//	setting, err := systemService.GetSetting(ctx, "app.name")
//	if err != nil {
//	    return err
//	}
//
//	// Create dictionary entry
//	dict, err := systemService.CreateDictionary(ctx, &structs.CreateDictionaryInput{
//	    Key:         "user.status",
//	    Value:       "active",
//	    Description: "User status types",
//	})
//
//	// Update menu structure
//	menu, err := systemService.UpdateMenu(ctx, menuID, &structs.UpdateMenuInput{
//	    Name:  "Dashboard",
//	    Path:  "/dashboard",
//	    Icon:  "dashboard",
//	    Order: 1,
//	})
//
// The package provides centralized system configuration management with support
// for dynamic settings, hierarchical menus, and extensible dictionaries.
package system
