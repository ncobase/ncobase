package service

import (
	"context"
	"errors"
	"ncobase/core/access/data"
	"ncobase/core/access/data/repository"
	"ncobase/core/access/structs"

	"github.com/ncobase/ncore/ecode"
)

// UserRoleServiceInterface is the interface for the service.
type UserRoleServiceInterface interface {
	AddRoleToUser(ctx context.Context, u string, r string) error
	CreateUserRole(ctx context.Context, body *structs.UserRole) (*structs.UserRole, error)
	GetUserRoles(ctx context.Context, u string) ([]*structs.ReadRole, error)
	GetUsersByRoleID(ctx context.Context, roleID string) ([]string, error)
	DeleteUserRoleByUserID(ctx context.Context, u string) error
	DeleteUserRoleByRoleID(ctx context.Context, roleID string) error
	RemoveRoleFromUser(ctx context.Context, u string, r string) error
}

// userRoleService is the struct for the service.
type userRoleService struct {
	userRole repository.UserRoleRepositoryInterface
}

// NewUserRoleService creates a new service.
func NewUserRoleService(d *data.Data) UserRoleServiceInterface {
	return &userRoleService{
		userRole: repository.NewUserRoleRepository(d),
	}
}

// AddRoleToUser adds a role to a user.
func (s *userRoleService) AddRoleToUser(ctx context.Context, u string, r string) error {
	_, err := s.userRole.Create(ctx, &structs.UserRole{UserID: u, RoleID: r})
	if err := handleEntError(ctx, "UserRole", err); err != nil {
		return err
	}

	return nil
}

// CreateUserRole creates a new user role.
func (s *userRoleService) CreateUserRole(ctx context.Context, body *structs.UserRole) (*structs.UserRole, error) {
	if body.UserID == "" || body.RoleID == "" {
		return nil, errors.New("UserID and RoleID are required")
	}
	userRole, err := s.userRole.Create(ctx, body)
	if err := handleEntError(ctx, "UserRole", err); err != nil {
		return nil, err
	}

	return repository.SerializeUserRole(userRole), nil
}

// GetUserRoles retrieves all roles associated with a user.
func (s *userRoleService) GetUserRoles(ctx context.Context, u string) ([]*structs.ReadRole, error) {
	roles, err := s.userRole.GetRolesByUserID(ctx, u)
	if err != nil {
		return nil, err
	}

	return repository.SerializeRoles(roles), nil
}

// GetUsersByRoleID retrieves users by role ID.
func (s *userRoleService) GetUsersByRoleID(ctx context.Context, roleID string) ([]string, error) {
	if roleID == "" {
		return nil, errors.New(ecode.FieldIsRequired("roleID"))
	}
	userIDs, err := s.userRole.GetUsersByRoleID(ctx, roleID)
	if err := handleEntError(ctx, "UserRole", err); err != nil {
		return nil, err
	}

	return userIDs, nil
}

// DeleteUserRoleByUserID deletes user roles by user ID.
func (s *userRoleService) DeleteUserRoleByUserID(ctx context.Context, u string) error {
	if u == "" {
		return errors.New(ecode.FieldIsRequired("userID"))
	}
	err := s.userRole.DeleteAllByUserID(ctx, u)
	if err := handleEntError(ctx, "UserRole", err); err != nil {
		return err
	}

	return nil
}

// DeleteUserRoleByRoleID deletes user roles by role ID.
func (s *userRoleService) DeleteUserRoleByRoleID(ctx context.Context, roleID string) error {
	if roleID == "" {
		return errors.New(ecode.FieldIsRequired("roleID"))
	}
	err := s.userRole.DeleteAllByRoleID(ctx, roleID)
	if err := handleEntError(ctx, "UserRole", err); err != nil {
		return err
	}

	return nil
}

// RemoveRoleFromUser removes a role from a user.
func (s *userRoleService) RemoveRoleFromUser(ctx context.Context, u string, r string) error {
	err := s.userRole.Delete(ctx, u, r)
	if err := handleEntError(ctx, "UserRole", err); err != nil {
		return err
	}
	return nil
}
