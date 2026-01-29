package service

import (
	"context"
	"errors"
	"fmt"
	accessStructs "ncobase/core/access/structs"
	"ncobase/core/auth/data/repository"
	"ncobase/core/auth/structs"
	"ncobase/core/auth/wrapper"
	userStructs "ncobase/core/user/structs"
	"net/http"
	"strings"
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
	Registered   bool        `json:"registered,omitempty"`
	AccessToken  string      `json:"access_token,omitempty"`
	RefreshToken string      `json:"refresh_token,omitempty"`
	SessionID    string      `json:"session_id,omitempty"`
	TokenType    string      `json:"token_type,omitempty"`
	ExpiresIn    int64       `json:"expires_in,omitempty"`
	SpaceIDs     []string    `json:"space_ids,omitempty"`
	DefaultSpace *types.JSON `json:"default_space,omitempty"`
	MFARequired  bool        `json:"mfa_required,omitempty"`
	MFAToken     string      `json:"mfa_token,omitempty"`
	MFAMethods   []string    `json:"mfa_methods,omitempty"`
}

// GetUserSpacesRolesPermissions retrieves user roles and permissions
func GetUserSpacesRolesPermissions(
	ctx context.Context,
	userID string,
	asw *wrapper.AccessServiceWrapper,
	tsw *wrapper.SpaceServiceWrapper,
) (spaceID string, roleSlugs []string, permissionCodes []string, isAdmin bool, err error) {
	spaceID = ctxutil.GetSpaceID(ctx)
	logger.Debugf(ctx, "Getting permissions for user %s, space %s", userID, spaceID)

	// Ensure we have a space context (domain) for space-specific roles and Casbin checks.
	// Login flow often has no space_id in context yet.
	if spaceID == "" {
		if defaultSpace, spaceErr := tsw.GetUserSpace(ctx, userID); spaceErr == nil && defaultSpace != nil {
			spaceID = defaultSpace.ID
			ctx = ctxutil.SetSpaceID(ctx, spaceID)
		}
	}

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

	// Get space-specific roles if space context exists
	if spaceID != "" {
		roleIDs, roleErr := tsw.GetUserRolesInSpace(ctx, userID, spaceID)
		if roleErr == nil && len(roleIDs) > 0 {
			spaceRoles, _ := asw.GetByIDs(ctx, roleIDs)
			for _, role := range spaceRoles {
				if !utils.Contains(roleSlugs, role.Slug) {
					roleSlugs = append(roleSlugs, role.Slug)
				}
			}
			logger.Debugf(ctx, "Found %d space roles for user", len(spaceRoles))
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
	adminRoles := []string{"super-admin", "system-admin", "enterprise-admin", "space-admin"}
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

			permissionCodes = append(permissionCodes, buildPermissionCodes(perm)...)
		}

		logger.Debugf(ctx, "Role %s has %d permissions", role.Slug, len(rolePermissions))
	}

	return permissionCodes, nil
}

// CreateUserTokenPayload creates token payload with user permissions
func CreateUserTokenPayload(
	ctx context.Context,
	user *userStructs.ReadUser,
	spaceIDs []string,
	asw *wrapper.AccessServiceWrapper,
	tsw *wrapper.SpaceServiceWrapper,
) (types.JSON, error) {
	if user.ID == "" {
		return nil, errors.New("userID is required")
	}

	spaceID, roleSlugs, permissionCodes, isAdmin, err := GetUserSpacesRolesPermissions(ctx, user.ID, asw, tsw)
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
		"space_id":     spaceID,
		"space_ids":    spaceIDs,
		"roles":        roleSlugs,
		"permissions":  permissionCodes,
		"user_status":  user.Status,
		"is_certified": user.IsCertified,
	}, nil
}

func buildPermissionCodes(perm *accessStructs.ReadPermission) []string {
	if perm == nil {
		return nil
	}

	codes := make(map[string]struct{})
	add := func(code string) {
		if code == "" {
			return
		}
		codes[code] = struct{}{}
	}

	if perm.Action != "" && perm.Subject != "" {
		add(fmt.Sprintf("%s:%s", perm.Action, perm.Subject))
	}

	if perm.Action == "*" && perm.Subject == "*" {
		add("*:*")
		return mapKeys(codes)
	}

	if looksLikeHTTPPermission(perm.Action, perm.Subject) {
		semanticAction := mapHTTPAction(perm.Action)
		subject := extractSubjectFromPath(perm.Subject)
		if semanticAction != "" && subject != "" {
			addPermissionVariants(codes, semanticAction, subject)
		}
	} else if looksLikeSemanticPermission(perm.Action, perm.Subject) {
		addPermissionVariants(codes, strings.ToLower(perm.Action), strings.ToLower(perm.Subject))
	}

	return mapKeys(codes)
}

func addPermissionVariants(codes map[string]struct{}, action, subject string) {
	if action == "" || subject == "" {
		return
	}

	add := func(code string) {
		if code == "" {
			return
		}
		codes[code] = struct{}{}
	}

	base := strings.ToLower(subject)
	act := strings.ToLower(action)
	add(fmt.Sprintf("%s:%s", act, base))

	if singular := singularize(base); singular != base {
		add(fmt.Sprintf("%s:%s", act, singular))
	}
	if plural := pluralize(base); plural != base {
		add(fmt.Sprintf("%s:%s", act, plural))
	}
}

