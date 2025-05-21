package service

import (
	"context"
	"errors"
	accessStructs "ncobase/access/structs"
	"ncobase/initialize/data"

	"github.com/ncobase/ncore/logging/logger"
)

// checkRolesInitialized checks if roles are already initialized.
func (s *Service) checkRolesInitialized(ctx context.Context) error {
	params := &accessStructs.ListRoleParams{}
	count := s.acs.Role.CountX(ctx, params)
	if count > 0 {
		logger.Infof(ctx, "Roles already exist, skipping initialization")
		return nil
	}

	return s.initRoles(ctx)
}

// initRoles initializes roles.
func (s *Service) initRoles(ctx context.Context) error {
	logger.Infof(ctx, "Initializing system roles...")

	// Initialize system roles
	for _, role := range data.SystemDefaultRoles {
		// Check if role already exists
		existingRole, err := s.acs.Role.GetBySlug(ctx, role.Slug)
		if err == nil && existingRole != nil {
			logger.Infof(ctx, "Role %s already exists, skipping", role.Slug)
			continue
		}

		_, err = s.acs.Role.Create(ctx, &role)
		if err != nil {
			// Check if error is due to unique constraint violation
			if errors.Is(err, errors.New("slug_key")) || errors.Is(err, errors.New("duplicate key")) {
				logger.Warnf(ctx, "Role %s already exists (caught duplicate key), skipping", role.Name)
				continue
			}

			logger.Warnf(ctx, "Error creating role %s: %v", role.Name, err)
			continue // Skip this role but continue with others
		}
		logger.Debugf(ctx, "Created role: %s", role.Name)
	}

	// Initialize organization-specific roles
	for _, role := range data.OrganizationRoles {
		// Check if role already exists
		existingRole, err := s.acs.Role.GetBySlug(ctx, role.Role.Slug)
		if err == nil && existingRole != nil {
			logger.Infof(ctx, "Organization role %s already exists, skipping", role.Role.Slug)
			continue
		}

		_, err = s.acs.Role.Create(ctx, &accessStructs.CreateRoleBody{
			RoleBody: role.Role,
		})
		if err != nil {
			// Check if error is due to unique constraint violation
			if errors.Is(err, errors.New("slug_key")) || errors.Is(err, errors.New("duplicate key")) {
				logger.Warnf(ctx, "Organization role %s already exists (caught duplicate key), skipping", role.Role.Name)
				continue
			}

			logger.Warnf(ctx, "Error creating organization role %s: %v", role.Role.Name, err)
			continue // Skip this role but continue with others
		}
		logger.Debugf(ctx, "Created organization role: %s", role.Role.Name)
	}

	count := s.acs.Role.CountX(ctx, &accessStructs.ListRoleParams{})
	logger.Infof(ctx, "Role initialization completed, %d roles now in system", count)

	return nil
}
