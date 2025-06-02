package service

import (
	"context"
	"errors"
	"fmt"
	accessStructs "ncobase/access/structs"
	"ncobase/auth/data/ent"
	"ncobase/auth/structs"
	"ncobase/auth/wrapper"
	userStructs "ncobase/user/structs"
	"net/http"
	"time"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/messaging/email"
	"github.com/ncobase/ncore/net/cookie"
	"github.com/ncobase/ncore/security/jwt"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils"
	"github.com/ncobase/ncore/validation/validator"
)

// AuthResponse represents authentication response
type AuthResponse struct {
	Registered    bool        `json:"registered,omitempty"`
	AccessToken   string      `json:"access_token,omitempty"`
	RefreshToken  string      `json:"refresh_token,omitempty"`
	SessionID     string      `json:"session_id,omitempty"`
	TokenType     string      `json:"token_type,omitempty"`
	ExpiresIn     int64       `json:"expires_in,omitempty"`
	TenantIDs     []string    `json:"tenant_ids,omitempty"`
	DefaultTenant *types.JSON `json:"default_tenant,omitempty"`
}

// GetUserTenantsRolesPermissions retrieves user roles and permissions
func GetUserTenantsRolesPermissions(
	ctx context.Context,
	userID string,
	asw *wrapper.AccessServiceWrapper,
	tsw *wrapper.TenantServiceWrapper,
) (tenantID string, roleSlugs []string, permissionCodes []string, isAdmin bool, err error) {
	tenantID = ctxutil.GetTenantID(ctx)
	logger.Debugf(ctx, "Getting permissions for user %s, tenant %s", userID, tenantID)

	// Get global roles first
	globalRoles, err := asw.GetUserRoles(ctx, userID)
	if err != nil {
		logger.Warnf(ctx, "Failed to get global roles for user %s: %v", userID, err)
	} else {
		for _, role := range globalRoles {
			roleSlugs = append(roleSlugs, role.Slug)
		}
		logger.Debugf(ctx, "Found %d global roles for user", len(globalRoles))
	}

	// Get tenant-specific roles if tenant context exists
	if tenantID != "" {
		roleIDs, roleErr := tsw.GetUserRolesInTenant(ctx, userID, tenantID)
		if roleErr == nil && len(roleIDs) > 0 {
			tenantRoles, _ := asw.GetByIDs(ctx, roleIDs)
			for _, role := range tenantRoles {
				if !utils.Contains(roleSlugs, role.Slug) {
					roleSlugs = append(roleSlugs, role.Slug)
				}
			}
			logger.Debugf(ctx, "Found %d tenant roles for user", len(tenantRoles))
		}
	}

	isAdmin = isAdminRole(roleSlugs)
	if isAdmin {
		logger.Debugf(ctx, "User has admin privileges")
	}

	permissionCodes, err = getPermissionsForRoles(ctx, asw, globalRoles, isAdmin)
	if err != nil {
		logger.Warnf(ctx, "Failed to get permissions: %v", err)
	}

	roleSlugs = utils.RemoveDuplicates(roleSlugs)
	permissionCodes = utils.RemoveDuplicates(permissionCodes)

	logger.Infof(ctx, "User %s has %d roles, %d permissions, isAdmin: %v",
		userID, len(roleSlugs), len(permissionCodes), isAdmin)
	return
}

// isAdminRole checks if any role has admin privileges
func isAdminRole(roleSlugs []string) bool {
	adminRoles := []string{"super-admin", "system-admin", "enterprise-admin", "tenant-admin"}
	for _, roleSlug := range roleSlugs {
		for _, adminRole := range adminRoles {
			if roleSlug == adminRole {
				return true
			}
		}
	}
	return false
}

