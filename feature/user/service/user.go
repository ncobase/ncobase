// Considered improvements are made here

package service

import (
	"context"
	"errors"
	"fmt"
	"ncobase/common/crypto"
	"ncobase/common/ecode"
	"ncobase/common/log"
	accessService "ncobase/feature/access/service"
	"ncobase/feature/user/data"
	"ncobase/feature/user/data/ent"
	"ncobase/feature/user/data/repository"
	"ncobase/feature/user/structs"
	"ncobase/helper"
)

// UserServiceInterface is the interface for the service.
type UserServiceInterface interface {
	GetMe(ctx context.Context) (*structs.UserMeshes, error)
	Get(ctx context.Context, username string) (*structs.UserMeshes, error)
	UpdatePassword(ctx context.Context, body *structs.UserPassword) error
	CreateUser(ctx context.Context, body *structs.UserMeshes) (*structs.UserMeshes, error)
	GetByID(ctx context.Context, u string) (*structs.UserMeshes, error)
	Delete(ctx context.Context, u string) error
	FindByID(ctx context.Context, id string) (*structs.UserMeshes, error)
	FindUser(ctx context.Context, m *structs.FindUser) (*structs.UserMeshes, error)
	VerifyPassword(ctx context.Context, userID string, password string) any
	Serialize(user *ent.User, sp ...*serializeUserParams) *structs.UserMeshes
	CountX(ctx context.Context, params *structs.ListUserParams) int
}

// userService is the struct for the service.
type userService struct {
	user        repository.UserRepositoryInterface
	userProfile repository.UserProfileRepositoryInterface
	as          *accessService.Service
}

// NewUserService creates a new service.
func NewUserService(d *data.Data, as *accessService.Service) UserServiceInterface {
	return &userService{
		user:        repository.NewUserRepository(d),
		userProfile: repository.NewUserProfileRepository(d),
		as:          as,
	}
}

// GetMe get current user service
func (s *userService) GetMe(ctx context.Context) (*structs.UserMeshes, error) {
	user, err := s.user.GetByID(ctx, helper.GetUserID(ctx))
	if err != nil {
		return nil, err
	}

	return s.Serialize(user, &serializeUserParams{WithProfile: true, WithRoles: true, WithTenants: true, WithGroups: true}), nil
}

// Get get user service
func (s *userService) Get(ctx context.Context, username string) (*structs.UserMeshes, error) {
	if username == "" {
		return nil, errors.New(ecode.FieldIsInvalid("username"))
	}
	user, err := s.FindUser(ctx, &structs.FindUser{Username: username})
	if err := handleEntError("User", err); err != nil {
		return nil, err
	}
	return user, nil
}

// UpdatePassword update user password service
func (s *userService) UpdatePassword(ctx context.Context, body *structs.UserPassword) error {
	if body.NewPassword == "" {
		return errors.New(ecode.FieldIsEmpty("new password"))
	}
	if body.Confirm != body.NewPassword {
		return errors.New(ecode.FieldIsInvalid("confirm password"))
	}
	verifyResult := s.VerifyPassword(ctx, helper.GetUserID(ctx), body.OldPassword)
	switch v := verifyResult.(type) {
	case VerifyPasswordResult:
		if v.Valid == false {
			return errors.New(v.Error)
		} else if v.Valid && v.NeedsPasswordSet == true { // print a log for user's first password setting
			log.Printf(context.Background(), "User %s is setting password for the first time", helper.GetUserID(ctx))
		}
	case error:
		return v
	}

	err := s.updatePassword(ctx, &structs.UserPassword{
		User:        helper.GetUserID(ctx),
		OldPassword: body.OldPassword,
		NewPassword: body.NewPassword,
		Confirm:     body.Confirm,
	})

	if err := handleEntError("User", err); err != nil {
		return err
	}

	return nil
}

// CreateUser creates a new user.
func (s *userService) CreateUser(ctx context.Context, body *structs.UserMeshes) (*structs.UserMeshes, error) {
	if body.User != nil && body.User.Username == "" {
		return nil, errors.New(ecode.FieldIsInvalid("username"))
	}

	user, err := s.user.Create(ctx, body.User)
	if err := handleEntError("User", err); err != nil {
		return nil, err
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
		if err := handleEntError("UserProfile", err); err != nil {
			return nil, err
		}
	}

	return s.Serialize(user), nil
}

