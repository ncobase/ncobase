package service

import (
	"context"
	"fmt"
	accessStructs "ncobase/core/access/structs"
	"ncobase/core/system/structs"
	userStructs "ncobase/core/user/structs"
)

// HealthcheckItem represents a specific system component check
type HealthcheckItem struct {
	Component   string `json:"component"`
	Status      string `json:"status"`
	Description string `json:"description,omitempty"`
	Error       string `json:"error,omitempty"`
}

// HealthcheckResult holds the combined health check results
type HealthcheckResult struct {
	Status  string            `json:"status"`
	Items   []HealthcheckItem `json:"items"`
	Summary string            `json:"summary"`
}

// PerformHealthcheck checks the health of the initialized system
func (s *Service) PerformHealthcheck(ctx context.Context) (*HealthcheckResult, error) {
	if !s.IsInitialized(ctx) {
		return &HealthcheckResult{
			Status:  "not_initialized",
			Summary: "System is not initialized",
		}, nil
	}

	result := &HealthcheckResult{
		Status: "healthy",
		Items:  make([]HealthcheckItem, 0),
	}

	// Check spaces
	space, err := s.ts.Space.GetBySlug(ctx, "digital-enterprise")
	if err != nil || space == nil {
		result.Items = append(result.Items, HealthcheckItem{
			Component:   "spaces",
			Status:      "error",
			Description: "Default space check",
			Error:       fmt.Sprintf("Default space not found: %v", err),
		})
		result.Status = "error"
	} else {
		result.Items = append(result.Items, HealthcheckItem{
			Component:   "spaces",
			Status:      "healthy",
			Description: "Default space exists",
		})
	}

	// Check users
	userParams := userStructs.ListUserParams{}
	userCount := s.us.User.CountX(ctx, &userParams)
	if userCount < 3 { // Expected minimum: super, admin, user
		result.Items = append(result.Items, HealthcheckItem{
			Component:   "users",
			Status:      "warning",
			Description: "User count check",
			Error:       fmt.Sprintf("Expected at least 3 users, found %d", userCount),
		})
		if result.Status == "healthy" {
			result.Status = "warning"
		}
	} else {
		result.Items = append(result.Items, HealthcheckItem{
			Component:   "users",
			Status:      "healthy",
			Description: "User count check",
		})
	}

	// Check roles
	roleParams := accessStructs.ListRoleParams{}
	roleCount := s.acs.Role.CountX(ctx, &roleParams)
	if roleCount < 3 { // Expected minimum: super-admin, admin, user
		result.Items = append(result.Items, HealthcheckItem{
			Component:   "roles",
			Status:      "warning",
			Description: "Role count check",
			Error:       fmt.Sprintf("Expected at least 3 roles, found %d", roleCount),
		})
		if result.Status == "healthy" {
			result.Status = "warning"
		}
	} else {
		result.Items = append(result.Items, HealthcheckItem{
			Component:   "roles",
			Status:      "healthy",
			Description: "Role count check",
		})
	}

	// Check menus
	menuParams := structs.ListMenuParams{}
	menuCount := s.sys.Menu.CountX(ctx, &menuParams)
	if menuCount < 10 { // Arbitrary minimum based on expected menu structure
		result.Items = append(result.Items, HealthcheckItem{
			Component:   "menus",
			Status:      "warning",
			Description: "Menu count check",
			Error:       fmt.Sprintf("Expected at least 10 menus, found %d", menuCount),
		})
		if result.Status == "healthy" {
			result.Status = "warning"
		}
	} else {
		result.Items = append(result.Items, HealthcheckItem{
			Component:   "menus",
			Status:      "healthy",
			Description: "Menu count check",
		})
	}

	// Generate summary
	errorCount := 0
	warningCount := 0
	for _, item := range result.Items {
		if item.Status == "error" {
			errorCount++
		} else if item.Status == "warning" {
			warningCount++
		}
	}

	if errorCount > 0 {
		result.Summary = fmt.Sprintf("System health check found %d errors and %d warnings",
			errorCount, warningCount)
	} else if warningCount > 0 {
		result.Summary = fmt.Sprintf("System health check found %d warnings", warningCount)
	} else {
		result.Summary = "System is healthy and properly initialized"
	}

	return result, nil
}
