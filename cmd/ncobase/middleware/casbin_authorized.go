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

// CasbinAuthorized is a middleware that checks user authorization using Casbin
// Model: r = sub, dom, obj, act, v4, v5
func CasbinAuthorized(em ext.ManagerInterface, enforcer *casbin.Enforcer, whiteList []string) gin.HandlerFunc {
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

		action := mapHTTPMethodToAction(httpMethod)

		hasPermission := checkPermission(ctx, enforcer, userID, username, tenantID,
			resource, httpMethod, action, roles, permissions, isAdmin)

		eventMetadata := types.JSON{
			"username": username, "tenant_id": tenantID, "resource": resource,
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

		em.PublishEvent("security.access_granted", types.JSON{
			"user_id": userID, "details": "Access granted to resource: " + resource, "metadata": eventMetadata,
		})

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
	userID, username, tenantID, resource, httpMethod, action string,
	roles, permissions []string, isAdmin bool) bool {

	// Strategy 1: Super admin bypass (handled by Casbin matcher)
	if isAdmin && hasAdminRole(roles) {
		logger.Debugf(ctx, "Admin user detected: %s", userID)
		// Still go through Casbin for consistent logging and rule evaluation
	}

	// Strategy 2: Check wildcard permissions (application level)
	if hasWildcardPermission(permissions) {
		logger.Debugf(ctx, "Wildcard permission granted for user %s", userID)
		return true
	}

	// Strategy 3: Check roles using Casbin
	if enforcer != nil {
		domain := tenantID
		if domain == "" {
			domain = "*"
		}

		// Check roles with HTTP method
		for _, role := range roles {
			// 6-parameter format: sub, dom, obj, act, v4, v5
			if allowed, err := enforcer.Enforce(role, domain, resource, httpMethod, "", ""); err == nil && allowed {
				logger.Debugf(ctx, "Casbin permission (HTTP method) granted for role %s", role)
				return true
			}

			// Check with semantic action
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

		// Strategy 4: Direct check permissions
		if allowed, err := enforcer.Enforce(userID, domain, resource, httpMethod, "", ""); err == nil && allowed {
			logger.Debugf(ctx, "Direct user permission (HTTP method) granted for %s", userID)
			return true
		}
		if allowed, err := enforcer.Enforce(userID, domain, resource, action, "", ""); err == nil && allowed {
			logger.Debugf(ctx, "Direct user permission (semantic action) granted for %s", userID)
			return true
		}

		// Strategy 5: Username-based permissions
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
