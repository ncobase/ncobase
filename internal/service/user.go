// Considered improvements are made here

package service

import (
	"context"
	"fmt"
	"ncobase/internal/data/ent"
	"ncobase/internal/data/structs"
	"ncobase/internal/helper"

	"github.com/ncobase/common/crypto"
	"github.com/ncobase/common/ecode"
	"github.com/ncobase/common/log"
	"github.com/ncobase/common/resp"

	"github.com/gin-gonic/gin"
)

// GetMeService get current user service
func (svc *Service) GetMeService(c *gin.Context) (*resp.Exception, error) {
	user, err := svc.user.GetByID(c, helper.GetUserID(c))
	if err != nil {
		return resp.BadRequest(err.Error()), err
	}
	return &resp.Exception{
		Data: svc.serializeUser(user, &serializeUserParams{WithProfile: true, WithRoles: true, WithTenants: true, WithGroups: true}),
	}, nil
}

// GetUserService get user service
func (svc *Service) GetUserService(c *gin.Context, username string) (*resp.Exception, error) {
	if username == "" {
		return resp.BadRequest(ecode.FieldIsInvalid("username")), nil
	}
	user, err := svc.findUser(c, &structs.FindUser{Username: username})
	if exception, err := handleError("User", err); exception != nil {
		return exception, err
	}
	return &resp.Exception{
		Data: user,
	}, nil
}

// UpdatePasswordService update user password service
func (svc *Service) UpdatePasswordService(c *gin.Context, body *structs.UserPassword) (*resp.Exception, error) {
	if body.NewPassword == "" {
		return resp.BadRequest(ecode.FieldIsEmpty("new password")), nil
	}
	if body.Confirm != body.NewPassword {
		return resp.BadRequest(ecode.FieldIsInvalid("confirm password")), nil
	}
	verifyResult := svc.verifyUserPassword(c, helper.GetUserID(c), body.OldPassword)
	switch v := verifyResult.(type) {
	case VerifyPasswordResult:
		if v.Valid == false {
			return resp.BadRequest(v.Error), nil
		} else if v.Valid && v.NeedsPasswordSet == true { // print a log for user's first password setting
			log.Printf(context.Background(), "User %s is setting password for the first time", helper.GetUserID(c))
		}
	case error:
		return resp.InternalServer(v.Error()), nil
	}

	err := svc.updateUserPassword(c, &structs.UserPassword{
		User:        helper.GetUserID(c),
		OldPassword: body.OldPassword,
		NewPassword: body.NewPassword,
		Confirm:     body.Confirm,
	})

	if exception, err := handleError("User", err); exception != nil {
		return exception, err
	}

	return nil, nil
}

// CreateUserService creates a new user.
func (svc *Service) CreateUserService(ctx context.Context, body *structs.UserMeshes) (*resp.Exception, error) {
	if body.User != nil && body.User.Username == "" {
		return resp.BadRequest(ecode.FieldIsInvalid("username")), nil
	}

	user, err := svc.user.Create(ctx, body.User)
	if exception, err := handleError("User", err); exception != nil {
		return exception, err
	}

	if body.Profile != nil {
		_, err := svc.userProfile.Create(ctx, &structs.UserProfileBody{
			ID:          user.ID,
			DisplayName: body.Profile.DisplayName,
			ShortBio:    body.Profile.ShortBio,
			About:       body.Profile.About,
			Thumbnail:   body.Profile.Thumbnail,
			Links:       body.Profile.Links,
		})
		if exception, err := handleError("UserProfile", err); exception != nil {
			return exception, err
		}
	}

	return &resp.Exception{
		Data: svc.serializeUser(user),
	}, nil
}

// GetUserByIDService retrieves a user by their ID.
func (svc *Service) GetUserByIDService(ctx context.Context, userID string) (*resp.Exception, error) {
	user, err := svc.user.GetByID(ctx, userID)
	if exception, err := handleError("User", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializeUser(user, &serializeUserParams{WithProfile: true}),
	}, nil
}