// getPermissionsForRoles gets all permissions for the given roles
func getPermissionsForRoles(ctx context.Context, asw *wrapper.AccessServiceWrapper, roles []*accessStructs.ReadRole, isAdmin bool) ([]string, error) {
	if len(roles) == 0 {
		return []string{}, nil
	}

	var permissionCodes []string

	for _, role := range roles {
		rolePermissions, err := asw.GetRolePermissions(ctx, role.ID)
		if err != nil {
			logger.Warnf(ctx, "Failed to get permissions for role %s: %v", role.Slug, err)
			continue
		}

		for _, perm := range rolePermissions {
			if perm.Action == "*" && perm.Subject == "*" {
				permissionCodes = []string{"*:*"}
				logger.Infof(ctx, "User has super admin wildcard permission")
				return permissionCodes, nil
			}

			permCode := fmt.Sprintf("%s:%s", perm.Action, perm.Subject)
			permissionCodes = append(permissionCodes, permCode)
		}

		logger.Debugf(ctx, "Role %s has %d permissions", role.Slug, len(rolePermissions))
	}

	return permissionCodes, nil
}

// CreateUserTokenPayload creates token payload with user permissions
func CreateUserTokenPayload(
	ctx context.Context,
	user *userStructs.ReadUser,
	tenantIDs []string,
	asw *wrapper.AccessServiceWrapper,
	tsw *wrapper.TenantServiceWrapper,
) (types.JSON, error) {
	if user.ID == "" {
		return nil, errors.New("userID is required")
	}

	tenantID, roleSlugs, permissionCodes, isAdmin, err := GetUserTenantsRolesPermissions(ctx, user.ID, asw, tsw)
	if err != nil {
		logger.Errorf(ctx, "Failed to get user roles and permissions: %v", err)
		roleSlugs = []string{}
		permissionCodes = []string{}
		isAdmin = false
	}

	return types.JSON{
		"user_id":      user.ID,
		"username":     user.Username,
		"email":        user.Email,
		"is_admin":     isAdmin,
		"tenant_id":    tenantID,
		"tenant_ids":   tenantIDs,
		"roles":        roleSlugs,
		"permissions":  permissionCodes,
		"user_status":  user.Status,
		"is_certified": user.IsCertified,
	}, nil
}

// generateUserTokens generates access and refresh tokens for API authentication
func generateUserTokens(jtm *jwt.TokenManager, payload map[string]any, tokenID string) (string, string) {
	userID, ok := payload["user_id"].(string)
	if !ok || userID == "" {
		return "", ""
	}

	// Generate access token (shorter expiry for security)
	accessToken, _ := jtm.GenerateAccessTokenWithExpiry(tokenID, payload, 2*time.Hour)

	// Generate refresh token (longer expiry)
	refreshToken, _ := jtm.GenerateRefreshTokenWithExpiry(tokenID, types.JSON{
		"user_id": userID,
	}, 7*24*time.Hour)

	return accessToken, refreshToken
}

// generateAuthResponse generates authentication response with tokens and session
func generateAuthResponse(
	ctx context.Context,
	jtm *jwt.TokenManager,
	client *ent.Client,
	payload map[string]any,
	sessionSvc SessionServiceInterface,
	loginMethod string,
) (*AuthResponse, error) {
	userID, ok := payload["user_id"].(string)
	if !ok || userID == "" {
		return nil, errors.New("user_id not found in payload")
	}

	// Create auth token record in transaction
	tx, err := client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				logger.Errorf(ctx, "Failed to rollback transaction: %v", rbErr)
			}
		}
	}()

	authToken, err := createAuthToken(ctx, tx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth token: %w", err)
	}

	// Generate tokens for API authentication
	accessToken, refreshToken := generateUserTokens(jtm, payload, authToken.ID)
	if accessToken == "" || refreshToken == "" {
		return nil, errors.New("failed to generate tokens")
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Create session for cookie-based authentication
	var sessionID string
	if sessionSvc != nil {
		ip, userAgent, _ := ctxutil.GetClientInfo(ctx)
		sessionBody := &structs.SessionBody{
			UserID:      userID,
			IPAddress:   ip,
			UserAgent:   userAgent,
			LoginMethod: loginMethod,
			DeviceInfo: &types.JSON{
				"browser": ctxutil.GetParsedUserAgent(ctx).Browser,
				"os":      ctxutil.GetParsedUserAgent(ctx).OS,
				"mobile":  ctxutil.GetParsedUserAgent(ctx).Mobile,
			},
		}

		session, err := sessionSvc.Create(ctx, sessionBody, authToken.ID)
		if err != nil {
			logger.Warnf(ctx, "Failed to create session: %v", err)
		} else {
			sessionID = session.ID
		}
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionID:    sessionID,
		TokenType:    "Bearer",
		ExpiresIn:    2 * 60 * 60, // 2 hours in seconds
	}, nil
}

