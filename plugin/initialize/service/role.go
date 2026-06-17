package service

import (
	"context"
	"fmt"
	accessStructs "ncobase/core/access/structs"
	"strings"

	"github.com/ncobase/ncore/logging/logger"
)

// checkRolesInitialized checks if roles are already initialized
func (s *Service) checkRolesInitialized(ctx context.Context) error {
	params := &accessStructs.ListRoleParams{}
	count := s.acs.Role.CountX(ctx, params)
	if count > 0 {
		logger.Infof(ctx, "Roles already exist, synchronizing missing default roles")
		return s.syncDefaultRoles(ctx)
	}

	return s.initRoles(ctx)
}

// initRoles initializes roles using current data mode
func (s *Service) initRoles(ctx context.Context) error {
	logger.Infof(ctx, "Initializing system roles in %s mode...", s.state.DataMode)

	if err := s.syncDefaultRoles(ctx); err != nil {
		return err
	}

	logger.Infof(ctx, "All essential roles validated successfully for %s mode", s.state.DataMode)
	return nil
}

// syncDefaultRoles creates missing default roles without overwriting operator customizations.
func (s *Service) syncDefaultRoles(ctx context.Context) error {
	dataLoader := s.getDataLoader()
	roles := dataLoader.GetRoles()

	var createdCount int
	for _, role := range roles {
		existingRole, err := s.acs.Role.GetBySlug(ctx, role.Slug)
		if err == nil && existingRole != nil {
			logger.Debugf(ctx, "Role %s already exists, skipping", role.Slug)
			continue
		}

		_, err = s.acs.Role.Create(ctx, &role)
		if err != nil {
			if isDuplicateRoleSeedError(err) {
				logger.Warnf(ctx, "Role %s already exists (caught duplicate key), skipping", role.Name)
				continue
			}

			logger.Errorf(ctx, "Error creating role %s: %v", role.Name, err)
			return fmt.Errorf("failed to create role '%s': %w", role.Name, err)
		}
		logger.Debugf(ctx, "Created role: %s", role.Name)
		createdCount++
	}

	count := s.acs.Role.CountX(ctx, &accessStructs.ListRoleParams{})
	logger.Infof(ctx, "Role synchronization completed, created %d missing roles, %d roles now in system", createdCount, count)

	// validate essential roles
	if err := s.validateEssentialRoles(ctx); err != nil {
		return err
	}
	return nil
}

func isDuplicateRoleSeedError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "slug_key") ||
		strings.Contains(message, "duplicate key") ||
		strings.Contains(message, "constraint failed")
}

// getEssentialRoles returns essential roles based on current data mode
func (s *Service) getEssentialRoles() []string {
	switch s.state.DataMode {
	case "website":
		return []string{"super-admin", "admin"}
	case "company":
		return []string{"super-admin", "system-admin", "company-admin"}
	case "enterprise":
		return []string{"super-admin", "system-admin", "enterprise-admin"}
	default:
		return []string{"super-admin", "admin"}
	}
}

// validateEssentialRoles validates that essential roles exist for the current mode
func (s *Service) validateEssentialRoles(ctx context.Context) error {
	essential := s.getEssentialRoles()

	for _, slug := range essential {
		role, err := s.acs.Role.GetBySlug(ctx, slug)
		if err != nil || role == nil {
			logger.Errorf(ctx, "Essential role '%s' was not created for %s mode", slug, s.state.DataMode)
			return fmt.Errorf("essential role '%s' was not created during initialization", slug)
		}
		logger.Debugf(ctx, "Essential role '%s' validated successfully", slug)
	}

	return nil
}
