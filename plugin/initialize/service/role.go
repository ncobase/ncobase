package service

import (
	"context"
	"errors"
	"fmt"
	accessStructs "ncobase/access/structs"

	"github.com/ncobase/ncore/logging/logger"
)

// checkRolesInitialized checks if roles are already initialized
func (s *Service) checkRolesInitialized(ctx context.Context) error {
	params := &accessStructs.ListRoleParams{}
	count := s.acs.Role.CountX(ctx, params)
	if count > 0 {
		logger.Infof(ctx, "Roles already exist, skipping initialization")
		return nil
	}

	return s.initRoles(ctx)
}

// initRoles initializes roles using current data mode
func (s *Service) initRoles(ctx context.Context) error {
	logger.Infof(ctx, "Initializing system roles in %s mode...", s.state.DataMode)

	dataLoader := s.getDataLoader()
	roles := dataLoader.GetRoles()

	for _, role := range roles {
		existingRole, err := s.acs.Role.GetBySlug(ctx, role.Slug)
		if err == nil && existingRole != nil {
			logger.Infof(ctx, "Role %s already exists, skipping", role.Slug)
			continue
		}

		_, err = s.acs.Role.Create(ctx, &role)
		if err != nil {
			if errors.Is(err, errors.New("slug_key")) || errors.Is(err, errors.New("duplicate key")) {
				logger.Warnf(ctx, "Role %s already exists (caught duplicate key), skipping", role.Name)
				continue
			}

			logger.Errorf(ctx, "Error creating role %s: %v", role.Name, err)
			return fmt.Errorf("failed to create role '%s': %w", role.Name, err)
		}
		logger.Debugf(ctx, "Created role: %s", role.Name)
	}

	count := s.acs.Role.CountX(ctx, &accessStructs.ListRoleParams{})
	logger.Infof(ctx, "Role initialization completed, %d roles now in system", count)

	// Validate essential roles exist based on mode
	essential := []string{"super-admin", "system-admin"}
	if s.state.DataMode == "enterprise" {
		essential = append(essential, "enterprise-admin")
	}

	for _, slug := range essential {
		role, err := s.acs.Role.GetBySlug(ctx, slug)
		if err != nil || role == nil {
			logger.Errorf(ctx, "Essential role '%s' was not created", slug)
			return fmt.Errorf("essential role '%s' was not created during initialization", slug)
		}
	}

	logger.Infof(ctx, "All essential roles validated successfully for %s mode", s.state.DataMode)
	return nil
}
