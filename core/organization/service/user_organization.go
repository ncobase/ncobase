package service

import (
	"context"
	"ncobase/core/organization/data"
	"ncobase/core/organization/data/repository"
	"ncobase/core/organization/structs"
	"ncobase/core/organization/wrapper"
	"time"
)

// UserOrganizationServiceInterface is the interface for the service.
type UserOrganizationServiceInterface interface {
	AddUserToOrganization(ctx context.Context, u string, o string, role structs.UserRole) (*structs.OrganizationMember, error)
	RemoveUserFromOrganization(ctx context.Context, u string, o string) error
	GetUserOrganizations(ctx context.Context, u string) ([]*structs.ReadOrganization, error)
	GetUserOrganizationIds(ctx context.Context, u string) ([]string, error)
	GetOrganizationMembers(ctx context.Context, o string) ([]*structs.OrganizationMember, error)
	IsUserMember(ctx context.Context, o string, u string) (bool, error)
	HasRole(ctx context.Context, o string, u string, role structs.UserRole) (bool, error)
	GetUserRole(ctx context.Context, o string, u string) (structs.UserRole, error)
	GetMembersByRole(ctx context.Context, o string, role structs.UserRole) ([]*structs.OrganizationMember, error)
}

// userOrganizationService is the struct for the service.
type userOrganizationService struct {
	os  OrganizationServiceInterface
	r   repository.UserOrganizationRepositoryInterface
	usw *wrapper.UserServiceWrapper
}

// NewUserOrganizationService creates a new service
func NewUserOrganizationService(d *data.Data, os OrganizationServiceInterface, usw *wrapper.UserServiceWrapper) UserOrganizationServiceInterface {
	return &userOrganizationService{
		os:  os,
		r:   repository.NewUserOrganizationRepository(d),
		usw: usw,
	}
}

// AddUserToOrganization adds a user to an organization
func (s *userOrganizationService) AddUserToOrganization(ctx context.Context, u string, o string, role structs.UserRole) (*structs.OrganizationMember, error) {
	if !structs.IsValidUserRole(role) {
		role = structs.RoleMember
	}

	userOrganization := &structs.UserOrganization{
		UserID: u,
		OrgID:  o,
		Role:   role,
	}

	_, err := s.r.Create(ctx, userOrganization)
	if err := handleEntError(ctx, "UserOrganization", err); err != nil {
		return nil, err
	}

	member := &structs.OrganizationMember{
		ID:      u,
		UserID:  u,
		Role:    role,
		AddedAt: time.Now().UnixMilli(),
	}

	// Enrich with user details using wrapper
	s.enrichMemberInfo(ctx, member)

	return member, nil
}

// RemoveUserFromOrganization removes a user from an organization.
func (s *userOrganizationService) RemoveUserFromOrganization(ctx context.Context, u string, o string) error {
	err := s.r.Delete(ctx, u, o)
	if err := handleEntError(ctx, "UserOrganization", err); err != nil {
		return err
	}
	return nil
}

// GetUserOrganizationIds retrieves all organization IDs associated with a user.
func (s *userOrganizationService) GetUserOrganizationIds(ctx context.Context, u string) ([]string, error) {
	organizationIDs, err := s.r.GetOrganizationsByUserID(ctx, u)
	if err := handleEntError(ctx, "UserOrganization", err); err != nil {
		return nil, err
	}

	return organizationIDs, nil
}

// GetUserOrganizations retrieves all organizations associated with a user.
func (s *userOrganizationService) GetUserOrganizations(ctx context.Context, u string) ([]*structs.ReadOrganization, error) {
	organizationIDs, err := s.r.GetOrganizationsByUserID(ctx, u)
	if err := handleEntError(ctx, "UserOrganization", err); err != nil {
		return nil, err
	}

	rows, err := s.os.GetByIDs(ctx, organizationIDs)
	if err := handleEntError(ctx, "Organization", err); err != nil {
		return nil, err
	}

	return rows, nil
}

// GetOrganizationMembers retrieves all members of a specific organization
func (s *userOrganizationService) GetOrganizationMembers(ctx context.Context, o string) ([]*structs.OrganizationMember, error) {
	userOrganizations, err := s.r.GetByOrgID(ctx, o)
	if err := handleEntError(ctx, "UserOrganization", err); err != nil {
		return nil, err
	}

	var members []*structs.OrganizationMember
	for _, uo := range userOrganizations {
		member := &structs.OrganizationMember{
			ID:      uo.UserID,
			UserID:  uo.UserID,
			Role:    structs.UserRole(uo.Role),
			AddedAt: uo.CreatedAt,
		}

		s.enrichMemberInfo(ctx, member)
		members = append(members, member)
	}

	return members, nil
}

// IsUserMember checks if a user is a member of an organization.
func (s *userOrganizationService) IsUserMember(ctx context.Context, o string, u string) (bool, error) {
	isMember, err := s.r.IsUserInOrganization(ctx, u, o)
	if err := handleEntError(ctx, "UserOrganization", err); err != nil {
		return false, err
	}

	return isMember, nil
}

// HasRole checks if a user has a specific role in an organization.
func (s *userOrganizationService) HasRole(ctx context.Context, o string, u string, role structs.UserRole) (bool, error) {
	hasRole, err := s.r.UserHasRole(ctx, u, o, role)
	if err := handleEntError(ctx, "UserOrganization", err); err != nil {
		return false, err
	}

	return hasRole, nil
}

// GetUserRole retrieves a user's role in an organization.
func (s *userOrganizationService) GetUserRole(ctx context.Context, o string, u string) (structs.UserRole, error) {
	userOrganization, err := s.r.GetUserOrganization(ctx, u, o)
	if err != nil {
		return "", err
	}

	return structs.UserRole(userOrganization.Role), nil
}

// GetMembersByRole retrieves all members with a specific role
func (s *userOrganizationService) GetMembersByRole(ctx context.Context, o string, role structs.UserRole) ([]*structs.OrganizationMember, error) {
	userOrganizations, err := s.r.GetByOrgIDAndRole(ctx, o, role)
	if err := handleEntError(ctx, "UserOrganization", err); err != nil {
		return nil, err
	}

	var members []*structs.OrganizationMember
	for _, uo := range userOrganizations {
		member := &structs.OrganizationMember{
			ID:      uo.UserID,
			UserID:  uo.UserID,
			Role:    structs.UserRole(uo.Role),
			AddedAt: uo.CreatedAt,
		}

		s.enrichMemberInfo(ctx, member)
		members = append(members, member)
	}

	return members, nil
}

// enrichMemberInfo enriches member with user details
func (s *userOrganizationService) enrichMemberInfo(ctx context.Context, member *structs.OrganizationMember) {
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