// DeleteUserService deletes a user by their ID.
func (svc *Service) DeleteUserService(ctx context.Context, userID string) (*resp.Exception, error) {
	err := svc.user.Delete(ctx, userID)
	if exception, err := handleError("User", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "User deleted successfully",
	}, nil
}

// AddUserToTenantService adds a user to a tenant.
func (svc *Service) AddUserToTenantService(ctx context.Context, userID string, tenantID string) (*resp.Exception, error) {
	_, err := svc.userTenant.Create(ctx, &structs.UserTenant{UserID: userID, TenantID: tenantID})
	if exception, err := handleError("UserTenant", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "User added to tenant successfully",
	}, nil
}

// RemoveUserFromTenantService removes a user from a tenant.
func (svc *Service) RemoveUserFromTenantService(ctx context.Context, userID string, tenantID string) (*resp.Exception, error) {
	err := svc.userTenant.Delete(ctx, userID, tenantID)
	if exception, err := handleError("UserTenant", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "User removed from tenant successfully",
	}, nil
}

// AddRoleToUserService adds a role to a user.
func (svc *Service) AddRoleToUserService(ctx context.Context, userID string, roleID string) (*resp.Exception, error) {
	_, err := svc.userRole.Create(ctx, &structs.UserRole{UserID: userID, RoleID: roleID})
	if exception, err := handleError("UserRole", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "Role added to user successfully",
	}, nil
}

// RemoveRoleFromUserService removes a role from a user.
func (svc *Service) RemoveRoleFromUserService(ctx context.Context, userID string, roleID string) (*resp.Exception, error) {
	err := svc.userRole.Delete(ctx, userID, roleID)
	if exception, err := handleError("UserRole", err); exception != nil {
		return exception, err
	}
	return &resp.Exception{
		Data: "Role removed from user successfully",
	}, nil
}

// AddUserToGroupService adds a user to a group.
func (svc *Service) AddUserToGroupService(ctx context.Context, userID string, groupID string) (*resp.Exception, error) {
	_, err := svc.userGroup.Create(ctx, &structs.UserGroup{UserID: userID, GroupID: groupID})
	if exception, err := handleError("UserGroup", err); exception != nil {
		return exception, err
	}
	return &resp.Exception{
		Data: "User added to group successfully",
	}, nil
}

// RemoveUserFromGroupService removes a user from a group.
func (svc *Service) RemoveUserFromGroupService(ctx context.Context, userID string, groupID string) (*resp.Exception, error) {
	err := svc.userGroup.Delete(ctx, userID, groupID)
	if exception, err := handleError("UserGroup", err); exception != nil {
		return exception, err
	}
	return &resp.Exception{
		Data: "User removed from group successfully",
	}, nil
}

// AddRoleToUserInTenantService adds a role to a user in a tenant.
func (svc *Service) AddRoleToUserInTenantService(ctx context.Context, userID string, tenantID string, roleID string) (*resp.Exception, error) {
	_, err := svc.userTenantRole.Create(ctx, &structs.UserTenantRole{UserID: userID, TenantID: tenantID, RoleID: roleID})
	if exception, err := handleError("UserTenantRole", err); exception != nil {
		return exception, err
	}
	return &resp.Exception{
		Data: "Role added to user in tenant successfully",
	}, nil
}

// RemoveRoleFromUserInTenantService removes a role from a user in a tenant.
func (svc *Service) RemoveRoleFromUserInTenantService(ctx context.Context, userID string, tenantID string, roleID string) (*resp.Exception, error) {
	err := svc.userTenantRole.DeleteByUserIDAndTenantIDAndRoleID(ctx, userID, tenantID, roleID)
	if exception, err := handleError("UserTenantRole", err); exception != nil {
		return exception, err
	}
	return &resp.Exception{
		Data: "Role removed from user in tenant successfully",
	}, nil
}

// GetUserRolesService retrieves all roles associated with a user.
func (svc *Service) GetUserRolesService(ctx context.Context, userID string) (*resp.Exception, error) {
	roles, err := svc.userRole.GetRolesByUserID(ctx, userID)
	if exception, err := handleError("UserRole", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: roles,
	}, nil
}

