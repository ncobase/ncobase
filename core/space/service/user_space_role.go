package service

import (
	"context"
	"fmt"
	"ncobase/core/space/data"
	"ncobase/core/space/data/repository"
	"ncobase/core/space/structs"

	"github.com/ncobase/ncore/logging/logger"
)

// UserSpaceRoleServiceInterface is the interface for the service.
type UserSpaceRoleServiceInterface interface {
	AddRoleToUserInSpace(ctx context.Context, u, t, r string) (*structs.UserSpaceRole, error)
	GetUserRolesInSpace(ctx context.Context, u, t string) ([]string, error)
	RemoveRoleFromUserInSpace(ctx context.Context, u, t, r string) error
	IsUserInRoleInSpace(ctx context.Context, u, t, r string) (bool, error)
	GetSpaceUsersByRole(ctx context.Context, t, r string) ([]string, error)
	ListSpaceUsers(ctx context.Context, spaceID string, params *structs.ListSpaceUsersParams) (*structs.SpaceUsersListResponse, error)
	UpdateUserSpaceRole(ctx context.Context, userID, spaceID string, req *structs.UpdateUserSpaceRoleRequest) (*structs.UserSpaceRoleResponse, error)
	BulkUpdateUserSpaceRoles(ctx context.Context, spaceID string, req *structs.BulkUpdateUserSpaceRolesRequest) (*structs.BulkUpdateResponse, error)
}

// userSpaceRoleService is the struct for the service.
type userSpaceRoleService struct {
	userSpaceRole repository.UserSpaceRoleRepositoryInterface
}

// NewUserSpaceRoleService creates a new service.
func NewUserSpaceRoleService(d *data.Data) UserSpaceRoleServiceInterface {
	return &userSpaceRoleService{
		userSpaceRole: repository.NewUserSpaceRoleRepository(d),
	}
}

// AddRoleToUserInSpace adds a role to a user in a space.
func (s *userSpaceRoleService) AddRoleToUserInSpace(ctx context.Context, u, t, r string) (*structs.UserSpaceRole, error) {
	row, err := s.userSpaceRole.Create(ctx, &structs.UserSpaceRole{UserID: u, SpaceID: t, RoleID: r})
	if err := handleEntError(ctx, "UserSpaceRole", err); err != nil {
		return nil, err
	}
	return repository.SerializeUserSpaceRole(row), nil
}

// GetUserRolesInSpace retrieves all roles associated with a user in a space.
func (s *userSpaceRoleService) GetUserRolesInSpace(ctx context.Context, u string, t string) ([]string, error) {
	roleIDs, err := s.userSpaceRole.GetRolesByUserAndSpace(ctx, u, t)
	if err := handleEntError(ctx, "UserSpaceRole", err); err != nil {
		return nil, err
	}
	return roleIDs, nil
}

// RemoveRoleFromUserInSpace removes a role from a user in a space.
func (s *userSpaceRoleService) RemoveRoleFromUserInSpace(ctx context.Context, u, t, r string) error {
	err := s.userSpaceRole.DeleteByUserIDAndSpaceIDAndRoleID(ctx, u, t, r)
	if err := handleEntError(ctx, "UserSpaceRole", err); err != nil {
		return err
	}
	return nil
}

// IsUserInRoleInSpace checks if a user has a specific role in a space.
func (s *userSpaceRoleService) IsUserInRoleInSpace(ctx context.Context, u, t, r string) (bool, error) {
	hasRole, err := s.userSpaceRole.IsUserInRoleInSpace(ctx, u, t, r)
	if err := handleEntError(ctx, "UserSpaceRole", err); err != nil {
		return false, err
	}
	return hasRole, nil
}

// GetSpaceUsersByRole retrieves all users with a specific role
func (s *userSpaceRoleService) GetSpaceUsersByRole(ctx context.Context, t, r string) ([]string, error) {
	userSpaceRoles, err := s.userSpaceRole.GetBySpaceID(ctx, t)
	if err := handleEntError(ctx, "UserSpaceRole", err); err != nil {
		return nil, err
	}

	var userIDs []string
	for _, utr := range userSpaceRoles {
		if utr.RoleID == r {
			userIDs = append(userIDs, utr.UserID)
		}
	}

	return userIDs, nil
}

