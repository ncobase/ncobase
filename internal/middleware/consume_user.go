package middleware

import (
	"context"
	"fmt"
	accessStructs "ncobase/core/access/structs"
	authStructs "ncobase/core/auth/structs"
	"strings"
	"time"

	"github.com/ncobase/ncore/consts"
	"github.com/ncobase/ncore/ctxutil"
	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/cookie"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/security/jwt"
	"github.com/ncobase/ncore/utils"

	"github.com/gin-gonic/gin"
)

// ConsumeUser middleware supports both JWT token and session cookie authentication
func ConsumeUser(em ext.ManagerInterface, whiteList []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipPath(c.Request, whiteList) {
			c.Next()
			return
		}

		// Get Service wrapper manager
		sm := GetServiceManager(em)

		// Try JWT token authentication first
		if token := extractToken(c); token != "" {
			asw := sm.AuthServiceWrapper()
			if jtm := asw.GetTokenManager(); jtm != nil {
				if handleTokenAuth(c, jtm, token) {
					c.Next()
					return
				}
			}
		}

		// Try session cookie authentication
		if sessionID, err := cookie.GetSessionID(c.Request); err == nil && sessionID != "" {
			if handleSessionAuth(c, sm, sessionID) {
				c.Next()
				return
			}
		}

		// No valid authentication found, continue without user context
		c.Next()
	}
}

// extractToken extracts JWT token from request
func extractToken(c *gin.Context) string {
	// Try Authorization header first
	if authHeader := c.GetHeader(consts.AuthorizationKey); authHeader != "" {
		if strings.HasPrefix(authHeader, consts.BearerKey) {
			return strings.TrimPrefix(authHeader, consts.BearerKey)
		}
	}

	// Try query parameter for API access
	if queryToken := c.Query("token"); queryToken != "" {
		return queryToken
	}

	return ""
}

// handleTokenAuth handles JWT token authentication
func handleTokenAuth(c *gin.Context, jtm *jwt.TokenManager, token string) bool {
	claims, err := jtm.DecodeToken(token)
	if err != nil {
		logger.Debugf(c.Request.Context(), "Token validation failed: %v", err)
		return false
	}

	// Set user context from token claims
	ctx := setUserContextFromToken(c, claims)
	c.Request = c.Request.WithContext(ctx)

	// Check if token needs refresh
	if shouldRefreshToken(claims) {
		newToken, refreshed, err := jtm.RefreshTokenIfNeeded(token, 10*time.Minute)
		if err != nil {
			logger.Warnf(ctx, "Failed to refresh token: %v", err)
		} else if refreshed {
			c.Header(consts.AuthorizationKey, consts.BearerKey+newToken)
			logger.Infof(ctx, "Token refreshed for user %s", jwt.GetPayloadString(claims, "user_id"))
		}
	}

	return true
}

// handleSessionAuth handles session authentication with complete user context
func handleSessionAuth(c *gin.Context, sm *ServiceManager, sessionID string) bool {
	ctx := c.Request.Context()
	asw := sm.AuthServiceWrapper()

	// Get session from service
	session, err := asw.GetSessionByID(ctx, sessionID)
	if err != nil {
		logger.Debugf(ctx, "Session not found: %v", err)
		return false
	}

	// Check if session is active
	if !session.IsActive {
		logger.Debugf(ctx, "Session inactive: %s", sessionID)
		return false
	}

	// Check session expiration
	if session.ExpiresAt != nil && time.Now().UnixMilli() > *session.ExpiresAt {
		logger.Debugf(ctx, "Session expired: %s", sessionID)
		go func() {
			if err := asw.DeleteSession(context.Background(), sessionID); err != nil {
				logger.Warnf(context.Background(), "Failed to delete expired session: %v", err)
			}
		}()
		return false
	}

	// Set complete user context from session
	ctx = setCompleteUserContextFromSession(c, session, sm)
	c.Request = c.Request.WithContext(ctx)

	// Update session last access time asynchronously
	go func() {
		if err := asw.UpdateSessionLastAccess(context.Background(), session.TokenID); err != nil {
			logger.Warnf(context.Background(), "Failed to update session last access: %v", err)
		}
	}()

	return true
}

