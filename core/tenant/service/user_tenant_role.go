package service

import (
	"context"
	"fmt"
	"ncobase/tenant/data"
	"ncobase/tenant/data/ent"
	"ncobase/tenant/data/repository"
	"ncobase/tenant/structs"

	"github.com/ncobase/ncore/logging/logger"
)

// UserTenantRoleServiceInterface is the interface for the service.
type UserTenantRoleServiceInterface interface {
	AddRoleToUserInTenant(ctx context.Context, u, t, r string) (*structs.UserTenantRole, error)
	GetUserRolesInTenant(ctx context.Context, u, t string) ([]string, error)
	RemoveRoleFromUserInTenant(ctx context.Context, u, t, r string) error
	IsUserInRoleInTenant(ctx context.Context, u, t, r string) (bool, error)
	GetTenantUsersByRole(ctx context.Context, t, r string) ([]string, error)
	ListTenantUsers(ctx context.Context, tenantID string, params *structs.ListTenantUsersParams) (*structs.TenantUsersListResponse, error)
	UpdateUserTenantRole(ctx context.Context, userID, tenantID string, req *structs.UpdateUserTenantRoleRequest) (*structs.UserTenantRoleResponse, error)
	BulkUpdateUserTenantRoles(ctx context.Context, tenantID string, req *structs.BulkUpdateUserTenantRolesRequest) (*structs.BulkUpdateResponse, error)
}

// userTenantRoleService is the struct for the service.
type userTenantRoleService struct {
	userTenantRole repository.UserTenantRoleRepositoryInterface
}

// NewUserTenantRoleService creates a new service.
func NewUserTenantRoleService(d *data.Data) UserTenantRoleServiceInterface {
	return &userTenantRoleService{
		userTenantRole: repository.NewUserTenantRoleRepository(d),
	}
}

// AddRoleToUserInTenant adds a role to a user in a tenant.
func (s *userTenantRoleService) AddRoleToUserInTenant(ctx context.Context, u, t, r string) (*structs.UserTenantRole, error) {
	row, err := s.userTenantRole.Create(ctx, &structs.UserTenantRole{UserID: u, TenantID: t, RoleID: r})
	if err := handleEntError(ctx, "UserTenantRole", err); err != nil {
		return nil, err
	}
	return s.SerializeUserTenantRole(row), nil
}

// GetUserRolesInTenant retrieves all roles associated with a user in a tenant.
func (s *userTenantRoleService) GetUserRolesInTenant(ctx context.Context, u string, t string) ([]string, error) {
	roleIDs, err := s.userTenantRole.GetRolesByUserAndTenant(ctx, u, t)
	if err := handleEntError(ctx, "UserTenantRole", err); err != nil {
		return nil, err
	}
	return roleIDs, nil
}

// RemoveRoleFromUserInTenant removes a role from a user in a tenant.
func (s *userTenantRoleService) RemoveRoleFromUserInTenant(ctx context.Context, u, t, r string) error {
	err := s.userTenantRole.DeleteByUserIDAndTenantIDAndRoleID(ctx, u, t, r)
	if err := handleEntError(ctx, "UserTenantRole", err); err != nil {
		return err
	}
	return nil
}

// IsUserInRoleInTenant checks if a user has a specific role in a tenant.
func (s *userTenantRoleService) IsUserInRoleInTenant(ctx context.Context, u, t, r string) (bool, error) {
	hasRole, err := s.userTenantRole.IsUserInRoleInTenant(ctx, u, t, r)
	if err := handleEntError(ctx, "UserTenantRole", err); err != nil {
		return false, err
	}
	return hasRole, nil
}

// GetTenantUsersByRole retrieves all users with a specific role in a tenant.
func (s *userTenantRoleService) GetTenantUsersByRole(ctx context.Context, t, r string) ([]string, error) {
	userTenantRoles, err := s.userTenantRole.GetByTenantID(ctx, t)
	if err := handleEntError(ctx, "UserTenantRole", err); err != nil {
		return nil, err
	}

	var userIDs []string
	for _, utr := range userTenantRoles {
		if utr.RoleID == r {
			userIDs = append(userIDs, utr.UserID)
		}
	}

	return userIDs, nil
}

