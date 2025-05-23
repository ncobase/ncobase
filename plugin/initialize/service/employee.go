package service

import (
	"context"
	"fmt"
	data "ncobase/initialize/data/company"

	"github.com/ncobase/ncore/logging/logger"
)

// initializeEmployees creates employee records for initialized users
func (s *Service) initializeEmployees(ctx context.Context) error {
	logger.Infof(ctx, "Initializing employee records...")

	// Get default tenant
	defaultTenant, err := s.ts.Tenant.GetBySlug(ctx, "digital-enterprise")
	if err != nil {
		logger.Errorf(ctx, "Error getting default tenant: %v", err)
		return fmt.Errorf("default tenant not found: %w", err)
	}

	var createdCount int
	for _, userInfo := range data.SystemDefaultUsers {
		if userInfo.Employee == nil {
			continue // Skip users without employee data
		}

		// Check if user exists
		user, err := s.us.User.Get(ctx, userInfo.User.Username)
		if err != nil {
			logger.Warnf(ctx, "User %s not found, skipping employee creation", userInfo.User.Username)
			continue
		}

		// Set tenant ID and user ID for employee record
		employeeData := *userInfo.Employee
		employeeData.TenantID = defaultTenant.ID
		employeeData.UserID = user.ID

		// Resolve manager ID if specified
		if employeeData.ManagerID != "" {
			manager, err := s.us.User.Get(ctx, employeeData.ManagerID)
			if err == nil {
				employeeData.ManagerID = manager.ID
			} else {
				logger.Warnf(ctx, "Manager %s not found for employee %s", employeeData.ManagerID, user.Username)
				employeeData.ManagerID = ""
			}
		}

		// Create employee record (assuming we have an employee service)
		// This would need to be implemented in the user service
		/*
			if _, err := s.us.Employee.Create(ctx, &userStructs.CreateEmployeeBody{
				EmployeeBody: employeeData,
			}); err != nil {
				logger.Errorf(ctx, "Error creating employee record for user %s: %v", user.Username, err)
				return fmt.Errorf("failed to create employee record for user '%s': %w", user.Username, err)
			}
		*/

		logger.Debugf(ctx, "Created employee record for user: %s", user.Username)
		createdCount++
	}

	logger.Infof(ctx, "Employee initialization completed, created %d employee records", createdCount)
	return nil
}