// GetUserGroupsService retrieves all groups associated with a user.
func (svc *Service) GetUserGroupsService(ctx context.Context, userID string) (*resp.Exception, error) {
	groups, err := svc.userGroup.GetByUserID(ctx, userID)
	if exception, err := handleError("UserGroup", err); exception != nil {
		return exception, err
	}
	return &resp.Exception{
		Data: groups,
	}, nil
}

// CreateUserRoleService creates a new user role.
func (svc *Service) CreateUserRoleService(ctx context.Context, body *structs.UserRole) (*resp.Exception, error) {
	if body.UserID == "" || body.RoleID == "" {
		return resp.BadRequest("UserID and RoleID are required"), nil
	}
	userRole, err := svc.userRole.Create(ctx, body)
	if exception, err := handleError("UserRole", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: userRole,
	}, nil
}

// GetUserRoleByUserIDService retrieves user roles by user ID.
func (svc *Service) GetUserRoleByUserIDService(ctx context.Context, userID string) (*resp.Exception, error) {
	if userID == "" {
		return resp.BadRequest("UserID is required"), nil
	}
	userRoles, err := svc.userRole.GetRolesByUserID(ctx, userID)
	if exception, err := handleError("UserRole", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: userRoles,
	}, nil
}

// GetUsersByRoleIDService retrieves users by role ID.
func (svc *Service) GetUsersByRoleIDService(ctx context.Context, roleID string) (*resp.Exception, error) {
	if roleID == "" {
		return resp.BadRequest("RoleID is required"), nil
	}
	users, err := svc.userRole.GetUsersByRoleID(ctx, roleID)
	if exception, err := handleError("UserRole", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: users,
	}, nil
}

// DeleteUserRoleByUserIDService deletes user roles by user ID.
func (svc *Service) DeleteUserRoleByUserIDService(ctx context.Context, userID string) (*resp.Exception, error) {
	if userID == "" {
		return resp.BadRequest("UserID is required"), nil
	}
	err := svc.userRole.DeleteAllByID(ctx, userID)
	if exception, err := handleError("UserRole", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "User roles deleted successfully",
	}, nil
}

// DeleteUserRoleByRoleIDService deletes user roles by role ID.
func (svc *Service) DeleteUserRoleByRoleIDService(ctx context.Context, roleID string) (*resp.Exception, error) {
	if roleID == "" {
		return resp.BadRequest("RoleID is required"), nil
	}
	err := svc.userRole.DeleteAllByRoleID(ctx, roleID)
	if exception, err := handleError("UserRole", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "User roles deleted successfully",
	}, nil
}

// ****** Internal methods of service

// readUser read user by ID
func (svc *Service) readUser(c *gin.Context, userID string) (*resp.Exception, error) {
	if userID == "" {
		return resp.BadRequest(ecode.FieldIsInvalid("userID")), nil
	}

	user, err := svc.findUserByID(c, userID)

	if exception, err := handleError("User", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: user,
	}, nil
}

// VerifyPasswordResult Verify password result
type VerifyPasswordResult struct {
	Valid            bool
	NeedsPasswordSet bool
	Error            string
}

// verifyUserPassword verify user password
func (svc *Service) verifyUserPassword(c *gin.Context, userID string, password string) any {
	user, err := svc.user.FindUser(c, &structs.FindUser{ID: userID})
	if ent.IsNotFound(err) {
		return VerifyPasswordResult{Valid: false, NeedsPasswordSet: false, Error: "user not found"}
	} else if err != nil {
		return VerifyPasswordResult{Valid: false, NeedsPasswordSet: false, Error: fmt.Sprintf("error getting user by ID: %v", err)}
	}
	if user.Password == "" {
		return VerifyPasswordResult{Valid: true, NeedsPasswordSet: true, Error: "user password not set"}
	}

	if crypto.ComparePassword(user.Password, password) {
		return VerifyPasswordResult{Valid: true, NeedsPasswordSet: false, Error: ""}
	}

	return VerifyPasswordResult{Valid: false, NeedsPasswordSet: false, Error: "wrong password"}
}