// GetByID retrieves a user by their ID.
func (s *userService) GetByID(ctx context.Context, u string) (*structs.UserMeshes, error) {
	user, err := s.user.GetByID(ctx, u)
	if err := handleEntError("User", err); err != nil {
		return nil, err
	}

	return s.Serialize(user, &serializeUserParams{WithProfile: true}), nil
}

// Delete deletes a user by their ID.
func (s *userService) Delete(ctx context.Context, u string) error {
	err := s.user.Delete(ctx, u)
	if err := handleEntError("User", err); err != nil {
		return err
	}
	return nil
}

// FindByID find user by ID
func (s *userService) FindByID(ctx context.Context, id string) (*structs.UserMeshes, error) {
	user, err := s.user.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.Serialize(user, &serializeUserParams{WithProfile: true}), nil
}

// FindUser find user by username, email, or phone
func (s *userService) FindUser(ctx context.Context, m *structs.FindUser) (*structs.UserMeshes, error) {
	user, err := s.user.Find(ctx, &structs.FindUser{
		Username: m.Username,
		Email:    m.Email,
		Phone:    m.Phone,
	})
	if err != nil {
		return nil, err
	}
	return s.Serialize(user, &serializeUserParams{WithProfile: true}), nil
}

// VerifyPasswordResult Verify password result
type VerifyPasswordResult struct {
	Valid            bool
	NeedsPasswordSet bool
	Error            string
}

// VerifyPassword verify user password
func (s *userService) VerifyPassword(ctx context.Context, u string, password string) any {
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

// updatePassword update user password
func (s *userService) updatePassword(ctx context.Context, body *structs.UserPassword) error {
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

func (s *userService) Serialize(user *ent.User, sp ...*serializeUserParams) *structs.UserMeshes {
	ctx := context.Background()
	um := &structs.UserMeshes{
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

	params := &serializeUserParams{}
	if len(sp) > 0 {
		params = sp[0]
	}

	if params.WithProfile {
		if profile, _ := s.userProfile.Get(ctx, user.ID); profile != nil {
			um.Profile = &structs.UserProfileBody{
				DisplayName: profile.DisplayName,
				ShortBio:    profile.ShortBio,
				About:       &profile.About,
				Thumbnail:   &profile.Thumbnail,
				Links:       &profile.Links,
				Extras:      &profile.Extras,
			}
		}
	}

	// if params.WithTenants {
	// 	if tenants, _ := s.ts.UserTenant.UserBelongTenants(ctx, user.ID); len(tenants) > 0 {
	// 		for _, tenant := range tenants {
	// 			um.Tenants = append(um.Tenants, tenant)
	// 		}
	// 	}
	// }

	if params.WithRoles {
		if len(um.Tenants) > 0 {
			for _, tenant := range um.Tenants {
				roleIDs, _ := s.as.UserTenantRole.GetUserRolesInTenant(ctx, user.ID, tenant.ID)
				roles, _ := s.as.Role.GetByIDs(ctx, roleIDs)
				for _, role := range roles {
					um.Roles = append(um.Roles, role)
				}
			}
			// TODO: remove duplicate roles if needed
			// seenRoles := make(map[string]struct{})
			// for _, tenant := range um.Tenants {
			// 	roles, _ := s.userTenantRole.GetRolesByUserAndTenant(ctx, user.ID, tenant.ID)
			// 	for _, role := range roles {
			// 		roleID := role.ID
			// 		if _, found := seenRoles[roleID]; !found {
			// 			um.Roles = append(um.Roles, s.serializeRole(role))
			// 			seenRoles[roleID] = struct{}{}
			// 		}
			// 	}
			// }
		} else {
			roles, _ := s.as.UserRole.GetUserRoles(ctx, user.ID)
			for _, role := range roles {
				um.Roles = append(um.Roles, role)
			}
		}
	}

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
