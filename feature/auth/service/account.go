package service

import (
	"context"
	"errors"
	"fmt"
	"ncobase/common/jwt"
	"ncobase/common/paging"
	"ncobase/common/types"
	"ncobase/common/validator"
	accessService "ncobase/feature/access/service"
	"ncobase/feature/auth/data"
	"ncobase/feature/auth/data/ent"
	codeAuthEnt "ncobase/feature/auth/data/ent/codeauth"
	"ncobase/feature/auth/middleware"
	"ncobase/feature/auth/structs"
	tenantService "ncobase/feature/tenant/service"
	tenantStructs "ncobase/feature/tenant/structs"
	userService "ncobase/feature/user/service"
	userStructs "ncobase/feature/user/structs"
	"ncobase/helper"
)

// AccountServiceInterface is the interface for the service.
type AccountServiceInterface interface {
	Login(ctx context.Context, body *structs.LoginBody) (*types.JSON, error)
	Register(ctx context.Context, body *structs.RegisterBody) (*types.JSON, error)
	GetMe(ctx context.Context) (*structs.AccountMeshes, error)
	UpdatePassword(ctx context.Context, body *userStructs.UserPassword) error
	Tenant(ctx context.Context) (*tenantStructs.ReadTenant, error)
	Tenants(ctx context.Context) (paging.Result[*tenantStructs.ReadTenant], error)
}

// accountService is the struct for the service.
type accountService struct {
	d   *data.Data
	cas CodeAuthServiceInterface
	ats AuthTenantServiceInterface
	us  *userService.Service
	as  *accessService.Service
	ts  *tenantService.Service
}

// NewAccountService creates a new service.
func NewAccountService(d *data.Data, cas CodeAuthServiceInterface, ats AuthTenantServiceInterface, us *userService.Service, as *accessService.Service, ts *tenantService.Service) AccountServiceInterface {
	return &accountService{
		d:   d,
		cas: cas,
		ats: ats,
		us:  us,
		as:  as,
		ts:  ts,
	}
}

// Login login service
func (s *accountService) Login(ctx context.Context, body *structs.LoginBody) (*types.JSON, error) {
	conf := helper.GetConfig(ctx)
	client := s.d.GetEntClient()

	user, err := s.us.User.FindUser(ctx, &userStructs.FindUser{Username: body.Username})
	if err := handleEntError("User", err); err != nil {
		return nil, err
	}

	if user.Status != 0 {
		return nil, errors.New("account has been disabled, please contact the administrator")
	}

	verifyResult := s.us.User.VerifyPassword(ctx, user.ID, body.Password)
	switch v := verifyResult.(type) {
	case userService.VerifyPasswordResult:
		if v.Valid == false {
			return nil, errors.New(v.Error)
		} else if v.Valid && v.NeedsPasswordSet == true {
			// The user has not set a password and the mailbox is empty
			if validator.IsEmpty(user.Email) {
				return nil, errors.New("has not set a password, and the mailbox is empty, please contact the administrator")
			}
			return s.cas.SendCode(ctx, &structs.SendCodeBody{Email: user.Email})
		}
	case error:
		return nil, v
	}

	return generateTokensForUser(ctx, conf, client, user)
}

// Register register service
func (s *accountService) Register(ctx context.Context, body *structs.RegisterBody) (*types.JSON, error) {
	conf := helper.GetConfig(ctx)
	client := s.d.GetEntClient()

	// Decode register token
	payload, err := decodeRegisterToken(conf.Auth.JWT.Secret, body.RegisterToken)
	if err != nil {
		return nil, errors.New("register token decode failed")
	}

	// Verify user existence
	existedUser, err := s.us.User.FindUser(ctx, &userStructs.FindUser{
		Username: body.Username,
		Email:    payload["email"].(string),
		Phone:    body.Phone,
	})
	if err != nil && existedUser != nil {
		return nil, errors.New(getExistMessage(&userStructs.FindUser{
			Username: existedUser.Username,
			Email:    existedUser.Email,
			Phone:    existedUser.Phone,
		}, body))
	}

	// Disable code
	if err := disableCodeAuth(ctx, client, payload["id"].(string)); err != nil {
		return nil, err
	}

	// Create user, profile, tenant and tokens in a transaction
	tx, err := client.Tx(ctx)
	if err != nil {
		return nil, err
	}

	rst, err := createUserAndProfile(ctx, s, body, payload)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, err
	}

	user := rst["user"].(*userStructs.ReadUser)

	if _, err := s.ats.IsCreateTenant(ctx, &tenantStructs.CreateTenantBody{
		TenantBody: tenantStructs.TenantBody{Name: body.Tenant, CreatedBy: &user.ID, UpdatedBy: &user.ID},
	}); err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, err
	}

	authToken, err := createAuthToken(ctx, tx, user.ID)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, err
	}

	accessToken, refreshToken := middleware.GenerateUserToken(conf.Auth.JWT.Secret, user.ID, authToken.ID)
	if accessToken == "" || refreshToken == "" {
		if err := tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, errors.New("authorize is not created")
	}

	return &types.JSON{
		"id":           user.ID,
		"access_token": accessToken,
	}, tx.Commit()
}