// updateUserPassword update user password
func (svc *Service) updateUserPassword(ctx context.Context, body *structs.UserPassword) error {
	err := svc.user.UpdatePassword(ctx, body)
	if err != nil {
		log.Printf(context.Background(), "Error updating password for user %s: %v", body.User, err)
	}

	return err
}

// findUserByID find user by ID
func (svc *Service) findUserByID(ctx context.Context, id string) (structs.UserMeshes, error) {
	user, err := svc.user.GetByID(ctx, id)
	if err != nil {
		return structs.UserMeshes{}, err
	}
	return svc.serializeUser(user, &serializeUserParams{WithProfile: true}), nil
}

// findUser find user by username, email, or phone
func (svc *Service) findUser(c *gin.Context, m *structs.FindUser) (structs.UserMeshes, error) {
	ctx := helper.FromGinContext(c)
	user, err := svc.user.Find(ctx, &structs.FindUser{
		Username: m.Username,
		Email:    m.Email,
		Phone:    m.Phone,
	})
	if err != nil {
		return structs.UserMeshes{}, err
	}
	return svc.serializeUser(user, &serializeUserParams{WithProfile: true}), nil
}

// SerializeParams serialize params
type serializeUserParams struct {
	WithProfile bool
	WithRoles   bool
	WithTenants bool
	WithGroups  bool
}

func (svc *Service) serializeUser(user *ent.User, sp ...*serializeUserParams) structs.UserMeshes {
	ctx := context.Background()
	um := structs.UserMeshes{
		User: &structs.UserBody{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			Phone:       user.Phone,
			IsCertified: user.IsCertified,
			IsAdmin:     user.IsAdmin,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		},
	}

	params := &serializeUserParams{}
	if len(sp) > 0 {
		params = sp[0]
	}

	if params.WithProfile {
		profile, _ := svc.userProfile.Get(ctx, user.ID)
		um.Profile = &structs.UserProfileBody{
			DisplayName: profile.DisplayName,
			ShortBio:    profile.ShortBio,
			About:       &profile.About,
			Thumbnail:   &profile.Thumbnail,
			Links:       &profile.Links,
		}
	}

	if params.WithTenants {
		tenants, _ := svc.userTenant.GetTenantsByUserID(ctx, user.ID)
		for _, tenant := range tenants {
			um.Tenants = append(um.Tenants, svc.serializeTenant(tenant))
		}
	}

	if params.WithRoles {
		if len(um.Tenants) > 0 {
			for _, tenant := range um.Tenants {
				roles, _ := svc.userTenantRole.GetRolesByUserAndTenant(ctx, user.ID, tenant.ID)
				for _, role := range roles {
					um.Roles = append(um.Roles, svc.serializeRole(role))
				}
			}
			// TODO: remove duplicate roles if needed
			// seenRoles := make(map[string]struct{})
			// for _, tenant := range um.Tenants {
			// 	roles, _ := svc.userTenantRole.GetRolesByUserAndTenant(ctx, user.ID, tenant.ID)
			// 	for _, role := range roles {
			// 		roleID := role.ID
			// 		if _, found := seenRoles[roleID]; !found {
			// 			um.Roles = append(um.Roles, svc.serializeRole(role))
			// 			seenRoles[roleID] = struct{}{}
			// 		}
			// 	}
			// }
		} else {
			roles, _ := svc.userRole.GetRolesByUserID(ctx, user.ID)
			for _, role := range roles {
				um.Roles = append(um.Roles, svc.serializeRole(role))
			}
		}
	}

	// TODO: group belongs to tenant
	// if params.WithGroups && len(um.Tenants) > 0 {
	// 	groups, _ := svc.userGroup.GetGroupsByUserID(ctx, user.ID)
	// 	for _, group := range groups {
	// 		um.Groups = append(um.Groups, svc.serializeGroup(group))
	// 	}
	// }

	return um
}
