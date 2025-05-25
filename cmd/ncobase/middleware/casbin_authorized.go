package middleware

import (
	"context"
	"strings"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ActionMapping maps HTTP methods to permission actions
var ActionMapping = map[string]string{
	"GET":     "read",
	"POST":    "create",
	"PUT":     "update",
	"PATCH":   "update",
	"DELETE":  "delete",
	"HEAD":    "read",
	"OPTIONS": "read",
}

// CasbinAuthorized middleware with enhanced permission mapping
func CasbinAuthorized(enforcer *casbin.Enforcer, whiteList []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipPath(c.Request, whiteList) {
			c.Next()
			return
		}

		ctx := ctxutil.FromGinContext(c)

		// Validate user authentication
		userID := ctxutil.GetUserID(ctx)
		if validator.IsEmpty(userID) {
			resp.Fail(c.Writer, resp.UnAuthorized(ecode.Text(ecode.Unauthorized)))
			c.Abort()
			return
		}

		// Get request info
		resource := c.Request.URL.Path
		httpMethod := c.Request.Method
		tenantID := ctxutil.GetTenantID(ctx)

		// Get user authorization info
		username := ctxutil.GetUsername(ctx)
		roles := ctxutil.GetUserRoles(ctx)
		permissions := ctxutil.GetUserPermissions(ctx)
		isAdmin := ctxutil.GetUserIsAdmin(ctx)

		// Convert HTTP method to permission action
		permissionAction := mapHTTPMethodToAction(httpMethod)
		permissionSubject := mapResourceToSubject(resource)

		logger.WithFields(ctx, logrus.Fields{
			"user_id":   userID,
			"tenant_id": tenantID,
			"resource":  resource,
			"method":    httpMethod,
			"action":    permissionAction,
			"subject":   permissionSubject,
			"roles":     roles,
			"is_admin":  isAdmin,
		}).Debug("Checking authorization with mapped permissions")

		// Check permission using both original and mapped approaches
		hasPermission := checkPermission(ctx, enforcer, userID, tenantID,
			resource, httpMethod, permissionAction, permissionSubject, roles, permissions, isAdmin)

		if !hasPermission {
			logger.WithFields(ctx, logrus.Fields{
				"username": username,
				"resource": resource,
				"action":   permissionAction,
				"subject":  permissionSubject,
			}).Warn("Access denied")

			resp.Fail(c.Writer, resp.Forbidden("Access denied"))
			c.Abort()
			return
		}

		logger.WithFields(ctx, logrus.Fields{
			"username": username,
			"resource": resource,
			"action":   permissionAction,
		}).Debug("Access granted")

		c.Next()
	}
}

// mapHTTPMethodToAction converts HTTP method to permission action
func mapHTTPMethodToAction(method string) string {
	if action, exists := ActionMapping[strings.ToUpper(method)]; exists {
		return action
	}
	return strings.ToLower(method) // fallback to original method
}

// mapResourceToSubject extracts permission subject from resource path
func mapResourceToSubject(resource string) string {
	// Extract subject from resource path
	parts := strings.Split(strings.Trim(resource, "/"), "/")
	if len(parts) >= 2 {
		// For paths like /api/v1/users -> users
		return parts[len(parts)-1]
	}

	return "resource" // default fallback
}

// checkPermission performs authorization check
func checkPermission(ctx context.Context, enforcer *casbin.Enforcer,
	userID, tenantID, resource, httpMethod, permissionAction, permissionSubject string,
	roles, permissions []string, isAdmin bool) bool {

	// Strategy 1: Super admin bypass with role verification
	if isAdmin && hasAdminRole(roles) {
		logger.Debugf(ctx, "Admin access granted for user %s", userID)
		return true
	}

	// Strategy 2: Check mapped permission format (action:subject)
	mappedPermission := permissionAction + ":" + permissionSubject
	if hasSpecificPermission(permissions, mappedPermission) {
		logger.Debugf(ctx, "Mapped permission %s granted for user %s", mappedPermission, userID)
		return true
	}

	// Strategy 3: Check wildcard permissions
	if hasWildcardPermission(permissions) {
		logger.Debugf(ctx, "Wildcard permission granted for user %s", userID)
		return true
	}

	// Strategy 4: Check pattern-based permissions
	if hasPatternPermission(permissions, permissionAction, permissionSubject) {
		logger.Debugf(ctx, "Pattern permission granted for user %s", userID)
		return true
	}

	// Strategy 5: Casbin role-based authorization (both original and mapped)
	if enforcer != nil {
		domain := tenantID
		if domain == "" {
			domain = "*"
		}

		// Check with user roles using both original and mapped formats
		for _, role := range roles {
			// Try original resource/method format
			if allowed, err := enforcer.Enforce(role, domain, resource, httpMethod); err == nil && allowed {
				logger.Debugf(ctx, "Casbin permission (original) granted for role %s", role)
				return true
			}

			// Try mapped action/subject format
			if allowed, err := enforcer.Enforce(role, domain, permissionSubject, permissionAction); err == nil && allowed {
				logger.Debugf(ctx, "Casbin permission (mapped) granted for role %s", role)
				return true
			}

			// Try wildcard domain
			if domain != "*" {
				if allowed, err := enforcer.Enforce(role, "*", resource, httpMethod); err == nil && allowed {
					logger.Debugf(ctx, "Casbin permission (wildcard domain) granted for role %s", role)
					return true
				}
			}
		}

		// Strategy 6: Direct user permission check
		if allowed, err := enforcer.Enforce(userID, domain, resource, httpMethod); err == nil && allowed {
			logger.Debugf(ctx, "Direct user permission granted for %s", userID)
			return true
		}
	}

	return false
}