// setUserContextFromToken sets user context from JWT token claims
func setUserContextFromToken(c *gin.Context, claims map[string]any) context.Context {
	ctx := c.Request.Context()
	if _, ok := ctxutil.GetGinContext(ctx); !ok {
		ctx = ctxutil.WithGinContext(ctx, c)
	}

	// Extract user info from token
	if userID := jwt.GetPayloadString(claims, "user_id"); userID != "" {
		ctx = ctxutil.SetUserID(ctx, userID)
	}

	if spaceID := jwt.GetPayloadString(claims, "space_id"); spaceID != "" {
		ctx = ctxutil.SetSpaceID(ctx, spaceID)
	}

	// Set user attributes
	roles := jwt.GetPayloadStringSlice(claims, "roles")
	permissions := jwt.GetPayloadStringSlice(claims, "permissions")
	isAdmin := jwt.GetPayloadBool(claims, "is_admin")
	username := jwt.GetPayloadString(claims, "username")
	email := jwt.GetPayloadString(claims, "email")
	userStatus := jwt.GetPayloadInt(claims, "user_status")
	isCertified := jwt.GetPayloadBool(claims, "is_certified")

	ctx = ctxutil.SetUsername(ctx, username)
	ctx = ctxutil.SetUserEmail(ctx, email)
	ctx = ctxutil.SetUserStatus(ctx, userStatus)
	ctx = ctxutil.SetUserIsCertified(ctx, isCertified)
	ctx = ctxutil.SetUserRoles(ctx, roles)
	ctx = ctxutil.SetUserPermissions(ctx, permissions)
	ctx = ctxutil.SetUserIsAdmin(ctx, isAdmin)

	if spaceIDs := jwt.GetPayloadStringSlice(claims, "space_ids"); len(spaceIDs) > 0 {
		ctx = ctxutil.SetUserSpaceIDs(ctx, spaceIDs)
	}

	// Set in Gin context for compatibility
	c.Set("user_id", jwt.GetPayloadString(claims, "user_id"))
	c.Set("username", username)
	c.Set("email", email)
	c.Set("roles", roles)
	c.Set("permissions", permissions)
	c.Set("is_admin", isAdmin)
	c.Set("user_status", userStatus)
	c.Set("is_certified", isCertified)
	c.Set("auth_method", "token")

	return ctx
}

