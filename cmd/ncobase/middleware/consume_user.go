package middleware

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/ncobase/ncore/types"

	"github.com/gin-gonic/gin"
)

// refreshToken refreshes an access token if it's about to expire
func refreshToken(jtm *jwt.TokenManager, oldToken string) (string, error) {
	// Decode the token to get the payload
	tokenData, err := jtm.DecodeToken(oldToken)
	if err != nil {
		return "", fmt.Errorf("failed to decode token: %v", err)
	}

	// Extract payload and user ID
	payload, ok := tokenData["payload"].(map[string]any)
	if !ok {
		return "", errors.New("invalid token payload format")
	}

	userID, ok := payload["user_id"].(string)
	if !ok || userID == "" {
		return "", errors.New("user_id not found in token payload")
	}

	// Extract tenant ID and other information
	tenantID, _ := payload["tenant_id"].(string)
	roles, _ := payload["roles"].([]string)
	permissions, _ := payload["permissions"].([]string)
	isAdmin, _ := payload["is_admin"].(bool)
	tenantIDs, _ := payload["tenant_ids"].([]string)

	// Create updated payload
	newPayload := types.JSON{
		"user_id":     userID,
		"tenant_id":   tenantID,
		"tenant_ids":  tenantIDs,
		"roles":       roles,
		"permissions": permissions,
		"is_admin":    isAdmin,
	}

	// Generate only a new access token (keep the same refresh token)
	tokenID := tokenData["jti"].(string)
	accessToken, _ := jtm.GenerateAccessToken(tokenID, newPayload)

	return accessToken, nil
}

// isTokenExpiring checks if the token is expiring soon.
func isTokenExpiring(tokenData map[string]any) bool {
	exp, ok := tokenData["exp"].(int64)
	if !ok {
		return false
	}
	expirationTime := time.Unix(exp, 0)
	return time.Until(expirationTime) < 10*time.Minute // Assumes token should be refreshed if expiring within 10 minutes
}

// ConsumeUser consumes user authentication information.
func ConsumeUser(jtm *jwt.TokenManager, whiteList []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipPath(c.Request, whiteList) {
			c.Next()
			return
		}

		var token string
		// Retrieve token from request header, query, or cookie
		if authHeader := c.GetHeader(consts.AuthorizationKey); authHeader != "" {
			// Token from header
			if strings.HasPrefix(authHeader, consts.BearerKey) {
				token = strings.TrimPrefix(authHeader, consts.BearerKey)
			}
		} else if queryToken := c.Query("ak"); queryToken != "" {
			// Token from query
			token = queryToken
		} else if cookieToken, err := c.Cookie("access_token"); err == nil {
			// Token from cookie
			token = cookieToken
		}

		// If token is still empty, proceed to the next middleware
		if token == "" {
			c.Next()
			return
		}

		// Decode token
		tokenData, err := jtm.DecodeToken(token)
		if err != nil {
			handleTokenError(c, fmt.Errorf("failed to decode token: %v", err))
			return
		}

		// Extract payload from token data
		payload, ok := tokenData["payload"].(map[string]any)
		if !ok {
			handleTokenError(c, errors.New("invalid token payload format"))
			return
		}

		// Create context with gin
		ctx := ctxutil.WithGinContext(context.Background(), c)

		// Set user ID to context
		userID, ok := payload["user_id"].(string)
		if !ok || userID == "" {
			handleTokenError(c, errors.New("user_id not found in token payload"))
			return
		}
		ctx = ctxutil.SetUserID(ctx, userID)

		// Extract and set tenant ID
		tenantID := jwt.GetTenantIDFromToken(tokenData)
		if tenantID != "" {
			ctx = ctxutil.SetTenantID(ctx, tenantID)
		}

		// Extract and set tenant IDs
		if tenantIDs, ok := payload["tenant_ids"].([]interface{}); ok {
			var tenantIDStrings []string
			for _, id := range tenantIDs {
				if idStr, ok := id.(string); ok {
					tenantIDStrings = append(tenantIDStrings, idStr)
				}
			}
			ctx = ctxutil.SetUserTenantIDs(ctx, tenantIDStrings)
		}

		// Extract roles and permissions
		roles := jwt.GetRolesFromToken(tokenData)
		permissions := jwt.GetPermissionsFromToken(tokenData)
		isAdmin := jwt.IsAdminFromToken(tokenData)

		// Set roles, permissions and admin status to context
		ctx = ctxutil.SetUserRoles(ctx, roles)
		ctx = ctxutil.SetUserPermissions(ctx, permissions)
		ctx = ctxutil.SetUserIsAdmin(ctx, isAdmin)

		// Update request context
		c.Request = c.Request.WithContext(ctx)

		// Set values in Gin context for backward compatibility
		c.Set("roles", roles)
		c.Set("permissions", permissions)
		c.Set("is_admin", isAdmin)

		// Check if token is expiring soon, and refresh if necessary
		if isTokenExpiring(tokenData) {
			newToken, err := refreshToken(jtm, token)
			if err != nil {
				handleTokenError(c, fmt.Errorf("failed to refresh token: %v", err))
				return
			}
			c.Header(consts.AuthorizationKey, consts.BearerKey+newToken)
			cookie.SetAccessToken(c.Writer, newToken, "")
		}

		// Continue to next middleware or handler
		c.Next()
	}
}

// handleTokenError handles token decoding/validation errors.
func handleTokenError(c *gin.Context, err error) {
	var (
		status  int
		code    int
		message string
	)
	ctx := ctxutil.FromGinContext(c)
	// Logging the error
	logger.Errorf(ctx, "Token error: %v", err)

	switch {
	case strings.Contains(err.Error(), "token is expired"):
		status = http.StatusUnauthorized
		code = ecode.Unauthorized
		message = "Token has expired. Please login again."
	case strings.Contains(err.Error(), "token is invalid"):
		status = http.StatusUnauthorized
		code = ecode.Unauthorized
		message = "Token is invalid. Please check your token and try again."
	case strings.Contains(err.Error(), "signature is invalid"):
		status = http.StatusUnauthorized
		code = ecode.Unauthorized
		message = "Signature is invalid. Please check your token and try again."
	case strings.Contains(err.Error(), "token is malformed"):
		status = http.StatusUnauthorized
		code = ecode.Unauthorized
		message = "Token is malformed. Please check your token and try again."
	case strings.Contains(err.Error(), "token is missing"):
		status = http.StatusUnauthorized
		code = ecode.Unauthorized
		message = "Token is missing. Please provide a valid token."
	case strings.Contains(err.Error(), "user not found"):
		status = http.StatusUnauthorized
		code = ecode.Unauthorized
		message = "User not found. Please check your credentials and try again."
	default:
		status = http.StatusForbidden
		code = ecode.AccessDenied
		message = "Access denied. You do not have the necessary permissions to access this resource."
	}

	exception := &resp.Exception{
		Status:  status,
		Code:    code,
		Message: message,
	}
	resp.Fail(c.Writer, exception)
	c.Abort()
}
