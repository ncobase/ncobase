package initialize

import (
	"context"
	accessStructs "ncobase/access/structs"
	"ncobase/system/initialize/data"

	"github.com/ncobase/ncore/logging/logger"
)

// checkRolesInitialized checks if roles are already initialized.
func (s *Service) checkRolesInitialized(ctx context.Context) error {
	params := &accessStructs.ListRoleParams{}
	count := s.acs.Role.CountX(ctx, params)
	if count == 0 {
		return s.initRoles(ctx)
	}

	return nil
}

// initRoles initializes roles.
func (s *Service) initRoles(ctx context.Context) error {
	logger.Infof(ctx, "Initializing system roles...")

	for _, role := range data.SystemDefaultRoles {
		if _, err := s.acs.Role.Create(ctx, &role); err != nil {
			logger.Errorf(ctx, "Error creating role %s: %v", role.Name, err)
			return err
		}
		logger.Debugf(ctx, "Created role: %s", role.Name)
	}

	count := s.acs.Role.CountX(ctx, &accessStructs.ListRoleParams{})
	logger.Infof(ctx, "Role initialization completed, created %d roles", count)

	return nil
}
