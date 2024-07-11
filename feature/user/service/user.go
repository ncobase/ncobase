// Considered improvements are made here

package service

import (
	"context"
	"fmt"
	"ncobase/common/crypto"
	"ncobase/common/ecode"
	"ncobase/common/log"
	"ncobase/common/resp"
	"ncobase/feature/user/data"
	"ncobase/feature/user/data/ent"
	"ncobase/feature/user/data/repository"
	"ncobase/feature/user/structs"
	"ncobase/helper"
)

// UserServiceInterface is the interface for the service.
type UserServiceInterface interface {
	GetMeService(ctx context.Context) (*resp.Exception, error)
	GetUserService(ctx context.Context, username string) (*resp.Exception, error)
	UpdatePasswordService(ctx context.Context, body *structs.UserPassword) (*resp.Exception, error)
	CreateUserService(ctx context.Context, body *structs.UserMeshes) (*resp.Exception, error)
	GetUserByIDService(ctx context.Context, u string) (*resp.Exception, error)
	DeleteUserService(ctx context.Context, u string) (*resp.Exception, error)
	FindUserByID(ctx context.Context, id string) (structs.UserMeshes, error)
	FindUser(ctx context.Context, m *structs.FindUser) (structs.UserMeshes, error)
	VerifyUserPassword(ctx context.Context, userID string, password string) interface{}
	UpdateUserPassword(ctx context.Context, body *structs.UserPassword) error
	SerializeUser(user *ent.User, sp ...*serializeUserParams) structs.UserMeshes
	CountX(ctx context.Context, params *structs.ListUserParams) int
}

// userService is the struct for the service.
type userService struct {
	user        repository.UserRepositoryInterface
	userProfile repository.UserProfileRepositoryInterface
}

// NewUserService creates a new service.
func NewUserService(d *data.Data) UserServiceInterface {
	return &userService{
		user:        repository.NewUserRepository(d),
		userProfile: repository.NewUserProfileRepository(d),
	}
}

// GetMeService get current user service
func (s *userService) GetMeService(ctx context.Context) (*resp.Exception, error) {
	user, err := s.user.GetByID(ctx, helper.GetUserID(ctx))
	if err != nil {
		return resp.BadRequest(err.Error()), err
	}
	return &resp.Exception{
		Data: s.SerializeUser(user, &serializeUserParams{WithProfile: true, WithRoles: true, WithTenants: true, WithGroups: true}),
	}, nil
}

// GetUserService get user service
func (s *userService) GetUserService(ctx context.Context, username string) (*resp.Exception, error) {
	if username == "" {
		return resp.BadRequest(ecode.FieldIsInvalid("username")), nil
	}
	user, err := s.FindUser(ctx, &structs.FindUser{Username: username})
	if exception, err := handleEntError("User", err); exception != nil {
		return exception, err
	}
	return &resp.Exception{
		Data: user,
	}, nil
}

// UpdatePasswordService update user password service
func (s *userService) UpdatePasswordService(ctx context.Context, body *structs.UserPassword) (*resp.Exception, error) {
	if body.NewPassword == "" {
		return resp.BadRequest(ecode.FieldIsEmpty("new password")), nil
	}
	if body.Confirm != body.NewPassword {
		return resp.BadRequest(ecode.FieldIsInvalid("confirm password")), nil
	}
	verifyResult := s.VerifyUserPassword(ctx, helper.GetUserID(ctx), body.OldPassword)
	switch v := verifyResult.(type) {
	case VerifyPasswordResult:
		if v.Valid == false {
			return resp.BadRequest(v.Error), nil
		} else if v.Valid && v.NeedsPasswordSet == true { // print a log for user's first password setting
			log.Printf(context.Background(), "User %s is setting password for the first time", helper.GetUserID(ctx))
		}
	case error:
		return resp.InternalServer(v.Error()), nil
	}

	err := s.UpdateUserPassword(ctx, &structs.UserPassword{
		User:        helper.GetUserID(ctx),
		OldPassword: body.OldPassword,
		NewPassword: body.NewPassword,
		Confirm:     body.Confirm,
	})

	if exception, err := handleEntError("User", err); exception != nil {
		return exception, err
	}

	return nil, nil
}