// ListTenantUsers retrieves all users in a tenant with their roles.
func (s *userTenantRoleService) ListTenantUsers(ctx context.Context, tenantID string, params *structs.ListTenantUsersParams) (*structs.TenantUsersListResponse, error) {
	// Get all user tenant roles for this tenant
	userTenantRoles, err := s.userTenantRole.GetByTenantID(ctx, tenantID)
	if err := handleEntError(ctx, "UserTenantRole", err); err != nil {
		return nil, err
	}

	// Group roles by user
	userRolesMap := make(map[string][]string)
	for _, utr := range userTenantRoles {
		// Filter by role if specified
		if params.RoleID != "" && utr.RoleID != params.RoleID {
			continue
		}
		userRolesMap[utr.UserID] = append(userRolesMap[utr.UserID], utr.RoleID)
	}

	// Convert to response format
	var users []structs.TenantUserInfo
	for userID, roleIDs := range userRolesMap {
		users = append(users, structs.TenantUserInfo{
			UserID:  userID,
			RoleIDs: roleIDs,
		})
	}

	// TODO: Add pagination support
	response := &structs.TenantUsersListResponse{
		Users: users,
		Total: len(users),
	}

	return response, nil
}

// UpdateUserTenantRole updates a user's role in a tenant.
func (s *userTenantRoleService) UpdateUserTenantRole(ctx context.Context, userID, tenantID string, req *structs.UpdateUserTenantRoleRequest) (*structs.UserTenantRoleResponse, error) {
	// Remove old role
	if err := s.userTenantRole.DeleteByUserIDAndTenantIDAndRoleID(ctx, userID, tenantID, req.OldRoleID); err != nil {
		logger.Warnf(ctx, "Failed to remove old role %s for user %s in tenant %s: %v", req.OldRoleID, userID, tenantID, err)
	}

	// Add new role
	_, err := s.userTenantRole.Create(ctx, &structs.UserTenantRole{
		UserID:   userID,
		TenantID: tenantID,
		RoleID:   req.NewRoleID,
	})
	if err := handleEntError(ctx, "UserTenantRole", err); err != nil {
		return nil, err
	}

	response := &structs.UserTenantRoleResponse{
		UserID:   userID,
		TenantID: tenantID,
		RoleID:   req.NewRoleID,
		Status:   "updated",
	}

	return response, nil
}

// BulkUpdateUserTenantRoles performs bulk updates on user tenant roles.
func (s *userTenantRoleService) BulkUpdateUserTenantRoles(ctx context.Context, tenantID string, req *structs.BulkUpdateUserTenantRolesRequest) (*structs.BulkUpdateResponse, error) {
	response := &structs.BulkUpdateResponse{
		Total: len(req.Updates),
	}

	for _, update := range req.Updates {
		var err error
		var result *structs.UserTenantRoleResponse

		switch update.Operation {
		case "add":
			var userRole *structs.UserTenantRole
			userRole, err = s.AddRoleToUserInTenant(ctx, update.UserID, tenantID, update.RoleID)
			if err == nil {
				result = &structs.UserTenantRoleResponse{
					UserID:   userRole.UserID,
					TenantID: userRole.TenantID,
					RoleID:   userRole.RoleID,
					Status:   "added",
				}
			}

		case "remove":
			err = s.RemoveRoleFromUserInTenant(ctx, update.UserID, tenantID, update.RoleID)
			if err == nil {
				result = &structs.UserTenantRoleResponse{
					UserID:   update.UserID,
					TenantID: tenantID,
					RoleID:   update.RoleID,
					Status:   "removed",
				}
			}

		case "update":
			if update.OldRoleID == "" {
				err = fmt.Errorf("old_role_id is required for update operation")
			} else {
				updateReq := &structs.UpdateUserTenantRoleRequest{
					OldRoleID: update.OldRoleID,
					NewRoleID: update.RoleID,
				}
				result, err = s.UpdateUserTenantRole(ctx, update.UserID, tenantID, updateReq)
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

// SerializeUserTenantRole serializes a user tenant role.
func (s *userTenantRoleService) SerializeUserTenantRole(row *ent.UserTenantRole) *structs.UserTenantRole {
	return &structs.UserTenantRole{
		UserID:   row.UserID,
		TenantID: row.TenantID,
		RoleID:   row.RoleID,
	}
}
