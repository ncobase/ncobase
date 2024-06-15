// Considered improvements are made here

package service

import (
	"context"
	"fmt"
	"ncobase/internal/data/ent"
	"ncobase/internal/data/structs"
	"ncobase/internal/helper"
	"ncobase/pkg/crypto"
	"ncobase/pkg/ecode"
	"ncobase/pkg/log"
	"ncobase/pkg/resp"

	"github.com/gin-gonic/gin"
)

// GetMeService get current user service
func (svc *Service) GetMeService(c *gin.Context) (*resp.Exception, error) {
	return svc.readUser(c, helper.GetUserID(c))
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
func (svc *Service) UpdatePasswordService(c *gin.Context, body *structs.UserRequestBody) (*resp.Exception, error) {
	verifyResult := svc.verifyUserPassword(c, helper.GetUserID(c), body.OldPassword)
	switch v := verifyResult.(type) {
	case VerifyPasswordResult:
		if v.Valid == false {
			return resp.BadRequest(v.Error), nil
		} else if v.Valid && v.NeedsPasswordSet == true { // 记录第一次设置密码的日志
			log.Printf(context.Background(), "User %s is setting password for the first time", helper.GetUserID(c))
		}
	case error:
		return resp.InternalServer(v.Error()), nil
	}

	err := svc.updateUserPassword(c, helper.GetUserID(c), body.NewPassword)

	if exception, err := handleError("User", err); exception != nil {
		return exception, err
	}

	return nil, nil
}

// CreateUserService creates a new user.
func (svc *Service) CreateUserService(ctx context.Context, body *structs.UserRequestBody) (*resp.Exception, error) {
	if body.Username == "" {
		return resp.BadRequest(ecode.FieldIsInvalid("username")), nil
	}

	user, err := svc.user.Create(ctx, body)
	if exception, err := handleError("User", err); exception != nil {
		return exception, err
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
		Data: svc.serializeUser(user),
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

// AddUserToDomainService adds a user to a domain.
func (svc *Service) AddUserToDomainService(ctx context.Context, userID string, domainID string) (*resp.Exception, error) {
	_, err := svc.userDomain.Create(ctx, &structs.UserDomain{UserID: userID, DomainID: domainID})
	if exception, err := handleError("UserDomain", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "User added to domain successfully",
	}, nil
}

// RemoveUserFromDomainService removes a user from a domain.
func (svc *Service) RemoveUserFromDomainService(ctx context.Context, userID string, domainID string) (*resp.Exception, error) {
	err := svc.userDomain.Delete(ctx, userID, domainID)
	if exception, err := handleError("UserDomain", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "User removed from domain successfully",
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

// AddRoleToUserInDomainService adds a role to a user in a domain.
func (svc *Service) AddRoleToUserInDomainService(ctx context.Context, userID string, domainID string, roleID string) (*resp.Exception, error) {
	_, err := svc.userDomainRole.Create(ctx, &structs.UserDomainRole{UserID: userID, DomainID: domainID, RoleID: roleID})
	if exception, err := handleError("UserDomainRole", err); exception != nil {
		return exception, err
	}
	return &resp.Exception{
		Data: "Role added to user in domain successfully",
	}, nil
}

// RemoveRoleFromUserInDomainService removes a role from a user in a domain.
func (svc *Service) RemoveRoleFromUserInDomainService(ctx context.Context, userID string, domainID string, roleID string) (*resp.Exception, error) {
	err := svc.userDomainRole.DeleteByUserIDAndDomainIDAndRoleID(ctx, userID, domainID, roleID)
	if exception, err := handleError("UserDomainRole", err); exception != nil {
		return exception, err
	}
	return &resp.Exception{
		Data: "Role removed from user in domain successfully",
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
func (svc *Service) updateUserPassword(ctx context.Context, userID string, password string) error {
	err := svc.user.UpdatePassword(ctx, &structs.UserRequestBody{
		UserID:      userID,
		NewPassword: password,
	})
	if err != nil {
		log.Printf(context.Background(), "Error updating password for user %s: %v", userID, err)
	}

	return err
}

// findUserByID find user by ID
func (svc *Service) findUserByID(ctx context.Context, id string) (structs.User, error) {
	user, err := svc.user.GetByID(ctx, id)
	if err != nil {
		return structs.User{}, err
	}
	return svc.serializeUser(user, true), nil
}

// findUser find user by username, email, or phone
func (svc *Service) findUser(c *gin.Context, m *structs.FindUser) (structs.User, error) {
	ctx := helper.FromGinContext(c)
	user, err := svc.user.Find(ctx, &structs.FindUser{
		Username: m.Username,
		Email:    m.Email,
		Phone:    m.Phone,
	})
	if err != nil {
		return structs.User{}, err
	}
	return svc.serializeUser(user, true), nil
}

// serializeUser serialize user
func (svc *Service) serializeUser(user *ent.User, withProfile ...bool) structs.User {
	readUser := structs.User{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Phone:       user.Phone,
		IsCertified: user.IsCertified,
		IsAdmin:     user.IsAdmin,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
	if len(withProfile) > 0 && withProfile[0] {
		profile, _ := svc.userProfile.Get(context.Background(), user.ID)
		readUser.Profile = &structs.UserProfile{
			DisplayName: profile.DisplayName,
			ShortBio:    profile.ShortBio,
			About:       &profile.About,
			Thumbnail:   &profile.Thumbnail,
			Links:       &profile.Links,
		}
	}

	return readUser
}
