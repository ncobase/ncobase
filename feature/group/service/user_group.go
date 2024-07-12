package service

import (
	"context"
	"ncobase/feature/group/data"
	"ncobase/feature/group/data/ent"
	"ncobase/feature/group/data/repository"
	"ncobase/feature/group/structs"
)

// UserGroupServiceInterface is the interface for the service.
type UserGroupServiceInterface interface {
	AddUserToGroup(ctx context.Context, u string, g string) (*structs.UserGroup, error)
	RemoveUserFromGroup(ctx context.Context, u string, g string) error
	GetUserGroups(ctx context.Context, u string) ([]*structs.ReadGroup, error)
	GetUserGroupIds(ctx context.Context, u string) ([]string, error)
}

// userGroupService is the struct for the service.
type userGroupService struct {
	gs        GroupServiceInterface
	userGroup repository.UserGroupRepositoryInterface
}

// NewUserGroupService creates a new service.
func NewUserGroupService(d *data.Data, gs GroupServiceInterface) UserGroupServiceInterface {
	return &userGroupService{
		gs:        gs,
		userGroup: repository.NewUserGroupRepository(d),
	}
}

// AddUserToGroup adds a user to a group.
func (s *userGroupService) AddUserToGroup(ctx context.Context, u string, g string) (*structs.UserGroup, error) {
	row, err := s.userGroup.Create(ctx, &structs.UserGroup{UserID: u, GroupID: g})
	if err := handleEntError("UserGroup", err); err != nil {
		return nil, err
	}
	return s.Serialize(row), nil
}

// RemoveUserFromGroup removes a user from a group.
func (s *userGroupService) RemoveUserFromGroup(ctx context.Context, u string, g string) error {
	err := s.userGroup.Delete(ctx, u, g)
	if err := handleEntError("UserGroup", err); err != nil {
		return err
	}
	return nil
}

// GetUserGroupIds retrieves all group IDs associated with a user.
func (s *userGroupService) GetUserGroupIds(ctx context.Context, u string) ([]string, error) {
	groupIDs, err := s.userGroup.GetGroupsByUserID(ctx, u)
	if err := handleEntError("UserGroup", err); err != nil {
		return nil, err
	}

	return groupIDs, nil
}

// GetUserGroups retrieves all groups associated with a user.
func (s *userGroupService) GetUserGroups(ctx context.Context, u string) ([]*structs.ReadGroup, error) {
	groupIDs, err := s.userGroup.GetGroupsByUserID(ctx, u)
	if err := handleEntError("UserGroup", err); err != nil {
		return nil, err
	}

	rows, err := s.gs.GetByIDs(ctx, groupIDs)
	if err := handleEntError("Group", err); err != nil {
		return nil, err
	}

	return rows, nil
}

// Serializes serializes user groups.
func (s *userGroupService) Serializes(rows []*ent.UserGroup) []*structs.UserGroup {
	var rs []*structs.UserGroup
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize serializes a user group.
func (s *userGroupService) Serialize(row *ent.UserGroup) *structs.UserGroup {
	return &structs.UserGroup{
		UserID:  row.UserID,
		GroupID: row.GroupID,
	}
}
