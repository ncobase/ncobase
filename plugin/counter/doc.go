// Package counter provides auto-incrementing sequence generation for the Ncobase platform.
//
// This package implements a flexible counter system that generates unique sequential
// identifiers with customizable prefixes and formats. It's useful for generating
// order numbers, invoice IDs, ticket numbers, and other sequential identifiers.
//
// Key Features:
//   - Auto-incrementing counters with custom prefixes
//   - Multiple counter instances per entity type
//   - Thread-safe counter operations
//   - Customizable number formatting (padding, length)
//   - Counter reset and management
//   - Atomic increment operations
//
// Main Components:
//   - Handler: HTTP endpoints for counter operations
//   - Service: Business logic for counter management
//   - Repository: Data access layer for counter state
//
// Counter Format:
//   - Prefix: Custom string prefix (e.g., "INV", "ORD", "TKT")
//   - Separator: Optional separator between prefix and number
//   - Number: Auto-incrementing integer with padding
//   - Example: "INV-2024-00001", "ORD-00123", "TKT-456"
//
// Usage Example:
//
//	// Create a counter
//	counter, err := counterService.Create(ctx, &structs.CreateCounterInput{
//	    Name:      "invoice",
//	    Prefix:    "INV",
//	    Separator: "-",
//	    Padding:   5,
//	    StartFrom: 1,
//	})
//	if err != nil {
//	    return err
//	}
//
//	// Generate next number
//	nextNum, err := counterService.Next(ctx, "invoice")
//	if err != nil {
//	    return err
//	}
//	// Returns: "INV-00001"
//
//	// Get current value without incrementing
//	current, err := counterService.Current(ctx, "invoice")
//	if err != nil {
//	    return err
//	}
//
// The package provides reliable sequence generation suitable for business documents,
// tracking numbers, and any scenario requiring unique sequential identifiers.
package counter
