package service

import (
	"context"
	"ncobase/organization/data"
	"ncobase/organization/data/ent"
	"ncobase/organization/data/repository"
	"ncobase/organization/structs"
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

	return s.Serialize(row), nil
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

// Serializes serializes the organization roles.
func (s *organizationRoleService) Serializes(rows []*ent.OrganizationRole) []*structs.OrganizationRole {
	rs := make([]*structs.OrganizationRole, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize serializes the organization role.
func (s *organizationRoleService) Serialize(row *ent.OrganizationRole) *structs.OrganizationRole {
	return &structs.OrganizationRole{
		OrgID:  row.OrgID,
		RoleID: row.RoleID,
	}
}
