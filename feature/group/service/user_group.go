package service

import (
	"context"
	"ncobase/common/resp"
	"ncobase/feature/group/data"
	"ncobase/feature/group/data/repository"
	groupStructs "ncobase/feature/group/structs"
)

// UserGroupServiceInterface is the interface for the service.
type UserGroupServiceInterface interface {
	AddUserToGroupService(ctx context.Context, u string, g string) (*resp.Exception, error)
	RemoveUserFromGroupService(ctx context.Context, u string, g string) (*resp.Exception, error)
	GetUserGroupsService(ctx context.Context, u string) (*resp.Exception, error)
}

// userGroupService is the struct for the service.
type userGroupService struct {
	userGroup repository.UserGroupRepositoryInterface
}

// NewUserGroupService creates a new service.
func NewUserGroupService(d *data.Data) UserGroupServiceInterface {
	return &userGroupService{}
}

// AddUserToGroupService adds a user to a group.
func (s *userGroupService) AddUserToGroupService(ctx context.Context, u string, g string) (*resp.Exception, error) {
	_, err := s.userGroup.Create(ctx, &groupStructs.UserGroup{UserID: u, GroupID: g})
	if exception, err := handleEntError("UserGroup", err); exception != nil {
		return exception, err
	}
	return &resp.Exception{
		Data: "User added to group successfully",
	}, nil
}

// RemoveUserFromGroupService removes a user from a group.
func (s *userGroupService) RemoveUserFromGroupService(ctx context.Context, u string, g string) (*resp.Exception, error) {
	err := s.userGroup.Delete(ctx, u, g)
	if exception, err := handleEntError("UserGroup", err); exception != nil {
		return exception, err
	}
	return &resp.Exception{
		Data: "User removed from group successfully",
	}, nil
}

// GetUserGroupsService retrieves all groups associated with a user.
func (s *userGroupService) GetUserGroupsService(ctx context.Context, u string) (*resp.Exception, error) {
	groups, err := s.userGroup.GetByUserID(ctx, u)
	if exception, err := handleEntError("UserGroup", err); exception != nil {
		return exception, err
	}
	return &resp.Exception{
		Data: groups,
	}, nil
}