// SetSessionCookie sets session cookie for web authentication
func SetSessionCookie(ctx context.Context, w http.ResponseWriter, r *http.Request, sessionID string) error {
	if sessionID == "" {
		return nil // No session to set
	}

	domain := r.Host
	if err := cookie.SetSessionID(w, sessionID, domain); err != nil {
		logger.Errorf(ctx, "Failed to set session cookie: %v", err)
		return err
	}

	logger.Debugf(ctx, "Session cookie set: %s", sessionID)
	return nil
}

// ClearAuthenticationCookies clears all authentication cookies
func ClearAuthenticationCookies(w http.ResponseWriter) {
	cookie.ClearAll(w)
}

// createAuthToken creates a new auth token for a user
func createAuthToken(ctx context.Context, tx *ent.Tx, userID string) (*ent.AuthToken, error) {
	return tx.AuthToken.Create().SetUserID(userID).Save(ctx)
}

// RefreshUserToken refreshes user access and refresh tokens
func RefreshUserToken(jtm *jwt.TokenManager, userID, tokenID, originalRefreshToken string, refreshTokenExp int64) (string, string) {
	now := time.Now().Unix()
	diff := refreshTokenExp - now

	refreshToken := originalRefreshToken
	accessPayload := types.JSON{
		"user_id": userID,
	}
	accessToken, _ := jtm.GenerateAccessToken(tokenID, accessPayload)

	// Refresh the refresh token if it's close to expiry (15 days remaining)
	if diff < 60*60*24*15 {
		refreshPayload := types.JSON{
			"user_id": userID,
		}
		refreshToken, _ = jtm.GenerateRefreshToken(tokenID, refreshPayload)
	}

	return accessToken, refreshToken
}

// handleEntError handles ent errors consistently
func handleEntError(ctx context.Context, k string, err error) error {
	if ent.IsNotFound(err) {
		logger.Errorf(ctx, "Error not found in %s: %v", k, err)
		return errors.New(ecode.NotExist(k))
	}
	if ent.IsConstraintError(err) {
		logger.Errorf(ctx, "Error constraint in %s: %v", k, err)
		return errors.New(ecode.AlreadyExist(k))
	}
	if ent.IsNotSingular(err) {
		logger.Errorf(ctx, "Error not singular in %s: %v", k, err)
		return errors.New(ecode.NotSingular(k))
	}
	if validator.IsNotNil(err) {
		logger.Errorf(ctx, "Error internal in %s: %v", k, err)
		return err
	}
	return err
}

// sendAuthEmail sends email with authentication code
func sendAuthEmail(ctx context.Context, e, code string, registered bool) error {
	conf := ctxutil.GetConfig(ctx)
	template := email.Template{
		Subject:  "Email authentication",
		Template: "auth-email",
		Keyword:  "Sign in",
	}
	if registered {
		template.URL = conf.Frontend.SignInURL + "?code=" + code
	} else {
		template.Keyword = "Sign Up"
		template.URL = conf.Frontend.SignUpURL + "?code=" + code
	}
	_, err := ctxutil.SendEmailWithTemplate(ctx, e, template)
	return err
}

// sendRegisterMail sends email with register token
func sendRegisterMail(_ context.Context, jtm *jwt.TokenManager, codeAuth *ent.CodeAuth) (*types.JSON, error) {
	subject := "email-register"
	payload := types.JSON{"email": codeAuth.Email, "id": codeAuth.ID}
	registerToken, err := jtm.GenerateRegisterToken(codeAuth.ID, payload, subject)
	if err != nil {
		return nil, err
	}
	return &types.JSON{"email": codeAuth.Email, "register_token": registerToken}, nil
}