// CreateUserService creates a new user.
func (s *userService) CreateUserService(ctx context.Context, body *structs.UserMeshes) (*resp.Exception, error) {
	if body.User != nil && body.User.Username == "" {
		return resp.BadRequest(ecode.FieldIsInvalid("username")), nil
	}

	user, err := s.user.Create(ctx, body.User)
	if exception, err := handleEntError("User", err); exception != nil {
		return exception, err
	}

	if body.Profile != nil {
		_, err := s.userProfile.Create(ctx, &structs.UserProfileBody{
			ID:          user.ID,
			DisplayName: body.Profile.DisplayName,
			ShortBio:    body.Profile.ShortBio,
			About:       body.Profile.About,
			Thumbnail:   body.Profile.Thumbnail,
			Links:       body.Profile.Links,
			Extras:      body.Profile.Extras,
		})
		if exception, err := handleEntError("UserProfile", err); exception != nil {
			return exception, err
		}
	}

	return &resp.Exception{
		Data: s.SerializeUser(user),
	}, nil
}

// GetUserByIDService retrieves a user by their ID.
func (s *userService) GetUserByIDService(ctx context.Context, u string) (*resp.Exception, error) {
	user, err := s.user.GetByID(ctx, u)
	if exception, err := handleEntError("User", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: s.SerializeUser(user, &serializeUserParams{WithProfile: true}),
	}, nil
}

// DeleteUserService deletes a user by their ID.
func (s *userService) DeleteUserService(ctx context.Context, u string) (*resp.Exception, error) {
	err := s.user.Delete(ctx, u)
	if exception, err := handleEntError("User", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "User deleted successfully",
	}, nil
}

// FindUserByID find user by ID
func (s *userService) FindUserByID(ctx context.Context, id string) (structs.UserMeshes, error) {
	user, err := s.user.GetByID(ctx, id)
	if err != nil {
		return structs.UserMeshes{}, err
	}
	return s.SerializeUser(user, &serializeUserParams{WithProfile: true}), nil
}

// FindUser find user by username, email, or phone
func (s *userService) FindUser(ctx context.Context, m *structs.FindUser) (structs.UserMeshes, error) {
	user, err := s.user.Find(ctx, &structs.FindUser{
		Username: m.Username,
		Email:    m.Email,
		Phone:    m.Phone,
	})
	if err != nil {
		return structs.UserMeshes{}, err
	}
	return s.SerializeUser(user, &serializeUserParams{WithProfile: true}), nil
}

// VerifyPasswordResult Verify password result
type VerifyPasswordResult struct {
	Valid            bool
	NeedsPasswordSet bool
	Error            string
}

