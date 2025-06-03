package service

import (
	"context"
	"ncobase/space/data"
	"ncobase/space/data/ent"
	"ncobase/space/data/repository"
	"ncobase/space/structs"
	"ncobase/space/wrapper"
	"time"
)

// UserGroupServiceInterface is the interface for the service.
type UserGroupServiceInterface interface {
	AddUserToGroup(ctx context.Context, u string, g string, role structs.UserRole) (*structs.GroupMember, error)
	RemoveUserFromGroup(ctx context.Context, u string, g string) error
	GetUserGroups(ctx context.Context, u string) ([]*structs.ReadGroup, error)
	GetUserGroupIds(ctx context.Context, u string) ([]string, error)
	GetGroupMembers(ctx context.Context, g string) ([]*structs.GroupMember, error)
	IsUserMember(ctx context.Context, g string, u string) (bool, error)
	HasRole(ctx context.Context, g string, u string, role structs.UserRole) (bool, error)
	GetUserRole(ctx context.Context, g string, u string) (structs.UserRole, error)
	GetMembersByRole(ctx context.Context, g string, role structs.UserRole) ([]*structs.GroupMember, error)
}

// userGroupService is the struct for the service.
type userGroupService struct {
	gs  GroupServiceInterface
	r   repository.UserGroupRepositoryInterface
	usw *wrapper.UserServiceWrapper
}

// NewUserGroupService creates a new service
func NewUserGroupService(d *data.Data, gs GroupServiceInterface, usw *wrapper.UserServiceWrapper) UserGroupServiceInterface {
	return &userGroupService{
		gs:  gs,
		r:   repository.NewUserGroupRepository(d),
		usw: usw,
	}
}

// AddUserToGroup adds a user to a group
func (s *userGroupService) AddUserToGroup(ctx context.Context, u string, g string, role structs.UserRole) (*structs.GroupMember, error) {
	if !structs.IsValidUserRole(role) {
		role = structs.RoleMember
	}

	userGroup := &structs.UserGroup{
		UserID:  u,
		GroupID: g,
		Role:    role,
	}

	_, err := s.r.Create(ctx, userGroup)
	if err := handleEntError(ctx, "UserGroup", err); err != nil {
		return nil, err
	}

	member := &structs.GroupMember{
		ID:      u,
		UserID:  u,
		Role:    role,
		AddedAt: time.Now().UnixMilli(),
	}

	// Enrich with user details using wrapper
	s.enrichMemberInfo(ctx, member)

	return member, nil
}

// RemoveUserFromGroup removes a user from a group.
func (s *userGroupService) RemoveUserFromGroup(ctx context.Context, u string, g string) error {
	err := s.r.Delete(ctx, u, g)
	if err := handleEntError(ctx, "UserGroup", err); err != nil {
		return err
	}
	return nil
}

// GetUserGroupIds retrieves all group IDs associated with a user.
func (s *userGroupService) GetUserGroupIds(ctx context.Context, u string) ([]string, error) {
	groupIDs, err := s.r.GetGroupsByUserID(ctx, u)
	if err := handleEntError(ctx, "UserGroup", err); err != nil {
		return nil, err
	}

	return groupIDs, nil
}

// GetUserGroups retrieves all groups associated with a user.
func (s *userGroupService) GetUserGroups(ctx context.Context, u string) ([]*structs.ReadGroup, error) {
	groupIDs, err := s.r.GetGroupsByUserID(ctx, u)
	if err := handleEntError(ctx, "UserGroup", err); err != nil {
		return nil, err
	}

	rows, err := s.gs.GetByIDs(ctx, groupIDs)
	if err := handleEntError(ctx, "Group", err); err != nil {
		return nil, err
	}

	return rows, nil
}

// GetGroupMembers retrieves all members of a specific group
func (s *userGroupService) GetGroupMembers(ctx context.Context, g string) ([]*structs.GroupMember, error) {
	userGroups, err := s.r.GetByGroupID(ctx, g)
	if err := handleEntError(ctx, "UserGroup", err); err != nil {
		return nil, err
	}

	var members []*structs.GroupMember
	for _, ug := range userGroups {
		member := &structs.GroupMember{
			ID:      ug.UserID,
			UserID:  ug.UserID,
			Role:    structs.UserRole(ug.Role),
			AddedAt: ug.CreatedAt,
		}

		s.enrichMemberInfo(ctx, member)
		members = append(members, member)
	}

	return members, nil
}

// IsUserMember checks if a user is a member of a group.
func (s *userGroupService) IsUserMember(ctx context.Context, g string, u string) (bool, error) {
	isMember, err := s.r.IsUserInGroup(ctx, u, g)
	if err := handleEntError(ctx, "UserGroup", err); err != nil {
		return false, err
	}

	return isMember, nil
}

// HasRole checks if a user has a specific role in a group.
func (s *userGroupService) HasRole(ctx context.Context, g string, u string, role structs.UserRole) (bool, error) {
	hasRole, err := s.r.UserHasRole(ctx, u, g, role)
	if err := handleEntError(ctx, "UserGroup", err); err != nil {
		return false, err
	}

	return hasRole, nil
}

// GetUserRole retrieves a user's role in a group.
func (s *userGroupService) GetUserRole(ctx context.Context, g string, u string) (structs.UserRole, error) {
	userGroup, err := s.r.GetUserGroup(ctx, u, g)
	if err != nil {
		return "", err
	}

	return structs.UserRole(userGroup.Role), nil
}

// GetMembersByRole retrieves all members with a specific role
func (s *userGroupService) GetMembersByRole(ctx context.Context, g string, role structs.UserRole) ([]*structs.GroupMember, error) {
	userGroups, err := s.r.GetByGroupIDAndRole(ctx, g, role)
	if err := handleEntError(ctx, "UserGroup", err); err != nil {
		return nil, err
	}

	var members []*structs.GroupMember
	for _, ug := range userGroups {
		member := &structs.GroupMember{
			ID:      ug.UserID,
			UserID:  ug.UserID,
			Role:    structs.UserRole(ug.Role),
			AddedAt: ug.CreatedAt,
		}

		s.enrichMemberInfo(ctx, member)
		members = append(members, member)
	}

	return members, nil
}

// enrichMemberInfo enriches member with user details
func (s *userGroupService) enrichMemberInfo(ctx context.Context, member *structs.GroupMember) {
	// Get user basic info with fallback
	if user, err := s.usw.GetUserByID(ctx, member.UserID); err == nil && user != nil {
		if user.Username != "" {
			member.Name = user.Username
		}
		if user.Email != "" {
			member.Email = user.Email
		}
	}

	// Get user profile info with fallback
	if profile, err := s.usw.GetUserProfile(ctx, member.UserID); err == nil && profile != nil {
		if profile.DisplayName != "" {
			member.Name = profile.DisplayName
		}
		if profile.Thumbnail != nil && *profile.Thumbnail != "" {
			member.Avatar = *profile.Thumbnail
		}
	}
}

// Serializes serializes user groups.
func (s *userGroupService) Serializes(rows []*ent.UserGroup) []*structs.UserGroup {
	rs := make([]*structs.UserGroup, 0, len(rows))
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
