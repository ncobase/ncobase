package service

import (
	"context"
	"ncobase/common/resp"
	"ncobase/feature/access/data"
	"ncobase/feature/access/data/repository"
	"ncobase/feature/access/structs"
)

// UserRoleServiceInterface is the interface for the service.
type UserRoleServiceInterface interface {
	AddRoleToUserService(ctx context.Context, u string, r string) (*resp.Exception, error)
	CreateUserRoleService(ctx context.Context, body *structs.UserRole) (*resp.Exception, error)
	GetUserRolesService(ctx context.Context, u string) (*resp.Exception, error)
	GetUsersByRoleIDService(ctx context.Context, roleID string) (*resp.Exception, error)
	DeleteUserRoleByUserIDService(ctx context.Context, u string) (*resp.Exception, error)
	DeleteUserRoleByRoleIDService(ctx context.Context, roleID string) (*resp.Exception, error)
	RemoveRoleFromUserService(ctx context.Context, u string, r string) (*resp.Exception, error)
}

// userRoleService is the struct for the service.
type userRoleService struct {
	rs       RoleServiceInterface
	userRole repository.UserRoleRepository
}

// NewUserRoleService creates a new service.
func NewUserRoleService(d *data.Data, rs RoleServiceInterface) UserRoleServiceInterface {
	return &userRoleService{
		rs:       rs,
		userRole: repository.NewUserRoleRepository(d),
	}
}

// AddRoleToUserService adds a role to a user.
func (s *userRoleService) AddRoleToUserService(ctx context.Context, u string, r string) (*resp.Exception, error) {
	_, err := s.userRole.Create(ctx, &structs.UserRole{UserID: u, RoleID: r})
	if exception, err := handleEntError("UserRole", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "Role added to user successfully",
	}, nil
}

// CreateUserRoleService creates a new user role.
func (s *userRoleService) CreateUserRoleService(ctx context.Context, body *structs.UserRole) (*resp.Exception, error) {
	if body.UserID == "" || body.RoleID == "" {
		return resp.BadRequest("UserID and RoleID are required"), nil
	}
	userRole, err := s.userRole.Create(ctx, body)
	if exception, err := handleEntError("UserRole", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: userRole,
	}, nil
}

// GetUserRolesService retrieves all roles associated with a user.
func (s *userRoleService) GetUserRolesService(ctx context.Context, u string) (*resp.Exception, error) {
	roles, err := s.userRole.GetRolesByUserID(ctx, u)
	if exception, err := handleEntError("UserRole", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: roles,
	}, nil
}

// GetUsersByRoleIDService retrieves users by role ID.
func (s *userRoleService) GetUsersByRoleIDService(ctx context.Context, roleID string) (*resp.Exception, error) {
	if roleID == "" {
		return resp.BadRequest("RoleID is required"), nil
	}
	users, err := s.userRole.GetUsersByRoleID(ctx, roleID)
	if exception, err := handleEntError("UserRole", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: users,
	}, nil
}

// DeleteUserRoleByUserIDService deletes user roles by user ID.
func (s *userRoleService) DeleteUserRoleByUserIDService(ctx context.Context, u string) (*resp.Exception, error) {
	if u == "" {
		return resp.BadRequest("UserID is required"), nil
	}
	err := s.userRole.DeleteAllByUserID(ctx, u)
	if exception, err := handleEntError("UserRole", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "User roles deleted successfully",
	}, nil
}

// DeleteUserRoleByRoleIDService deletes user roles by role ID.
func (s *userRoleService) DeleteUserRoleByRoleIDService(ctx context.Context, roleID string) (*resp.Exception, error) {
	if roleID == "" {
		return resp.BadRequest("RoleID is required"), nil
	}
	err := s.userRole.DeleteAllByRoleID(ctx, roleID)
	if exception, err := handleEntError("UserRole", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "User roles deleted successfully",
	}, nil
}

// RemoveRoleFromUserService removes a role from a user.
func (s *userRoleService) RemoveRoleFromUserService(ctx context.Context, u string, r string) (*resp.Exception, error) {
	err := s.userRole.Delete(ctx, u, r)
	if exception, err := handleEntError("UserRole", err); exception != nil {
		return exception, err
	}
	return &resp.Exception{
		Data: "Role removed from user successfully",
	}, nil
}
