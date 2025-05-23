package service

import (
	"context"
	"errors"
	"fmt"
	accessService "ncobase/access/service"
	"ncobase/auth/data"
	"ncobase/auth/data/ent"
	codeAuthEnt "ncobase/auth/data/ent/codeauth"
	"ncobase/auth/structs"
	tenantService "ncobase/tenant/service"
	tenantStructs "ncobase/tenant/structs"
	userService "ncobase/user/service"
	userStructs "ncobase/user/structs"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/security/jwt"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation/validator"
)

// AccountServiceInterface is the interface for the service.
type AccountServiceInterface interface {
	Login(ctx context.Context, body *structs.LoginBody) (*types.JSON, error)
	Register(ctx context.Context, body *structs.RegisterBody) (*types.JSON, error)
	GetMe(ctx context.Context) (*structs.AccountMeshes, error)
	UpdatePassword(ctx context.Context, body *userStructs.UserPassword) error
	Tenant(ctx context.Context) (*tenantStructs.ReadTenant, error)
	Tenants(ctx context.Context) (paging.Result[*tenantStructs.ReadTenant], error)
	RefreshToken(ctx context.Context, refreshToken string) (*types.JSON, error)
}

// accountService is the struct for the service.
type accountService struct {
	d   *data.Data
	jtm *jwt.TokenManager
	cas CodeAuthServiceInterface
	ats AuthTenantServiceInterface
	us  *userService.Service
	as  *accessService.Service
	ts  *tenantService.Service
}

// NewAccountService creates a new service.
func NewAccountService(d *data.Data, jtm *jwt.TokenManager, cas CodeAuthServiceInterface, ats AuthTenantServiceInterface, us *userService.Service, as *accessService.Service, ts *tenantService.Service) AccountServiceInterface {
	return &accountService{
		d:   d,
		jtm: jtm,
		cas: cas,
		ats: ats,
		us:  us,
		as:  as,
		ts:  ts,
	}
}

// Login login service
func (s *accountService) Login(ctx context.Context, body *structs.LoginBody) (*types.JSON, error) {
	client := s.d.GetMasterEntClient()

	user, err := s.us.User.FindUser(ctx, &userStructs.FindUser{Username: body.Username})
	if err = handleEntError(ctx, "User", err); err != nil {
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
			if validator.IsEmpty(user.Email) {
				return nil, errors.New("has not set a password, and the mailbox is empty, please contact the administrator")
			}
			return s.cas.SendCode(ctx, &structs.SendCodeBody{Email: user.Email})
		}
	case error:
		return nil, v
	}

	// Get all tenants the user belongs to
	userTenants, _ := s.ts.UserTenant.UserBelongTenants(ctx, user.ID)
	var tenantIDs []string
	if len(userTenants) > 0 {
		for _, t := range userTenants {
			tenantIDs = append(tenantIDs, t.ID)
		}
	}

	// Set tenant ID in context if there's a default tenant
	defaultTenant, err := s.ts.UserTenant.UserBelongTenant(ctx, user.ID)
	if err == nil && defaultTenant != nil {
		ctx = ctxutil.SetTenantID(ctx, defaultTenant.ID)
	}

	// Create token payload
	payload, err := CreateUserTokenPayload(ctx, s.as, user.ID, tenantIDs)
	if err != nil {
		return nil, err
	}

	tokens, err := generateTokensForUser(ctx, s.jtm, client, payload)
	if err != nil {
		return nil, err
	}

	// Include tenant information in the response
	(*tokens)["tenant_ids"] = tenantIDs
	(*tokens)["default_tenant"] = defaultTenant

	return tokens, nil
}

// RefreshToken refreshes the access token using a refresh token
func (s *accountService) RefreshToken(ctx context.Context, refreshToken string) (*types.JSON, error) {
	client := s.d.GetMasterEntClient()

	// Verify the refresh token
	payload, err := s.jtm.DecodeToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Check if the token is a refresh token
	if payload["sub"] != "refresh" {
		return nil, errors.New("invalid token type")
	}

	// Extract user ID from payload
	payloadData, ok := payload["payload"].(map[string]any)
	if !ok {
		return nil, errors.New("invalid token payload")
	}

	userID, ok := payloadData["user_id"].(string)
	if !ok || userID == "" {
		return nil, errors.New("invalid user information in token")
	}

	// Validate user exists
	user, err := s.us.User.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Get all tenants the user belongs to
	userTenants, _ := s.ts.UserTenant.UserBelongTenants(ctx, user.ID)
	var tenantIDs []string
	if len(userTenants) > 0 {
		for _, t := range userTenants {
			tenantIDs = append(tenantIDs, t.ID)
		}
	}

	// Set tenant ID in context if there's a default tenant
	defaultTenant, err := s.ts.UserTenant.UserBelongTenant(ctx, user.ID)
	if err == nil && defaultTenant != nil {
		ctx = ctxutil.SetTenantID(ctx, defaultTenant.ID)
	}

	// Create token payload
	tokenPayload, err := CreateUserTokenPayload(ctx, s.as, user.ID, tenantIDs)
	if err != nil {
		return nil, err
	}

	// Create a new auth token entry
	tx, err := client.Tx(ctx)
	if err != nil {
		return nil, err
	}

	authToken, err := createAuthToken(ctx, tx, user.ID)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return nil, fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return nil, err
	}

	// Generate new access and refresh tokens with full payload
	accessToken, newRefreshToken := generateUserToken(s.jtm, tokenPayload, authToken.ID)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &types.JSON{
		"access_token":   accessToken,
		"refresh_token":  newRefreshToken,
		"tenant_ids":     tenantIDs,
		"default_tenant": defaultTenant,
	}, nil
}

