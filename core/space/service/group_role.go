package service

import (
	"context"
	"ncobase/space/data"
	"ncobase/space/data/ent"
	"ncobase/space/data/repository"
	"ncobase/space/structs"
)

// GroupRoleServiceInterface is the interface for the service.
type GroupRoleServiceInterface interface {
	AddRoleToGroup(ctx context.Context, groupID string, roleID string) (*structs.GroupRole, error)
	RemoveRoleFromGroup(ctx context.Context, groupID string, roleID string) error
	GetGroupRolesIds(ctx context.Context, groupID string) ([]string, error)
}

// groupRoleService is the struct for the service.
type groupRoleService struct {
	r repository.GroupRoleRepositoryInterface
}

// NewGroupRoleService creates a new service.
func NewGroupRoleService(d *data.Data) GroupRoleServiceInterface {
	return &groupRoleService{
		r: repository.NewGroupRoleRepository(d),
	}
}

// AddRoleToGroup adds a role to a group.
func (s *groupRoleService) AddRoleToGroup(ctx context.Context, groupID string, roleID string) (*structs.GroupRole, error) {
	row, err := s.r.Create(ctx, &structs.GroupRole{GroupID: groupID, RoleID: roleID})
	if err := handleEntError(ctx, "GroupRole", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// RemoveRoleFromGroup removes a role from a group.
func (s *groupRoleService) RemoveRoleFromGroup(ctx context.Context, groupID string, roleID string) error {
	err := s.r.Delete(ctx, groupID, roleID)
	if err := handleEntError(ctx, "GroupRole", err); err != nil {
		return err
	}

	return nil
}

// GetGroupRolesIds retrieves all roles under a group.
func (s *groupRoleService) GetGroupRolesIds(ctx context.Context, groupID string) ([]string, error) {
	roleIDs, err := s.r.GetRolesByGroupID(ctx, groupID)
	if err := handleEntError(ctx, "GroupRole", err); err != nil {
		return nil, err
	}
	return roleIDs, nil
}

// Serializes serializes the group roles.
func (s *groupRoleService) Serializes(rows []*ent.GroupRole) []*structs.GroupRole {
	var rs []*structs.GroupRole
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize serializes the group role.
func (s *groupRoleService) Serialize(row *ent.GroupRole) *structs.GroupRole {
	return &structs.GroupRole{
		GroupID: row.GroupID,
		RoleID:  row.RoleID,
	}
}
