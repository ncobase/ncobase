package service

import (
	"context"
	"errors"
	"ncobase/space/data"
	"ncobase/space/data/ent"
	"ncobase/space/data/repository"
	"ncobase/space/structs"
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
	gs          GroupServiceInterface
	r           repository.UserGroupRepositoryInterface
	user        UserServiceInterface
	userProfile UserProfileServiceInterface
}

// NewUserGroupService creates a new service.
func NewUserGroupService(d *data.Data, gs GroupServiceInterface) UserGroupServiceInterface {
	return &userGroupService{
		gs: gs,
		r:  repository.NewUserGroupRepository(d),
	}
}

// SetUserService sets the user service.
func (s *userGroupService) SetUserService(user UserServiceInterface, userProfile UserProfileServiceInterface) {
	s.user = user
	s.userProfile = userProfile
}

// AddUserToGroup adds a user to a group with a specific role.
func (s *userGroupService) AddUserToGroup(ctx context.Context, u string, g string, role structs.UserRole) (*structs.GroupMember, error) {
	// Validate role
	if !structs.IsValidUserRole(role) {
		role = structs.RoleMember // Set default role if invalid
	}

	// Get user details
	user, err := s.user.GetByID(ctx, u)
	if err != nil {
		return nil, errors.New("user not found")
	}
	userProfile, err := s.userProfile.Get(ctx, u)
	if err != nil {
		return nil, errors.New("user profile not found")
	}

	// Create the user-group relationship
	userGroup := &structs.UserGroup{
		UserID:  u,
		GroupID: g,
		Role:    role,
	}

	_, err = s.r.Create(ctx, userGroup)
	if err := handleEntError(ctx, "UserGroup", err); err != nil {
		return nil, err
	}

	// Create and return the member object with user details
	member := &structs.GroupMember{
		ID:      user.ID,
		UserID:  user.ID,
		Name:    userProfile.DisplayName,
		Email:   user.Email,
		Role:    role,
		AddedAt: time.Now().UnixMilli(),
		Avatar:  *userProfile.Thumbnail,
	}

	// TODO: Add last login
	// if user.LastLogin > 0 {
	// 	lastLogin := user.LastLogin
	// 	member.LastLogin = &lastLogin
	// }

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

// GetGroupMembers retrieves all members of a specific group.
func (s *userGroupService) GetGroupMembers(ctx context.Context, g string) ([]*structs.GroupMember, error) {
	userGroups, err := s.r.GetByGroupID(ctx, g)
	if err := handleEntError(ctx, "UserGroup", err); err != nil {
		return nil, err
	}

	// Get user details for each member
	var members []*structs.GroupMember
	for _, ug := range userGroups {
		user, err := s.user.GetByID(ctx, ug.UserID)
		if err != nil {
			continue // Skip if user not found
		}
		userProfile, err := s.userProfile.Get(ctx, ug.UserID)
		if err != nil {
			continue // Skip if user profile not found
		}

		member := &structs.GroupMember{
			ID:      user.ID,
			UserID:  user.ID,
			Name:    userProfile.DisplayName,
			Email:   user.Email,
			Role:    structs.UserRole(ug.Role),
			AddedAt: ug.CreatedAt,
			Avatar:  *userProfile.Thumbnail,
		}
		// TODO: Add last login
		// if user.LastLogin > 0 {
		// 	lastLogin := user.LastLogin
		// 	member.LastLogin = &lastLogin
		// }

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

// GetMembersByRole retrieves all members with a specific role in a group.
func (s *userGroupService) GetMembersByRole(ctx context.Context, g string, role structs.UserRole) ([]*structs.GroupMember, error) {
	userGroups, err := s.r.GetByGroupIDAndRole(ctx, g, role)
	if err := handleEntError(ctx, "UserGroup", err); err != nil {
		return nil, err
	}

	// Get user details for each member
	var members []*structs.GroupMember
	for _, ug := range userGroups {
		user, err := s.user.GetByID(ctx, ug.UserID)
		if err != nil {
			continue // Skip if user not found
		}

		userProfile, err := s.userProfile.Get(ctx, ug.UserID)
		if err != nil {
			continue // Skip if user profile not found
		}

		member := &structs.GroupMember{
			ID:      user.ID,
			UserID:  user.ID,
			Name:    userProfile.DisplayName,
			Email:   user.Email,
			Role:    structs.UserRole(ug.Role),
			AddedAt: ug.CreatedAt,
			Avatar:  *userProfile.Thumbnail,
		}
		// TODO: Add last login
		// if user.LastLogin > 0 {
		// 	lastLogin := user.LastLogin
		// 	member.LastLogin = &lastLogin
		// }

		members = append(members, member)
	}

	return members, nil
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