// Helper functions for Register
func decodeRegisterToken(secret, token string) (types.JSON, error) {
	decoded, err := jwt.DecodeToken(secret, token)
	if err != nil {
		return nil, err
	}
	if decoded["sub"] != "email-register" {
		return nil, fmt.Errorf("not valid authorize information")
	}
	return decoded["payload"].(types.JSON), nil
}

func getExistMessage(existedUser *userStructs.FindUser, body *structs.RegisterBody) string {
	switch {
	case existedUser.Username == body.Username:
		return "Username already exists"
	case existedUser.Phone == body.Phone:
		return "Phone already exists"
	default:
		return "Email already exists"
	}
}

func disableCodeAuth(ctx context.Context, client *ent.Client, id string) error {
	_, err := client.CodeAuth.Update().Where(codeAuthEnt.ID(id)).SetLogged(true).Save(ctx)
	return err
}

func createUserAndProfile(ctx context.Context, svc *accountService, body *structs.RegisterBody, payload types.JSON) (types.JSON, error) {
	// create user
	user, err := svc.us.User.CreateUser(ctx, &userStructs.UserBody{
		Username: body.Username,
		Email:    payload["email"].(string),
		Phone:    body.Phone,
	})

	if err != nil {
		return nil, err
	}

	// create user profile
	userProfile, err := svc.us.UserProfile.Create(ctx, &userStructs.UserProfileBody{
		ID:          user.ID,
		DisplayName: body.DisplayName,
		ShortBio:    body.ShortBio,
	})
	if err != nil {
		return nil, err
	}
	return types.JSON{"user": user, "profile": userProfile}, nil
}

func createAuthToken(ctx context.Context, tx *ent.Tx, userID string) (*ent.AuthToken, error) {
	return tx.AuthToken.Create().SetUserID(userID).Save(ctx)
}

// GetMe get current user service
func (s *accountService) GetMe(ctx context.Context) (*structs.AccountMeshes, error) {
	user, err := s.us.User.GetByID(ctx, helper.GetUserID(ctx))
	if err != nil {
		return nil, err
	}

	return s.Serialize(user, &serializeUserParams{WithProfile: true, WithRoles: true, WithTenants: true, WithGroups: true}), nil
}

// UpdatePassword update user password service
func (s *accountService) UpdatePassword(ctx context.Context, body *userStructs.UserPassword) error {
	body.User = helper.GetUserID(ctx)
	return s.us.User.UpdatePassword(ctx, body)
}

// SerializeParams serialize params
type serializeUserParams struct {
	WithProfile bool
	WithRoles   bool
	WithTenants bool
	WithGroups  bool
}

func (s *accountService) Serialize(user *userStructs.ReadUser, sp ...*serializeUserParams) *structs.AccountMeshes {
	ctx := context.Background()
	um := &structs.AccountMeshes{
		User: user,
	}

	params := &serializeUserParams{}
	if len(sp) > 0 {
		params = sp[0]
	}

	if params.WithProfile {
		if profile, _ := s.us.UserProfile.Get(ctx, user.ID); profile != nil {
			um.Profile = profile
		}
	}

	if params.WithTenants {
		if tenants, _ := s.ts.UserTenant.UserBelongTenants(ctx, user.ID); len(tenants) > 0 {
			um.Tenants = tenants
			// for _, tenant := range tenants {
			// 	um.Tenants = append(um.Tenants, tenant)
			// }
		}
	}

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

// Tenant retrieves the tenant associated with the user's account.
func (s *accountService) Tenant(ctx context.Context) (*tenantStructs.ReadTenant, error) {
	userID := helper.GetUserID(ctx)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	// Retrieve the tenant associated with the user
	row, err := s.ts.Tenant.GetByUser(ctx, userID)
	if err := handleEntError("Tenant", err); err != nil {
		return nil, err
	}

	return row, nil
}

// Tenants retrieves the tenant associated with the user's account.
func (s *accountService) Tenants(ctx context.Context) (paging.Result[*tenantStructs.ReadTenant], error) {
	userID := helper.GetUserID(ctx)
	if userID == "" {
		return paging.Result[*tenantStructs.ReadTenant]{}, errors.New("invalid user ID")
	}

	rows, err := s.ts.Tenant.List(ctx, &tenantStructs.ListTenantParams{
		User: userID,
	})
	if err := handleEntError("Tenants", err); err != nil {
		return paging.Result[*tenantStructs.ReadTenant]{}, err
	}

	return rows, nil
}
