package middleware

import (
	"context"
	"fmt"
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
		permissionSet := make(map[string]bool)

		// For global roles, get permissions
		if globalRoles, err := asw.GetUserRoles(ctx, userID); err == nil {
			for _, role := range globalRoles {
				if rolePermissions, err := asw.GetRolePermissions(ctx, role.ID); err == nil {
					for _, perm := range rolePermissions {
						permCode := fmt.Sprintf("%s:%s", perm.Action, perm.Subject)
						permissionSet[permCode] = true
					}
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
