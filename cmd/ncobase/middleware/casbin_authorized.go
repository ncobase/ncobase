package middleware

import (
	"context"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/security/jwt"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// CasbinAuthorized middleware for role-based access control using Casbin
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

		// Additional user validation
		if err := validateUserInContext(ctx, c); err != nil {
			logger.Errorf(ctx, "User validation failed: %v", err)
			resp.Fail(c.Writer, resp.UnAuthorized("User validation failed"))
			c.Abort()
			return
		}

		// Get request info
		resource := c.Request.URL.Path
		action := c.Request.Method
		tenantID := ctxutil.GetTenantID(ctx)

		// Get user authorization info
		roles := ctxutil.GetUserRoles(ctx)
		permissions := ctxutil.GetUserPermissions(ctx)
		isAdmin := ctxutil.GetUserIsAdmin(ctx)

		logger.WithFields(ctx, logrus.Fields{
			"user_id":   userID,
			"tenant_id": tenantID,
			"resource":  resource,
			"action":    action,
			"roles":     roles,
			"is_admin":  isAdmin,
		}).Debug("Checking authorization")

		// Check permission
		hasPermission := checkPermission(ctx, enforcer, userID, tenantID, resource, action, roles, permissions, isAdmin)
		if !hasPermission {
			logger.WithFields(ctx, logrus.Fields{
				"user_id":  userID,
				"resource": resource,
				"action":   action,
			}).Warn("Access denied")

			resp.Fail(c.Writer, resp.Forbidden("Access denied"))
			c.Abort()
			return
		}

		logger.WithFields(ctx, logrus.Fields{
			"user_id":  userID,
			"resource": resource,
			"action":   action,
		}).Debug("Access granted")

		c.Next()
	}
}

// validateUserInContext validates user information in context
func validateUserInContext(ctx context.Context, c *gin.Context) error {
	// Get username from context
	username, exists := c.Get("username")
	if !exists || username == "" {
		return nil // Username not required for basic validation
	}

	// Get user status
	userStatus, exists := c.Get("user_status")
	if exists {
		if status, ok := userStatus.(int64); ok && status != 0 {
			return jwt.TokenError("user account is disabled")
		}
	}

	// Additional admin role validation
	isAdmin := ctxutil.GetUserIsAdmin(ctx)
	if isAdmin {
		roles := ctxutil.GetUserRoles(ctx)
		if !hasAdminRole(roles) {
			logger.Warnf(ctx, "User claims admin status but has no admin role")
		}
	}

	return nil
}

// checkPermission performs authorization check
func checkPermission(ctx context.Context, enforcer *casbin.Enforcer, userID, tenantID, resource, action string, roles, permissions []string, isAdmin bool) bool {

	// Strategy 1: Super admin bypass with role verification
	if isAdmin && hasAdminRole(roles) {
		logger.Debugf(ctx, "Admin access granted for user %s", userID)
		return true
	}

	// Strategy 2: Check wildcard permissions
	if hasWildcardPermission(permissions) {
		logger.Debugf(ctx, "Wildcard permission granted for user %s", userID)
		return true
	}

	// Strategy 3: Casbin role-based authorization
	if enforcer != nil {
		domain := tenantID
		if domain == "" {
			domain = "*"
		}

		// Check with user roles
		for _, role := range roles {
			// Try specific tenant domain
			if allowed, err := enforcer.Enforce(role, domain, resource, action); err == nil && allowed {
				logger.Debugf(ctx, "Permission granted for role %s in domain %s", role, domain)
				return true
			}

			// Try wildcard domain
			if domain != "*" {
				if allowed, err := enforcer.Enforce(role, "*", resource, action); err == nil && allowed {
					logger.Debugf(ctx, "Permission granted for role %s in wildcard domain", role)
					return true
				}
			}
		}

		// Strategy 4: Direct user permission check (rare case)
		if allowed, err := enforcer.Enforce(userID, domain, resource, action); err == nil && allowed {
			logger.Debugf(ctx, "Direct user permission granted for %s", userID)
			return true
		}
	}

	return false
}