// VerifyUserPassword verify user password
func (s *userService) VerifyUserPassword(ctx context.Context, u string, password string) any {
	user, err := s.user.FindUser(ctx, &structs.FindUser{Username: u})
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

// UpdateUserPassword update user password
func (s *userService) UpdateUserPassword(ctx context.Context, body *structs.UserPassword) error {
	err := s.user.UpdatePassword(ctx, body)
	if err != nil {
		log.Printf(context.Background(), "Error updating password for user %s: %v", body.User, err)
	}

	return err
}

// SerializeParams serialize params
type serializeUserParams struct {
	WithProfile bool
	WithRoles   bool
	WithTenants bool
	WithGroups  bool
}

func (s *userService) SerializeUser(user *ent.User, sp ...*serializeUserParams) structs.UserMeshes {
	// ctx := context.Background()
	um := structs.UserMeshes{
		User: &structs.UserBody{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			Phone:       user.Phone,
			IsCertified: user.IsCertified,
			IsAdmin:     user.IsAdmin,
			Status:      user.Status,
			ExtraProps:  &user.Extras,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		},
	}

	// params := &serializeUserParams{}
	// if len(sp) > 0 {
	// 	params = sp[0]
	// }

	// if params.WithProfile {
	// 	if profile, _ := s.userProfile.Get(ctx, user.ID); profile != nil {
	// 		um.Profile = &structs.UserProfileBody{
	// 			DisplayName: profile.DisplayName,
	// 			ShortBio:    profile.ShortBio,
	// 			About:       &profile.About,
	// 			Thumbnail:   &profile.Thumbnail,
	// 			Links:       &profile.Links,
	// 			Extras:      &profile.Extras,
	// 		}
	// 	}
	// }
	//
	// if params.WithTenants {
	// 	if tenants, _ := s.userTenant.GetTenantsByUserID(ctx, user.ID); len(tenants) > 0 {
	// 		for _, tenant := range tenants {
	// 			um.Tenants = append(um.Tenants, s.serializeTenant(tenant))
	// 		}
	// 	}
	// }
	//
	// if params.WithRoles {
	// 	if len(um.Tenants) > 0 {
	// 		for _, tenant := range um.Tenants {
	// 			roles, _ := s.userTenantRole.GetRolesByUserAndTenant(ctx, user.ID, tenant.ID)
	// 			for _, role := range roles {
	// 				um.Roles = append(um.Roles, s.serializeRole(role))
	// 			}
	// 		}
	// 		// TODO: remove duplicate roles if needed
	// 		// seenRoles := make(map[string]struct{})
	// 		// for _, tenant := range um.Tenants {
	// 		// 	roles, _ := s.userTenantRole.GetRolesByUserAndTenant(ctx, user.ID, tenant.ID)
	// 		// 	for _, role := range roles {
	// 		// 		roleID := role.ID
	// 		// 		if _, found := seenRoles[roleID]; !found {
	// 		// 			um.Roles = append(um.Roles, s.serializeRole(role))
	// 		// 			seenRoles[roleID] = struct{}{}
	// 		// 		}
	// 		// 	}
	// 		// }
	// 	} else {
	// 		roles, _ := s.userRole.GetRolesByUserID(ctx, user.ID)
	// 		for _, role := range roles {
	// 			um.Roles = append(um.Roles, s.serializeRole(role))
	// 		}
	// 	}
	// }

	// TODO: group belongs to tenant
	// if params.WithGroups && len(um.Tenants) > 0 {
	// 	groups, _ := s.userGroup.GetGroupsByUserID(ctx, user.ID)
	// 	for _, group := range groups {
	// 		um.Groups = append(um.Groups, s.serializeGroup(group))
	// 	}
	// }

	return um
}

// CountX gets a count of users.
func (s *userService) CountX(ctx context.Context, params *structs.ListUserParams) int {
	return s.user.CountX(ctx, params)
}

// ****** Internal methods of service

// readUser read user by ID
func (s *userService) readUser(ctx context.Context, u string) (*resp.Exception, error) {
	if u == "" {
		return resp.BadRequest(ecode.FieldIsInvalid("user")), nil
	}

	user, err := s.FindUserByID(ctx, u)

	if exception, err := handleEntError("User", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: user,
	}, nil
}

// // serializeUserRoles serialize user roles
// func (svc *Service) serializeUserRoles(rows []*ent.UserRole) []*structs.UserRole {
// 	var userRoles []*structs.UserRole
// 	for _, row := range rows {
// 		userRoles = append(userRoles, svc.serializeUserRole(row))
// 	}
// 	return userRoles
// }
//
// // serializeUserRole serialize user role
// func (svc *Service) serializeUserRole(row *ent.UserRole) *structs.UserRole {
// 	return &structs.UserRole{
// 		UserID: row.UserID,
// 		RoleID: row.RoleID,
// 	}
// }
