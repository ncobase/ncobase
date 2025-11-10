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
	spaceStructs "ncobase/space/structs"
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
	Space(ctx context.Context) (*spaceStructs.ReadSpace, error)
	Spaces(ctx context.Context) (paging.Result[*spaceStructs.ReadSpace], error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
}

// accountService is the struct for the service.
type accountService struct {
	d   *data.Data
	jtm *jwt.TokenManager
	ep  event.PublisherInterface

	cas CodeAuthServiceInterface
	ats AuthSpaceServiceInterface
	ss  SessionServiceInterface

	usw  *wrapper.UserServiceWrapper
	tsw  *wrapper.SpaceServiceWrapper
	asw  *wrapper.AccessServiceWrapper
	ugsw *wrapper.OrganizationServiceWrapper
}

// NewAccountService creates a new service.
func NewAccountService(d *data.Data, jtm *jwt.TokenManager, ep event.PublisherInterface, cas CodeAuthServiceInterface, ats AuthSpaceServiceInterface, ss SessionServiceInterface,
	usw *wrapper.UserServiceWrapper,
	tsw *wrapper.SpaceServiceWrapper,
	asw *wrapper.AccessServiceWrapper,
	ugsw *wrapper.OrganizationServiceWrapper,
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

			registered, err := safeGetBool(*codeResult, "registered")
			if err != nil {
				return nil, fmt.Errorf("failed to extract registered field: %w", err)
			}

			return &AuthResponse{
				Registered: registered,
			}, nil
		}
	case error:
		return nil, v
	}

	// Get user spaces
	userSpaces, _ := s.tsw.GetUserSpaces(ctx, user.ID)
	var spaceIDs []string
	for _, t := range userSpaces {
		spaceIDs = append(spaceIDs, t.ID)
	}

	// Set default space context
	defaultSpace, err := s.tsw.GetUserSpace(ctx, user.ID)
	if err == nil && defaultSpace != nil {
		ctx = ctxutil.SetSpaceID(ctx, defaultSpace.ID)
	}

	// Create token payload
	payload, err := CreateUserTokenPayload(ctx, user, spaceIDs, s.asw, s.tsw)
	if err != nil {
		return nil, err
	}

	// Generate authentication response
	authResp, err := generateAuthResponse(ctx, s.jtm, client, payload, s.ss, "password")
	if err != nil {
		return nil, err
	}

	// Set additional response data
	authResp.SpaceIDs = spaceIDs
	if defaultSpace != nil {
		authResp.DefaultSpace = &types.JSON{
			"id":   defaultSpace.ID,
			"name": defaultSpace.Name,
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

	// Get user spaces
	userSpaces, _ := s.tsw.GetUserSpaces(ctx, user.ID)
	var spaceIDs []string
	for _, t := range userSpaces {
		spaceIDs = append(spaceIDs, t.ID)
	}

	// Set default space context
	defaultSpace, err := s.tsw.GetUserSpace(ctx, user.ID)
	if err == nil && defaultSpace != nil {
		ctx = ctxutil.SetSpaceID(ctx, defaultSpace.ID)
	}

	// Create token payload
	tokenPayload, err := CreateUserTokenPayload(ctx, user, spaceIDs, s.asw, s.tsw)
	if err != nil {
		return nil, err
	}

	// Generate new authentication response
	authResp, err := generateAuthResponse(ctx, s.jtm, client, tokenPayload, s.ss, "token_refresh")
	if err != nil {
		return nil, err
	}

	// Set additional response data
	authResp.SpaceIDs = spaceIDs
	if defaultSpace != nil {
		authResp.DefaultSpace = &types.JSON{
			"id":   defaultSpace.ID,
			"name": defaultSpace.Name,
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

	// Extract email from payload with type checking
	email, err := safeGetString(payload, "email")
	if err != nil {
		return nil, fmt.Errorf("invalid register token payload: %w", err)
	}

	// Extract code auth ID from payload with type checking
	codeAuthID, err := safeGetString(payload, "id")
	if err != nil {
		return nil, fmt.Errorf("invalid register token payload: %w", err)
	}

	// Check user existence
	existedUser, err := s.usw.FindUser(ctx, &userStructs.FindUser{
		Username: body.Username,
		Email:    email,
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
	if err = disableCodeAuth(ctx, client, codeAuthID); err != nil {
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

	// Create space if needed
	space, err := s.ats.IsCreateSpace(ctx, &spaceStructs.CreateSpaceBody{
		SpaceBody: spaceStructs.SpaceBody{
			Name:      body.Space,
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

	// Get space IDs
	var spaceIDs []string
	if space != nil {
		spaceIDs = append(spaceIDs, space.ID)
		ctx = ctxutil.SetSpaceID(ctx, space.ID)
	}

	// Create token payload
	tokenPayload, err := CreateUserTokenPayload(ctx, user, spaceIDs, s.asw, s.tsw)
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
	authResp.SpaceIDs = spaceIDs
	if space != nil {
		authResp.DefaultSpace = &types.JSON{
			"id":   space.ID,
			"name": space.Name,
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

		if space != nil {
			(*metadata)["space_id"] = space.ID
			(*metadata)["space_name"] = space.Name
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
		WithSpaces:      true,
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

// Space returns user's default space
func (s *accountService) Space(ctx context.Context) (*spaceStructs.ReadSpace, error) {
	userID := ctxutil.GetUserID(ctx)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	row, err := s.tsw.GetSpaceByUser(ctx, userID)
	if err = handleEntError(ctx, "Space", err); err != nil {
		return nil, err
	}

	return row, nil
}

// Spaces returns user's all spaces
func (s *accountService) Spaces(ctx context.Context) (paging.Result[*spaceStructs.ReadSpace], error) {
	userID := ctxutil.GetUserID(ctx)
	if userID == "" {
		return paging.Result[*spaceStructs.ReadSpace]{}, errors.New("invalid user ID")
	}

	rows, err := s.tsw.ListSpaces(ctx, &spaceStructs.ListSpaceParams{
		User: userID,
	})
	if err = handleEntError(ctx, "Spaces", err); err != nil {
		return paging.Result[*spaceStructs.ReadSpace]{}, err
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
	// Extract email from payload with type checking
	email, err := safeGetString(payload, "email")
	if err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	user, err := svc.usw.CreateUser(ctx, &userStructs.UserBody{
		Username: body.Username,
		Email:    email,
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
	WithSpaces      bool
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

	// Get user spaces
	if params.WithSpaces {
		if spaces, _ := s.tsw.GetUserSpaces(ctx, user.ID); len(spaces) > 0 {
			um.Spaces = spaces
		}
	}

	// Get roles and permissions together for efficiency
	if params.WithRoles || params.WithPermissions {
		roleSlugs, permissions, isAdmin, spaceID := s.getUserRolesAndPermissions(ctx, user.ID)

		if params.WithRoles {
			um.Roles = roleSlugs
		}

		if params.WithPermissions {
			um.Permissions = permissions
			um.IsAdmin = isAdmin
			um.SpaceID = spaceID
		}
	}

	// Get user orgs
	if params.WithGroups {
		orgs, _ := s.ugsw.GetUserGroups(ctx, user.ID)
		um.Groups = orgs
	}

	return um
}

// getUserRolesAndPermissions gets user roles and permissions efficiently
func (s *accountService) getUserRolesAndPermissions(ctx context.Context, userID string) ([]string, []string, bool, string) {
	// Get space context
	spaceID := ctxutil.GetSpaceID(ctx)
	if spaceID == "" {
		// Try to get default space for user
		if defaultSpace, err := s.tsw.GetUserSpace(ctx, userID); err == nil && defaultSpace != nil {
			spaceID = defaultSpace.ID
			ctx = ctxutil.SetSpaceID(ctx, spaceID)
		}
	}

	// Use existing helper function to get comprehensive role and permission data
	finalSpaceID, roleSlugs, permissionCodes, isAdmin, err := GetUserSpacesRolesPermissions(ctx, userID, s.asw, s.tsw)

	if err != nil {
		// Fallback: try to get basic role information
		roleSlugs = s.getFallbackRoles(ctx, userID)
		permissionCodes = []string{}
		isAdmin = ctxutil.GetUserIsAdmin(ctx)
		finalSpaceID = spaceID
	}

	return roleSlugs, permissionCodes, isAdmin, finalSpaceID
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

	// Try space-specific roles if space context exists
	spaceID := ctxutil.GetSpaceID(ctx)
	if spaceID != "" {
		if roleIDs, err := s.tsw.GetUserRolesInSpace(ctx, userID, spaceID); err == nil && len(roleIDs) > 0 {
			if spaceRoles, err := s.asw.GetByIDs(ctx, roleIDs); err == nil {
				for _, role := range spaceRoles {
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

// safeGetString safely extracts a string value from a map with type checking
func safeGetString(data types.JSON, key string) (string, error) {
	val, exists := data[key]
	if !exists {
		return "", fmt.Errorf("missing required field: %s", key)
	}

	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("field %s is not a string (got %T)", key, val)
	}

	if str == "" {
		return "", fmt.Errorf("field %s cannot be empty", key)
	}

	return str, nil
}

// safeGetBool safely extracts a boolean value from a map with type checking
func safeGetBool(data types.JSON, key string) (bool, error) {
	val, exists := data[key]
	if !exists {
		return false, fmt.Errorf("missing required field: %s", key)
	}

	b, ok := val.(bool)
	if !ok {
		return false, fmt.Errorf("field %s is not a boolean (got %T)", key, val)
	}

	return b, nil
}