// ListSpaceUsers retrieves all users in a space with their roles.
func (s *userSpaceRoleService) ListSpaceUsers(ctx context.Context, spaceID string, params *structs.ListSpaceUsersParams) (*structs.SpaceUsersListResponse, error) {
	// Get all user space roles for this space
	userSpaceRoles, err := s.userSpaceRole.GetBySpaceID(ctx, spaceID)
	if err := handleEntError(ctx, "UserSpaceRole", err); err != nil {
		return nil, err
	}

	// Group roles by user
	userRolesMap := make(map[string][]string)
	for _, utr := range userSpaceRoles {
		// Filter by role if specified
		if params.RoleID != "" && utr.RoleID != params.RoleID {
			continue
		}
		userRolesMap[utr.UserID] = append(userRolesMap[utr.UserID], utr.RoleID)
	}

	// Convert to response format
	var users []structs.SpaceUserInfo
	for userID, roleIDs := range userRolesMap {
		users = append(users, structs.SpaceUserInfo{
			UserID:  userID,
			RoleIDs: roleIDs,
		})
	}

	// TODO: Add pagination support
	response := &structs.SpaceUsersListResponse{
		Users: users,
		Total: len(users),
	}

	return response, nil
}

// UpdateUserSpaceRole updates a user's role in a space.
func (s *userSpaceRoleService) UpdateUserSpaceRole(ctx context.Context, userID, spaceID string, req *structs.UpdateUserSpaceRoleRequest) (*structs.UserSpaceRoleResponse, error) {
	// Remove old role
	if err := s.userSpaceRole.DeleteByUserIDAndSpaceIDAndRoleID(ctx, userID, spaceID, req.OldRoleID); err != nil {
		logger.Warnf(ctx, "Failed to remove old role %s for user %s in space %s: %v", req.OldRoleID, userID, spaceID, err)
	}

	// Add new role
	_, err := s.userSpaceRole.Create(ctx, &structs.UserSpaceRole{
		UserID:  userID,
		SpaceID: spaceID,
		RoleID:  req.NewRoleID,
	})
	if err := handleEntError(ctx, "UserSpaceRole", err); err != nil {
		return nil, err
	}

	response := &structs.UserSpaceRoleResponse{
		UserID:  userID,
		SpaceID: spaceID,
		RoleID:  req.NewRoleID,
		Status:  "updated",
	}

	return response, nil
}

// BulkUpdateUserSpaceRoles performs bulk updates on user space roles.
func (s *userSpaceRoleService) BulkUpdateUserSpaceRoles(ctx context.Context, spaceID string, req *structs.BulkUpdateUserSpaceRolesRequest) (*structs.BulkUpdateResponse, error) {
	response := &structs.BulkUpdateResponse{
		Total: len(req.Updates),
	}

	for _, update := range req.Updates {
		var err error
		var result *structs.UserSpaceRoleResponse

		switch update.Operation {
		case "add":
			var userRole *structs.UserSpaceRole
			userRole, err = s.AddRoleToUserInSpace(ctx, update.UserID, spaceID, update.RoleID)
			if err == nil {
				result = &structs.UserSpaceRoleResponse{
					UserID:  userRole.UserID,
					SpaceID: userRole.SpaceID,
					RoleID:  userRole.RoleID,
					Status:  "added",
				}
			}

		case "remove":
			err = s.RemoveRoleFromUserInSpace(ctx, update.UserID, spaceID, update.RoleID)
			if err == nil {
				result = &structs.UserSpaceRoleResponse{
					UserID:  update.UserID,
					SpaceID: spaceID,
					RoleID:  update.RoleID,
					Status:  "removed",
				}
			}

		case "update":
			if update.OldRoleID == "" {
				err = fmt.Errorf("old_role_id is required for update operation")
			} else {
				updateReq := &structs.UpdateUserSpaceRoleRequest{
					OldRoleID: update.OldRoleID,
					NewRoleID: update.RoleID,
				}
				result, err = s.UpdateUserSpaceRole(ctx, update.UserID, spaceID, updateReq)
			}

		default:
			err = fmt.Errorf("invalid operation: %s", update.Operation)
		}

		if err != nil {
			response.Failed++
			response.Errors = append(response.Errors, structs.BulkUpdateError{
				UserID: update.UserID,
				RoleID: update.RoleID,
				Error:  err.Error(),
			})
		} else {
			response.Success++
			if result != nil {
				response.Results = append(response.Results, *result)
			}
		}
	}

	return response, nil
}
