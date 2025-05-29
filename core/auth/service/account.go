package service

import (
	"context"
	"errors"
	"fmt"
	"ncobase/auth/data"
	"ncobase/auth/data/ent"
	codeAuthEnt "ncobase/auth/data/ent/codeauth"
	"ncobase/auth/event"
	"ncobase/auth/structs"
	"ncobase/auth/wrapper"
	tenantStructs "ncobase/tenant/structs"
	userService "ncobase/user/service"
	userStructs "ncobase/user/structs"
	"time"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/security/jwt"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation/validator"
)

// AccountServiceInterface is the interface for the service.
type AccountServiceInterface interface {
	Login(ctx context.Context, body *structs.LoginBody) (*AuthResponse, error)
	Register(ctx context.Context, body *structs.RegisterBody) (*AuthResponse, error)
	GetMe(ctx context.Context) (*structs.AccountMeshes, error)
	UpdatePassword(ctx context.Context, body *userStructs.UserPassword) error
	Tenant(ctx context.Context) (*tenantStructs.ReadTenant, error)
	Tenants(ctx context.Context) (paging.Result[*tenantStructs.ReadTenant], error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
}

// accountService is the struct for the service.
type accountService struct {
	d   *data.Data
	jtm *jwt.TokenManager
	ep  event.PublisherInterface

	cas CodeAuthServiceInterface
	ats AuthTenantServiceInterface
	ss  SessionServiceInterface

	usw  *wrapper.UserServiceWrapper
	tsw  *wrapper.TenantServiceWrapper
	asw  *wrapper.AccessServiceWrapper
	ugsw *wrapper.SpaceServiceWrapper
}

// NewAccountService creates a new service.
func NewAccountService(d *data.Data, jtm *jwt.TokenManager, ep event.PublisherInterface, cas CodeAuthServiceInterface, ats AuthTenantServiceInterface, ss SessionServiceInterface,
	usw *wrapper.UserServiceWrapper,
	tsw *wrapper.TenantServiceWrapper,
	asw *wrapper.AccessServiceWrapper,
	ugsw *wrapper.SpaceServiceWrapper,
) AccountServiceInterface {
	return &accountService{
		d:    d,
		jtm:  jtm,
		ep:   ep,
		cas:  cas,
		ats:  ats,
		ss:   ss,
		usw:  usw,
		tsw:  tsw,
		asw:  asw,
		ugsw: ugsw,
	}
}

// Login handles user login authentication
func (s *accountService) Login(ctx context.Context, body *structs.LoginBody) (*AuthResponse, error) {
	client := s.d.GetMasterEntClient()

	// Verify user credentials
	user, err := s.usw.FindUser(ctx, &userStructs.FindUser{Username: body.Username})
	if err = handleEntError(ctx, "User", err); err != nil {
		return nil, err
	}

	if user.Status != 0 {
		return nil, errors.New("account disabled, contact administrator")
	}

	// Verify password
	verifyResult := s.usw.VerifyPassword(ctx, user.ID, body.Password)
	switch v := verifyResult.(type) {
	case userService.VerifyPasswordResult:
		if !v.Valid {
			return nil, errors.New(v.Error)
		} else if v.Valid && v.NeedsPasswordSet {
			if validator.IsEmpty(user.Email) {
				return nil, errors.New("password not set and email empty, contact administrator")
			}
			// Send password reset code instead of login
			codeResult, err := s.cas.SendCode(ctx, &structs.SendCodeBody{Email: user.Email})
			if err != nil {
				return nil, err
			}

			return &AuthResponse{
				Registered: (*codeResult)["registered"].(bool),
			}, nil
		}
	case error:
		return nil, v
	}

	// Get user tenants
	userTenants, _ := s.tsw.GetUserTenants(ctx, user.ID)
	var tenantIDs []string
	for _, t := range userTenants {
		tenantIDs = append(tenantIDs, t.ID)
	}

	// Set default tenant context
	defaultTenant, err := s.tsw.GetUserTenant(ctx, user.ID)
	if err == nil && defaultTenant != nil {
		ctx = ctxutil.SetTenantID(ctx, defaultTenant.ID)
	}

	// Create token payload
	payload, err := CreateUserTokenPayload(ctx, s.asw, user, tenantIDs)
	if err != nil {
		return nil, err
	}

	// Generate authentication response
	authResp, err := generateAuthResponse(ctx, s.jtm, client, payload, s.ss, "password")
	if err != nil {
		return nil, err
	}

	// Set additional response data
	authResp.TenantIDs = tenantIDs
	if defaultTenant != nil {
		authResp.DefaultTenant = &types.JSON{
			"id":   defaultTenant.ID,
			"name": defaultTenant.Name,
		}
	}

	// Publish login event
	if s.ep != nil {
		ip, userAgent, sessionID := ctxutil.GetClientInfo(ctx)
		uaInfo := ctxutil.GetParsedUserAgent(ctx)

		metadata := &types.JSON{
			"ip_address":   ip,
			"user_agent":   userAgent,
			"session_id":   sessionID,
			"login_method": "password",
			"browser":      uaInfo.Browser,
			"os":           uaInfo.OS,
			"mobile":       uaInfo.Mobile,
			"referer":      ctxutil.GetReferer(ctx),
			"timestamp":    time.Now().UnixMilli(),
		}

		s.ep.PublishUserLogin(ctx, user.ID, metadata)
	}

	return authResp, nil
}

// RefreshToken refreshes access token using refresh token
func (s *accountService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	client := s.d.GetMasterEntClient()

	// Verify refresh token
	payload, err := s.jtm.DecodeToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

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
	user, err := s.usw.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Get user tenants
	userTenants, _ := s.tsw.GetUserTenants(ctx, user.ID)
	var tenantIDs []string
	for _, t := range userTenants {
		tenantIDs = append(tenantIDs, t.ID)
	}

	// Set default tenant context
	defaultTenant, err := s.tsw.GetUserTenant(ctx, user.ID)
	if err == nil && defaultTenant != nil {
		ctx = ctxutil.SetTenantID(ctx, defaultTenant.ID)
	}

	// Create token payload
	tokenPayload, err := CreateUserTokenPayload(ctx, s.asw, user, tenantIDs)
	if err != nil {
		return nil, err
	}

	// Generate new authentication response
	authResp, err := generateAuthResponse(ctx, s.jtm, client, tokenPayload, s.ss, "token_refresh")
	if err != nil {
		return nil, err
	}

	// Set additional response data
	authResp.TenantIDs = tenantIDs
	if defaultTenant != nil {
		authResp.DefaultTenant = &types.JSON{
			"id":   defaultTenant.ID,
			"name": defaultTenant.Name,
		}
	}

	// Publish token refresh event
	if s.ep != nil {
		ip, userAgent, _ := ctxutil.GetClientInfo(ctx)

		metadata := &types.JSON{
			"ip_address": ip,
			"user_agent": userAgent,
			"session_id": authResp.SessionID,
			"timestamp":  time.Now().UnixMilli(),
		}

		s.ep.PublishTokenRefreshed(ctx, user.ID, metadata)
	}

	return authResp, nil
}

// Register handles user registration
func (s *accountService) Register(ctx context.Context, body *structs.RegisterBody) (*AuthResponse, error) {
	client := s.d.GetMasterEntClient()

	// Decode register token
	payload, err := decodeRegisterToken(s.jtm, body.RegisterToken)
	if err != nil {
		return nil, errors.New("register token decode failed")
	}

	// Check user existence
	existedUser, err := s.usw.FindUser(ctx, &userStructs.FindUser{
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

	// Disable verification code
	if err = disableCodeAuth(ctx, client, payload["id"].(string)); err != nil {
		return nil, err
	}

	// Create user and profile in transaction
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

	// Create tenant if needed
	tenant, err := s.ats.IsCreateTenant(ctx, &tenantStructs.CreateTenantBody{
		TenantBody: tenantStructs.TenantBody{
			Name:      body.Tenant,
			CreatedBy: &user.ID,
			UpdatedBy: &user.ID,
		},
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
		ctx = ctxutil.SetTenantID(ctx, tenant.ID)
	}

	// Create token payload
	tokenPayload, err := CreateUserTokenPayload(ctx, s.asw, user, tenantIDs)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, err
	}

	// Generate authentication response
	authResp, err := generateAuthResponse(ctx, s.jtm, client, tokenPayload, s.ss, "registration")
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Set additional response data
	authResp.TenantIDs = tenantIDs
	if tenant != nil {
		authResp.DefaultTenant = &types.JSON{
			"id":   tenant.ID,
			"name": tenant.Name,
		}
	}

	// Publish user creation event
	if s.ep != nil {
		ip, userAgent, _ := ctxutil.GetClientInfo(ctx)
		uaInfo := ctxutil.GetParsedUserAgent(ctx)

		metadata := &types.JSON{
			"ip_address":          ip,
			"user_agent":          userAgent,
			"registration_method": "email_code",
			"session_id":          authResp.SessionID,
			"browser":             uaInfo.Browser,
			"os":                  uaInfo.OS,
			"mobile":              uaInfo.Mobile,
			"timestamp":           time.Now().UnixMilli(),
		}

		if tenant != nil {
			(*metadata)["tenant_id"] = tenant.ID
			(*metadata)["tenant_name"] = tenant.Name
		}

		s.ep.PublishUserCreated(ctx, user.ID, metadata)
	}

	return authResp, nil
}

// GetMe returns current user information
func (s *accountService) GetMe(ctx context.Context) (*structs.AccountMeshes, error) {
	user, err := s.usw.GetUserByID(ctx, ctxutil.GetUserID(ctx))
	if err != nil {
		return nil, err
	}

	return s.Serialize(user, &serializeUserParams{
		WithProfile:     true,
		WithRoles:       true,
		WithTenants:     true,
		WithGroups:      true,
		WithPermissions: true,
	}), nil
}

// UpdatePassword updates user password
func (s *accountService) UpdatePassword(ctx context.Context, body *userStructs.UserPassword) error {
	body.User = ctxutil.GetUserID(ctx)
	err := s.usw.UpdatePassword(ctx, body)

	if err == nil && s.ep != nil {
		ip, userAgent, _ := ctxutil.GetClientInfo(ctx)

		metadata := &types.JSON{
			"ip_address": ip,
			"user_agent": userAgent,
			"timestamp":  time.Now().UnixMilli(),
		}

		s.ep.PublishPasswordChanged(ctx, body.User, metadata)
	}

	return err
}

// Tenant returns user's default tenant
func (s *accountService) Tenant(ctx context.Context) (*tenantStructs.ReadTenant, error) {
	userID := ctxutil.GetUserID(ctx)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	row, err := s.tsw.GetTenantByUser(ctx, userID)
	if err = handleEntError(ctx, "Tenant", err); err != nil {
		return nil, err
	}

	return row, nil
}

// Tenants returns user's all tenants
func (s *accountService) Tenants(ctx context.Context) (paging.Result[*tenantStructs.ReadTenant], error) {
	userID := ctxutil.GetUserID(ctx)
	if userID == "" {
		return paging.Result[*tenantStructs.ReadTenant]{}, errors.New("invalid user ID")
	}

	rows, err := s.tsw.ListTenants(ctx, &tenantStructs.ListTenantParams{
		User: userID,
	})
	if err = handleEntError(ctx, "Tenants", err); err != nil {
		return paging.Result[*tenantStructs.ReadTenant]{}, err
	}

	return rows, nil
}

func decodeRegisterToken(jtm *jwt.TokenManager, token string) (types.JSON, error) {
	decoded, err := jtm.DecodeToken(token)
	if err != nil {
		return nil, err
	}
	if decoded["sub"] != "email-register" {
		return nil, fmt.Errorf("invalid authorize information")
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
	user, err := svc.usw.CreateUser(ctx, &userStructs.UserBody{
		Username: body.Username,
		Email:    payload["email"].(string),
		Phone:    body.Phone,
	})
	if err != nil {
		return nil, err
	}

	userProfile, err := svc.usw.CreateUserProfile(ctx, &userStructs.UserProfileBody{
		UserID:      user.ID,
		DisplayName: body.DisplayName,
		ShortBio:    body.ShortBio,
	})
	if err != nil {
		return nil, err
	}

	return types.JSON{"user": user, "profile": userProfile}, nil
}

// SerializeParams serialize params
type serializeUserParams struct {
	WithProfile     bool
	WithRoles       bool
	WithTenants     bool
	WithGroups      bool
	WithPermissions bool
}

// Serialize serializes user
func (s *accountService) Serialize(user *userStructs.ReadUser, sp ...*serializeUserParams) *structs.AccountMeshes {
	ctx := context.Background()
	um := &structs.AccountMeshes{
		User: user,
	}

	params := &serializeUserParams{}
	if len(sp) > 0 {
		params = sp[0]
	}

	// Get user profile
	if params.WithProfile {
		if profile, _ := s.usw.GetUserProfile(ctx, user.ID); profile != nil {
			um.Profile = profile
		}
	}

	// Get user tenants
	if params.WithTenants {
		if tenants, _ := s.tsw.GetUserTenants(ctx, user.ID); len(tenants) > 0 {
			um.Tenants = tenants
		}
	}

	// Get roles and permissions together for efficiency
	if params.WithRoles || params.WithPermissions {
		roleSlugs, permissions, isAdmin, tenantID := s.getUserRolesAndPermissions(ctx, user.ID)

		if params.WithRoles {
			um.Roles = roleSlugs
		}

		if params.WithPermissions {
			um.Permissions = permissions
			um.IsAdmin = isAdmin
			um.TenantID = tenantID
		}
	}

	// Get user groups
	if params.WithGroups {
		groups, _ := s.ugsw.GetUserGroups(ctx, user.ID)
		um.Groups = groups
	}

	return um
}

// getUserRolesAndPermissions gets user roles and permissions efficiently
func (s *accountService) getUserRolesAndPermissions(ctx context.Context, userID string) ([]string, []string, bool, string) {
	// Get tenant context
	tenantID := ctxutil.GetTenantID(ctx)
	if tenantID == "" {
		// Try to get default tenant for user
		if defaultTenant, err := s.tsw.GetUserTenant(ctx, userID); err == nil && defaultTenant != nil {
			tenantID = defaultTenant.ID
			ctx = ctxutil.SetTenantID(ctx, tenantID)
		}
	}

	// Use existing helper function to get comprehensive role and permission data
	finalTenantID, roleSlugs, permissionCodes, isAdmin, err := GetUserTenantsRolesPermissions(
		ctx, s.asw, userID,
	)

	if err != nil {
		// Fallback: try to get basic role information
		roleSlugs = s.getFallbackRoles(ctx, userID)
		permissionCodes = []string{}
		isAdmin = ctxutil.GetUserIsAdmin(ctx)
		finalTenantID = tenantID
	}

	return roleSlugs, permissionCodes, isAdmin, finalTenantID
}

// getFallbackRoles gets basic role information when main method fails
func (s *accountService) getFallbackRoles(ctx context.Context, userID string) []string {
	var roleSlugs []string

	// Try global roles
	if globalRoles, err := s.asw.GetUserRoles(ctx, userID); err == nil {
		for _, role := range globalRoles {
			roleSlugs = append(roleSlugs, role.Slug)
		}
	}

	// Try tenant-specific roles if tenant context exists
	tenantID := ctxutil.GetTenantID(ctx)
	if tenantID != "" {
		if roleIDs, err := s.asw.GetUserRolesInTenant(ctx, userID, tenantID); err == nil && len(roleIDs) > 0 {
			if tenantRoles, err := s.asw.GetByIDs(ctx, roleIDs); err == nil {
				for _, role := range tenantRoles {
					// Avoid duplicates
					found := false
					for _, existing := range roleSlugs {
						if existing == role.Slug {
							found = true
							break
						}
					}
					if !found {
						roleSlugs = append(roleSlugs, role.Slug)
					}
				}
			}
		}
	}

	return roleSlugs
}