// Register register service
func (s *accountService) Register(ctx context.Context, body *structs.RegisterBody) (*types.JSON, error) {
	client := s.d.GetMasterEntClient()

	// Decode register token
	payload, err := decodeRegisterToken(s.jtm, body.RegisterToken)
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
	if err = disableCodeAuth(ctx, client, payload["id"].(string)); err != nil {
		return nil, err
	}

	// Create user, profile, tenant and tokens in a transaction
	tx, err := client.Tx(ctx)
	if err != nil {
		return nil, err
	}

	rst, err := createUserAndProfile(ctx, s, body, payload)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, err
	}

	user := rst["user"].(*userStructs.ReadUser)

	// Create tenant
	tenant, err := s.ats.IsCreateTenant(ctx, &tenantStructs.CreateTenantBody{
		TenantBody: tenantStructs.TenantBody{Name: body.Tenant, CreatedBy: &user.ID, UpdatedBy: &user.ID},
	})
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, err
	}

	// Get tenant IDs
	var tenantIDs []string
	if tenant != nil {
		tenantIDs = append(tenantIDs, tenant.ID)
		// Set tenant in context for role assignment
		ctx = ctxutil.SetTenantID(ctx, tenant.ID)
	}

	// Create token payload
	tokenPayload, err := CreateUserTokenPayload(ctx, s.as, user.ID, tenantIDs)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, err
	}

	authToken, err := createAuthToken(ctx, tx, user.ID)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, err
	}

	accessToken, refreshToken := generateUserToken(s.jtm, tokenPayload, authToken.ID)
	if accessToken == "" || refreshToken == "" {
		if err := tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, errors.New("authorize is not created")
	}

	return &types.JSON{
		"access_token":   accessToken,
		"refresh_token":  refreshToken,
		"tenant_ids":     tenantIDs,
		"default_tenant": tenant,
	}, tx.Commit()
}

// decodeRegisterToken decodes the register token
func decodeRegisterToken(jtm *jwt.TokenManager, token string) (types.JSON, error) {
	decoded, err := jtm.DecodeToken(token)
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
		UserID:      user.ID,
		DisplayName: body.DisplayName,
		ShortBio:    body.ShortBio,
	})
	if err != nil {
		return nil, err
	}
	return types.JSON{"user": user, "profile": userProfile}, nil
}

// GetMe get current user service
func (s *accountService) GetMe(ctx context.Context) (*structs.AccountMeshes, error) {
	user, err := s.us.User.GetByID(ctx, ctxutil.GetUserID(ctx))
	if err != nil {
		return nil, err
	}

	return s.Serialize(user, &serializeUserParams{WithProfile: true, WithRoles: true, WithTenants: true, WithGroups: true}), nil
}

// UpdatePassword update user password service
func (s *accountService) UpdatePassword(ctx context.Context, body *userStructs.UserPassword) error {
	body.User = ctxutil.GetUserID(ctx)
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
	userID := ctxutil.GetUserID(ctx)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	// Retrieve the tenant associated with the user
	row, err := s.ts.Tenant.GetByUser(ctx, userID)
	if err = handleEntError(ctx, "Tenant", err); err != nil {
		return nil, err
	}

	return row, nil
}

// Tenants retrieves the tenant associated with the user's account.
func (s *accountService) Tenants(ctx context.Context) (paging.Result[*tenantStructs.ReadTenant], error) {
	userID := ctxutil.GetUserID(ctx)
	if userID == "" {
		return paging.Result[*tenantStructs.ReadTenant]{}, errors.New("invalid user ID")
	}

	rows, err := s.ts.Tenant.List(ctx, &tenantStructs.ListTenantParams{
		User: userID,
	})
	if err = handleEntError(ctx, "Tenants", err); err != nil {
		return paging.Result[*tenantStructs.ReadTenant]{}, err
	}

	return rows, nil
}
