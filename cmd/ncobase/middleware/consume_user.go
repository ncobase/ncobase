package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/ncobase/ncore/consts"
	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/cookie"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/security/jwt"

	"github.com/gin-gonic/gin"
)

// ConsumeUser middleware extracts and validates user information from JWT token
func ConsumeUser(jtm *jwt.TokenManager, whiteList []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipPath(c.Request, whiteList) {
			c.Next()
			return
		}

		token := extractToken(c)
		if token == "" {
			c.Next()
			return
		}

		// Decode and validate token
		claims, err := jtm.DecodeToken(token)
		if err != nil {
			handleTokenError(c, err)
			return
		}

		// Set user information in context
		ctx := setUserContext(c, claims)
		c.Request = c.Request.WithContext(ctx)

		// Check if token needs refresh
		if shouldRefreshToken(claims) {
			newToken, refreshed, err := jtm.RefreshTokenIfNeeded(token, 10*time.Minute)
			if err != nil {
				logger.Warnf(ctx, "Failed to refresh token: %v", err)
			} else if refreshed {
				c.Header(consts.AuthorizationKey, consts.BearerKey+newToken)
				cookie.SetAccessToken(c.Writer, newToken, "")
				logger.Infof(ctx, "Token refreshed for user %s", jwt.GetUserIDFromToken(claims))
			}
		}

		c.Next()
	}
}

// extractToken extracts token from various sources
func extractToken(c *gin.Context) string {
	// Try Authorization header first
	if authHeader := c.GetHeader(consts.AuthorizationKey); authHeader != "" {
		if strings.HasPrefix(authHeader, consts.BearerKey) {
			return strings.TrimPrefix(authHeader, consts.BearerKey)
		}
	}

	// Try query parameter
	if queryToken := c.Query("ak"); queryToken != "" {
		return queryToken
	}

	// Try cookie
	if cookieToken, err := c.Cookie("access_token"); err == nil {
		return cookieToken
	}

	return ""
}

// setUserContext sets user information in context from token claims
func setUserContext(c *gin.Context, claims map[string]any) context.Context {
	// Get context
	ctx := c.Request.Context()
	// If no Gin context exists, create one
	if _, ok := ctxutil.GetGinContext(ctx); !ok {
		ctx = ctxutil.WithGinContext(ctx, c)
	}

	// Set basic user info
	if userID := jwt.GetUserIDFromToken(claims); userID != "" {
		ctx = ctxutil.SetUserID(ctx, userID)
	}

	if tenantID := jwt.GetTenantIDFromToken(claims); tenantID != "" {
		ctx = ctxutil.SetTenantID(ctx, tenantID)
	}

	// Set user roles and permissions
	roles := jwt.GetRolesFromToken(claims)
	permissions := jwt.GetPermissionsFromToken(claims)
	isAdmin := jwt.IsAdminFromToken(claims)
	userID := jwt.GetUserIDFromToken(claims)
	username := jwt.GetUsernameFromToken(claims)
	email := jwt.GetEmailFromToken(claims)
	userStatus := jwt.GetUserStatusFromToken(claims)
	isCertified := jwt.IsCertifiedFromToken(claims)

	ctx = ctxutil.SetUsername(ctx, username)
	ctx = ctxutil.SetUserEmail(ctx, email)
	ctx = ctxutil.SetUserStatus(ctx, userStatus)
	ctx = ctxutil.SetUserIsCertified(ctx, isCertified)
	ctx = ctxutil.SetUserRoles(ctx, roles)
	ctx = ctxutil.SetUserPermissions(ctx, permissions)
	ctx = ctxutil.SetUserIsAdmin(ctx, isAdmin)

	// Set tenant IDs
	if tenantIDs := jwt.GetTenantIDsFromToken(claims); len(tenantIDs) > 0 {
		ctx = ctxutil.SetUserTenantIDs(ctx, tenantIDs)
	}

	// Set additional user info in Gin context for compatibility
	c.Set("user_id", userID)
	c.Set("username", username)
	c.Set("email", email)
	c.Set("roles", roles)
	c.Set("permissions", permissions)
	c.Set("is_admin", isAdmin)
	c.Set("user_status", userStatus)
	c.Set("is_certified", isCertified)

	return ctx
}

// shouldRefreshToken checks if token should be refreshed
func shouldRefreshToken(claims map[string]any) bool {
	// Check if token is stale (older than 1 hour)
	return jwt.IsTokenStale(claims, time.Hour)
}

// handleTokenError handles token validation errors
func handleTokenError(c *gin.Context, err error) {
	ctx := ctxutil.FromGinContext(c)
	logger.Errorf(ctx, "Token validation failed: %v", err)

	var status int
	var code int
	var message string

	switch {
	case strings.Contains(err.Error(), "expired"):
		status = http.StatusUnauthorized
		code = ecode.Unauthorized
		message = "Token has expired"
	case strings.Contains(err.Error(), "invalid"):
		status = http.StatusUnauthorized
		code = ecode.Unauthorized
		message = "Invalid token"
	case strings.Contains(err.Error(), "malformed"):
		status = http.StatusUnauthorized
		code = ecode.Unauthorized
		message = "Malformed token"
	default:
		status = http.StatusForbidden
		code = ecode.AccessDenied
		message = "Token validation failed"
	}

	exception := &resp.Exception{
		Status:  status,
		Code:    code,
		Message: message,
	}

	resp.Fail(c.Writer, exception)
	c.Abort()
}
