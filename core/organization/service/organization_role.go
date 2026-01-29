package service

import (
	"context"
	"ncobase/core/organization/data"
	"ncobase/core/organization/data/repository"
	"ncobase/core/organization/structs"
)

// OrganizationRoleServiceInterface is the interface for the service.
type OrganizationRoleServiceInterface interface {
	AddRoleToOrganization(ctx context.Context, organizationID string, roleID string) (*structs.OrganizationRole, error)
	RemoveRoleFromOrganization(ctx context.Context, organizationID string, roleID string) error
	GetOrganizationRolesIds(ctx context.Context, organizationID string) ([]string, error)
}

// organizationRoleService is the struct for the service.
type organizationRoleService struct {
	r repository.OrganizationRoleRepositoryInterface
}

// NewOrganizationRoleService creates a new service.
func NewOrganizationRoleService(d *data.Data) OrganizationRoleServiceInterface {
	return &organizationRoleService{
		r: repository.NewOrganizationRoleRepository(d),
	}
}

// AddRoleToOrganization adds a role to an organization.
func (s *organizationRoleService) AddRoleToOrganization(ctx context.Context, organizationID string, roleID string) (*structs.OrganizationRole, error) {
	row, err := s.r.Create(ctx, &structs.OrganizationRole{OrgID: organizationID, RoleID: roleID})
	if err := handleEntError(ctx, "OrganizationRole", err); err != nil {
		return nil, err
	}

	return repository.SerializeOrganizationRole(row), nil
}

// RemoveRoleFromOrganization removes a role from an organization.
func (s *organizationRoleService) RemoveRoleFromOrganization(ctx context.Context, organizationID string, roleID string) error {
	err := s.r.Delete(ctx, organizationID, roleID)
	if err := handleEntError(ctx, "OrganizationRole", err); err != nil {
		return err
	}

	return nil
}

// GetOrganizationRolesIds retrieves all roles under an organization.
func (s *organizationRoleService) GetOrganizationRolesIds(ctx context.Context, organizationID string) ([]string, error) {
	roleIDs, err := s.r.GetRolesByOrgID(ctx, organizationID)
	if err := handleEntError(ctx, "OrganizationRole", err); err != nil {
		return nil, err
	}
	return roleIDs, nil
}
