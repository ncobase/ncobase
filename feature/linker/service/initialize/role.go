package initialize

import (
	"context"
	"ncobase/common/log"
	accessStructs "ncobase/feature/access/structs"
)

// checkRolesInitialized checks if roles are already initialized.
func (s *InitializeService) checkRolesInitialized(ctx context.Context) error {
	params := &accessStructs.ListRoleParams{}
	count := s.acs.Role.CountX(ctx, params)
	if count == 0 {
		return s.initRoles(ctx)
	}

	return nil
}

// initRoles initializes roles.
func (s *InitializeService) initRoles(ctx context.Context) error {
	roles := []*accessStructs.CreateRoleBody{
		{
			RoleBody: accessStructs.RoleBody{
				Name:        "Super Admin",
				Slug:        "super-admin",
				Disabled:    false,
				Description: "Super Administrator role with all permissions",
				Extras:      nil,
			},
		},
		{
			RoleBody: accessStructs.RoleBody{
				Name:        "Admin",
				Slug:        "admin",
				Disabled:    false,
				Description: "Administrator role with some permissions",
				Extras:      nil,
			},
		},
		{
			RoleBody: accessStructs.RoleBody{
				Name:        "User",
				Slug:        "user",
				Disabled:    false,
				Description: "User role with some permissions",
				Extras:      nil,
			},
		},
	}

	for _, role := range roles {
		if _, err := s.acs.Role.Create(ctx, role); err != nil {
			log.Errorf(ctx, "initRoles error on create role: %v\n", err)
			return err
		}
	}

	count := s.acs.Role.CountX(ctx, &accessStructs.ListRoleParams{})
	log.Infof(ctx, "-------- initRoles done, created %d roles\n", count)

	return nil
}