func looksLikeHTTPPermission(action, subject string) bool {
	if subject == "" {
		return false
	}
	if strings.HasPrefix(subject, "/") {
		return true
	}
	upper := strings.ToUpper(action)
	if upper == "*" {
		return strings.HasPrefix(subject, "/")
	}
	return upper == "GET" || upper == "POST" || upper == "PUT" || upper == "PATCH" || upper == "DELETE" || upper == "HEAD" || upper == "OPTIONS"
}

func looksLikeSemanticPermission(action, subject string) bool {
	if subject == "" || strings.HasPrefix(subject, "/") {
		return false
	}
	switch strings.ToLower(action) {
	case "read", "create", "update", "delete", "manage", "*":
		return true
	default:
		return false
	}
}

func mapHTTPAction(action string) string {
	switch strings.ToUpper(action) {
	case "GET", "HEAD", "OPTIONS":
		return "read"
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	case "*":
		return "*"
	default:
		return strings.ToLower(action)
	}
}

func extractSubjectFromPath(path string) string {
	if path == "" {
		return ""
	}

	clean := strings.Split(path, "?")[0]
	clean = strings.TrimSuffix(clean, "*")
	clean = strings.TrimSuffix(clean, "/")
	clean = strings.TrimSpace(clean)

	segments := strings.FieldsFunc(clean, func(r rune) bool { return r == '/' })
	if len(segments) == 0 {
		return ""
	}

	idx := 0
	if segments[0] == "api" {
		idx = 1
	}
	if idx < len(segments) && isVersionSegment(segments[idx]) {
		idx++
	}
	if idx >= len(segments) {
		return ""
	}

	segment := strings.Trim(segments[idx], "{}")
	segment = strings.TrimSuffix(segment, "*")
	return strings.ToLower(segment)
}

func isVersionSegment(segment string) bool {
	if len(segment) < 2 || segment[0] != 'v' {
		return false
	}
	for i := 1; i < len(segment); i++ {
		if segment[i] < '0' || segment[i] > '9' {
			return false
		}
	}
	return true
}

func singularize(word string) string {
	if strings.HasSuffix(word, "ies") && len(word) > 3 {
		return word[:len(word)-3] + "y"
	}
	if strings.HasSuffix(word, "ches") || strings.HasSuffix(word, "shes") || strings.HasSuffix(word, "xes") || strings.HasSuffix(word, "zes") {
		return word[:len(word)-2]
	}
	if strings.HasSuffix(word, "ses") && len(word) > 3 {
		return word[:len(word)-2]
	}
	if strings.HasSuffix(word, "s") && len(word) > 1 {
		return word[:len(word)-1]
	}
	return word
}

func pluralize(word string) string {
	if strings.HasSuffix(word, "y") && len(word) > 1 {
		prev := word[len(word)-2]
		if !strings.ContainsRune("aeiou", rune(prev)) {
			return word[:len(word)-1] + "ies"
		}
	}
	if strings.HasSuffix(word, "s") || strings.HasSuffix(word, "x") || strings.HasSuffix(word, "z") || strings.HasSuffix(word, "ch") || strings.HasSuffix(word, "sh") {
		return word + "es"
	}
	return word + "s"
}

func mapKeys(codes map[string]struct{}) []string {
	if len(codes) == 0 {
		return nil
	}
	result := make([]string, 0, len(codes))
	for code := range codes {
		result = append(result, code)
	}
	return result
}

// generateUserTokens generates access and refresh tokens for API authentication
func generateUserTokens(jtm *jwt.TokenManager, payload map[string]any, tokenID string) (string, string) {
	userID, ok := payload["user_id"].(string)
	if !ok || userID == "" {
		return "", ""
	}

	// Generate access token (shorter expiry for security)
	accessToken, _ := jtm.GenerateAccessToken(tokenID, payload)

	// Generate refresh token (longer expiry)
	refreshToken, _ := jtm.GenerateRefreshToken(tokenID, types.JSON{
		"user_id": userID,
	})

	return accessToken, refreshToken
}

// generateAuthResponse generates authentication response with tokens and session
func generateAuthResponse(
	ctx context.Context,
	jtm *jwt.TokenManager,
	authTokenRepo repository.AuthTokenRepositoryInterface,
	payload map[string]any,
	sessionSvc SessionServiceInterface,
	loginMethod string,
) (*AuthResponse, error) {
	userID, ok := payload["user_id"].(string)
	if !ok || userID == "" {
		return nil, errors.New("user_id not found in payload")
	}

	if authTokenRepo == nil {
		return nil, errors.New("auth token repository not configured")
	}

	authToken, err := authTokenRepo.Create(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth token: %w", err)
	}

	// Generate tokens for API authentication
	accessToken, refreshToken := generateUserTokens(jtm, payload, authToken.ID)
	if accessToken == "" || refreshToken == "" {
		return nil, errors.New("failed to generate tokens")
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
	if repository.IsNotFound(err) {
		logger.Errorf(ctx, "Error not found in %s: %v", k, err)
		return errors.New(ecode.NotExist(k))
	}
	if repository.IsConstraintError(err) {
		logger.Errorf(ctx, "Error constraint in %s: %v", k, err)
		return errors.New(ecode.AlreadyExist(k))
	}
	if repository.IsNotSingular(err) {
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
func sendRegisterMail(_ context.Context, jtm *jwt.TokenManager, email, id string) (*types.JSON, error) {
	subject := "email-register"
	payload := types.JSON{"email": email, "id": id}
	registerToken, err := jtm.GenerateRegisterToken(id, payload, subject)
	if err != nil {
		return nil, err
	}
	return &types.JSON{"email": email, "register_token": registerToken}, nil
}
