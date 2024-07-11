package service

import (
	"context"
	"ncobase/common/resp"
	"ncobase/feature/group/data"
	"ncobase/feature/group/data/repository"
	"ncobase/feature/group/structs"
)

// GroupRoleServiceInterface is the interface for the service.
type GroupRoleServiceInterface interface {
	AddRoleToGroupService(ctx context.Context, groupID string, roleID string) (*resp.Exception, error)
	RemoveRoleFromGroupService(ctx context.Context, groupID string, roleID string) (*resp.Exception, error)
	GetGroupRolesService(ctx context.Context, groupID string) (*resp.Exception, error)
}

// groupRoleService is the struct for the service.
type groupRoleService struct {
	groupRole repository.GroupRoleRepositoryInterface
}

// NewGroupRoleService creates a new service.
func NewGroupRoleService(d *data.Data) GroupRoleServiceInterface {
	return &groupRoleService{
		groupRole: repository.NewGroupRoleRepository(d),
	}
}

// AddRoleToGroupService adds a role to a group.
func (s *groupRoleService) AddRoleToGroupService(ctx context.Context, groupID string, roleID string) (*resp.Exception, error) {
	_, err := s.groupRole.Create(ctx, &structs.GroupRole{GroupID: groupID, RoleID: roleID})
	if exception, err := handleEntError("GroupRole", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "Role added to group successfully",
	}, nil
}

// RemoveRoleFromGroupService removes a role from a group.
func (s *groupRoleService) RemoveRoleFromGroupService(ctx context.Context, groupID string, roleID string) (*resp.Exception, error) {
	err := s.groupRole.Delete(ctx, groupID, roleID)
	if exception, err := handleEntError("GroupRole", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "Role removed from group successfully",
	}, nil
}

// GetGroupRolesService retrieves roles associated with a group.
func (s *groupRoleService) GetGroupRolesService(ctx context.Context, groupID string) (*resp.Exception, error) {
	roles, err := s.groupRole.GetRolesByGroupID(ctx, groupID)
	if exception, err := handleEntError("GroupRole", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: roles,
	}, nil
}
