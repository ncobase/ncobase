package middleware

import (
	"context"
	"strings"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/ecode"
	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

// ActionMapping maps HTTP methods to semantic actions
var ActionMapping = map[string]string{
	"GET":     "read",
	"POST":    "create",
	"PUT":     "update",
	"PATCH":   "update",
	"DELETE":  "delete",
	"HEAD":    "read",
	"OPTIONS": "read",
}

// CasbinAuthorized middleware checks user authorization using Casbin
func CasbinAuthorized(em ext.ManagerInterface, whiteList []string) gin.HandlerFunc {

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
		// Get Service wrapper manager
		sm := GetServiceManager(em)
		// Get access wrapper
		asw := sm.AccessServiceWrapper()
		// Get Casbin enforcer
		enforcer := asw.GetEnforcer()

		// Get request info
		resource := c.Request.URL.Path
		httpMethod := c.Request.Method
		spaceID := ctxutil.GetSpaceID(ctx)

		// Get user authorization info
		username := ctxutil.GetUsername(ctx)
		roles := ctxutil.GetUserRoles(ctx)
		permissions := ctxutil.GetUserPermissions(ctx)
		isAdmin := ctxutil.GetUserIsAdmin(ctx)

		action := mapHTTPMethodToAction(httpMethod)

		var hasPermission bool
		if enforcer != nil {
			hasPermission = checkPermission(ctx, enforcer, userID, username, spaceID,
				resource, httpMethod, action, roles, permissions, isAdmin)
		} else {
			// Fallback to basic permission check if Casbin not available
			hasPermission = checkBasicPermission(ctx, roles, permissions, isAdmin)
		}

		eventMetadata := types.JSON{
			"username": username, "space_id": spaceID, "resource": resource,
			"method": httpMethod, "action": action, "roles": roles, "is_admin": isAdmin,
			"request_id": ctxutil.GetTraceID(ctx), "client_ip": c.ClientIP(),
			"user_agent": ctxutil.GetUserAgent(ctx),
		}

		if !hasPermission {
			em.PublishEvent("security.access_denied", types.JSON{
				"user_id": userID, "details": "Access denied to resource: " + resource, "metadata": eventMetadata,
			})
			resp.Fail(c.Writer, resp.Forbidden("Access denied"))
			c.Abort()
			return
		}

		// em.PublishEvent("security.access_granted", types.JSON{
		// 	"user_id": userID, "details": "Access granted to resource: " + resource, "metadata": eventMetadata,
		// })

		c.Next()
	}
}

// mapHTTPMethodToAction maps HTTP method to semantic action
func mapHTTPMethodToAction(method string) string {
	if action, exists := ActionMapping[strings.ToUpper(method)]; exists {
		return action
	}
	return strings.ToLower(method)
}

// checkPermission checks user authorization using multiple strategies
func checkPermission(ctx context.Context, enforcer *casbin.Enforcer,
	userID, username, spaceID, resource, httpMethod, action string,
	roles, permissions []string, isAdmin bool) bool {

	// Check wildcard permissions first
	if hasWildcardPermission(permissions) {
		return true
	}

	// Check roles using Casbin
	if enforcer != nil {
		domain := spaceID
		if domain == "" {
			domain = "*"
		}

		// Check roles with HTTP method and semantic action
		for _, role := range roles {
			if allowed, err := enforcer.Enforce(role, domain, resource, httpMethod, "", ""); err == nil && allowed {
				logger.Debugf(ctx, "Casbin permission (HTTP method) granted for role %s", role)
				return true
			}

			if allowed, err := enforcer.Enforce(role, domain, resource, action, "", ""); err == nil && allowed {
				logger.Debugf(ctx, "Casbin permission (semantic action) granted for role %s", role)
				return true
			}

			// Try wildcard domain if specific domain fails
			if domain != "*" {
				if allowed, err := enforcer.Enforce(role, "*", resource, httpMethod, "", ""); err == nil && allowed {
					logger.Debugf(ctx, "Casbin permission (wildcard domain, HTTP) granted for role %s", role)
					return true
				}
				if allowed, err := enforcer.Enforce(role, "*", resource, action, "", ""); err == nil && allowed {
					logger.Debugf(ctx, "Casbin permission (wildcard domain, semantic) granted for role %s", role)
					return true
				}
			}
		}

		// Direct user permissions
		if allowed, err := enforcer.Enforce(userID, domain, resource, httpMethod, "", ""); err == nil && allowed {
			logger.Debugf(ctx, "Direct user permission (HTTP method) granted for %s", userID)
			return true
		}
		if allowed, err := enforcer.Enforce(userID, domain, resource, action, "", ""); err == nil && allowed {
			logger.Debugf(ctx, "Direct user permission (semantic action) granted for %s", userID)
			return true
		}

		// Username-based permissions
		if username != "" {
			if allowed, err := enforcer.Enforce(username, domain, resource, httpMethod, "", ""); err == nil && allowed {
				logger.Debugf(ctx, "Username-based permission (HTTP method) granted for %s", username)
				return true
			}
			if allowed, err := enforcer.Enforce(username, domain, resource, action, "", ""); err == nil && allowed {
				logger.Debugf(ctx, "Username-based permission (semantic action) granted for %s", username)
				return true
			}
		}
	}

	return false
}

// checkBasicPermission provides fallback permission check when Casbin is unavailable
func checkBasicPermission(ctx context.Context, roles, permissions []string, isAdmin bool) bool {
	logger.Warnf(ctx, "Using basic permission check (Casbin unavailable) - falling back to restrictive mode")

	// Admin users have all permissions
	if isAdmin {
		return true
	}

	// Check for wildcard permissions
	if hasWildcardPermission(permissions) {
		return true
	}

	// Only admin roles are permitted when Casbin is unavailable
	if hasAdminRole(roles) {
		return true
	}

	// Deny access by default to prevent unauthorized access
	logger.Warnf(ctx, "Access denied due to Casbin unavailability (user has roles: %v, permissions: %v)", roles, permissions)
	return false
}