// setCompleteUserContextFromSession sets complete user context from session data
func setCompleteUserContextFromSession(c *gin.Context, session *authStructs.ReadSession, sm *ServiceManager) context.Context {
	ctx := c.Request.Context()
	if _, ok := ctxutil.GetGinContext(ctx); !ok {
		ctx = ctxutil.WithGinContext(ctx, c)
	}

	userID := session.UserID
	ctx = ctxutil.SetUserID(ctx, userID)

	// Get service wrappers
	usw := sm.UserServiceWrapper()
	asw := sm.AccessServiceWrapper()
	tsw := sm.SpaceServiceWrapper()

	// Get user details
	if user, err := usw.GetUserByID(ctx, userID); err == nil && user != nil {
		ctx = ctxutil.SetUsername(ctx, user.Username)
		ctx = ctxutil.SetUserEmail(ctx, user.Email)
		ctx = ctxutil.SetUserStatus(ctx, user.Status)
		ctx = ctxutil.SetUserIsCertified(ctx, user.IsCertified)

		// Set in Gin context
		c.Set("username", user.Username)
		c.Set("email", user.Email)
		c.Set("user_status", user.Status)
		c.Set("is_certified", user.IsCertified)
	}

	// Get user spaces
	var spaceIDs []string
	if spaces, err := tsw.GetUserSpaces(ctx, userID); err == nil && len(spaces) > 0 {
		for _, t := range spaces {
			spaceIDs = append(spaceIDs, t.ID)
		}
		ctx = ctxutil.SetUserSpaceIDs(ctx, spaceIDs)

		// Set default space if not already set
		if ctxutil.GetSpaceID(ctx) == "" {
			if defaultSpace, err := tsw.GetUserDefaultSpace(ctx, userID); err == nil && defaultSpace != nil {
				ctx = ctxutil.SetSpaceID(ctx, defaultSpace.ID)
			}
		}
	}

	// Get user roles and permissions
	var roles []string
	var permissions []string
	var isAdmin bool

	spaceID := ctxutil.GetSpaceID(ctx)

	// Get global roles
	if globalRoles, err := asw.GetUserRoles(ctx, userID); err == nil {
		for _, role := range globalRoles {
			roles = append(roles, role.Slug)
		}
	}

	// Get space-specific roles if space context exists
	if spaceID != "" {
		if roleIDs, err := asw.GetUserRolesInSpace(ctx, userID, spaceID); err == nil && len(roleIDs) > 0 {
			if spaceRoles, err := asw.GetRolesByIDs(ctx, roleIDs); err == nil {
				for _, role := range spaceRoles {
					if !utils.Contains(roles, role.Slug) {
						roles = append(roles, role.Slug)
					}
				}
			}
		}
	}

	// Check admin status
	isAdmin = hasAdminRole(roles)

	// Get permissions for all roles
	if len(roles) > 0 {
		permissionSet := make(map[string]struct{})

		// For global roles, get permissions
		if globalRoles, err := asw.GetUserRoles(ctx, userID); err == nil {
			addRolePermissions(ctx, asw, globalRoles, permissionSet)
		}

		// For space roles, get permissions
		if spaceID != "" {
			if roleIDs, err := asw.GetUserRolesInSpace(ctx, userID, spaceID); err == nil && len(roleIDs) > 0 {
				if spaceRoles, err := asw.GetRolesByIDs(ctx, roleIDs); err == nil {
					addRolePermissions(ctx, asw, spaceRoles, permissionSet)
				}
			}
		}

		// Convert set to slice
		for permCode := range permissionSet {
			permissions = append(permissions, permCode)
		}
	}

	// Set context values
	ctx = ctxutil.SetUserRoles(ctx, roles)
	ctx = ctxutil.SetUserPermissions(ctx, permissions)
	ctx = ctxutil.SetUserIsAdmin(ctx, isAdmin)

	// Set in Gin context for compatibility
	c.Set("user_id", userID)
	c.Set("session_id", session.ID)
	c.Set("auth_method", "session")
	c.Set("roles", roles)
	c.Set("permissions", permissions)
	c.Set("is_admin", isAdmin)
	return ctx
}

func addRolePermissions(ctx context.Context, asw *AccessServiceWrapper, roles []*accessStructs.ReadRole, permissionSet map[string]struct{}) {
	for _, role := range roles {
		rolePermissions, err := asw.GetRolePermissions(ctx, role.ID)
		if err != nil {
			continue
		}
		for _, perm := range rolePermissions {
			for _, code := range buildPermissionCodes(perm) {
				permissionSet[code] = struct{}{}
			}
		}
	}
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

// shouldRefreshToken checks if token should be refreshed
func shouldRefreshToken(claims map[string]any) bool {
	return jwt.IsTokenStale(claims, time.Hour)
}

// RequireAuth middleware requires authentication (either token or session)
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		userID := ctxutil.GetUserID(ctx)

		if userID == "" {
			resp.Fail(c.Writer, resp.UnAuthorized("Authentication required"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireTokenAuth middleware requires JWT token authentication specifically
func RequireTokenAuth(em ext.ManagerInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			resp.Fail(c.Writer, resp.UnAuthorized("JWT token required"))
			c.Abort()
			return
		}
		// Get service wrappers manager
		sm := GetServiceManager(em)
		// get access wrapper
		asw := sm.AuthServiceWrapper()
		if jtm := asw.GetTokenManager(); jtm != nil {
			if !handleTokenAuth(c, jtm, token) {
				resp.Fail(c.Writer, resp.UnAuthorized("Invalid JWT token"))
				c.Abort()
				return
			}
		} else {
			resp.Fail(c.Writer, resp.UnAuthorized("Token manager not available"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireSessionAuth middleware requires session cookie authentication specifically
func RequireSessionAuth(em ext.ManagerInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := cookie.GetSessionID(c.Request)
		if err != nil || sessionID == "" {
			resp.Fail(c.Writer, resp.UnAuthorized("Session required"))
			c.Abort()
			return
		}
		if !handleSessionAuth(c, GetServiceManager(em), sessionID) {
			resp.Fail(c.Writer, resp.UnAuthorized("Invalid session"))
			c.Abort()
			return
		}

		c.Next()
	}
}
