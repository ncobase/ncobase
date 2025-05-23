package middleware

import (
	"context"
	accessService "ncobase/access/service"
	tenantService "ncobase/tenant/service"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// CasbinAuthorized checks permissions using current Casbin model
func CasbinAuthorized(enforcer *casbin.Enforcer, whiteList []string, as *accessService.Service, ts *tenantService.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipPath(c.Request, whiteList) {
			c.Next()
			return
		}

		ctx := ctxutil.FromGinContext(c)
		currentUser := ctxutil.GetUserID(ctx)
		currentTenant := ctxutil.GetTenantID(ctx)

		// Check if user is authenticated
		if validator.IsEmpty(currentUser) {
			resp.Fail(c.Writer, resp.UnAuthorized(ecode.Text(ecode.Unauthorized)))
			c.Abort()
			return
		}

		// Resolve tenant if not present
		if validator.IsEmpty(currentTenant) {
			if tenant, err := ts.UserTenant.UserBelongTenant(ctx, currentUser); err == nil && tenant != nil {
				currentTenant = tenant.ID
				ctx = ctxutil.SetTenantID(ctx, currentTenant)
				c.Request = c.Request.WithContext(ctx)
			}
		}

		// Get request details
		resource := c.Request.URL.Path
		action := c.Request.Method

		// Get user information from context
		roles := ctxutil.GetUserRoles(ctx)
		permissions := ctxutil.GetUserPermissions(ctx)
		isAdmin := ctxutil.GetUserIsAdmin(ctx)

		logger.WithFields(ctx, logrus.Fields{
			"user_id":     currentUser,
			"tenant_id":   currentTenant,
			"resource":    resource,
			"action":      action,
			"roles":       roles,
			"permissions": permissions,
			"is_admin":    isAdmin,
		}).Info("Checking permission with Casbin")

		// Check user permission
		hasPermission, err := checkUserPermission(ctx, enforcer, currentUser, currentTenant, resource, action, roles, permissions, isAdmin)
		if err != nil {
			logger.Errorf(ctx, "Error checking permission: %v", err)
			resp.Fail(c.Writer, resp.InternalServer("Permission check failed"))
			c.Abort()
			return
		}

		if !hasPermission {
			logger.WithFields(ctx, logrus.Fields{
				"user_id":  currentUser,
				"resource": resource,
				"action":   action,
			}).Warn("Permission denied")

			resp.Fail(c.Writer, resp.Forbidden("Access denied"))
			c.Abort()
			return
		}

		logger.WithFields(ctx, logrus.Fields{
			"user_id":  currentUser,
			"resource": resource,
			"action":   action,
		}).Info("Permission granted")

		c.Next()
	}
}

// checkUserPermission checks permission using the 6-parameter Casbin model
func checkUserPermission(ctx context.Context, enforcer *casbin.Enforcer, userID, tenantID, resource, action string, roles, permissions []string, isAdmin bool) (bool, error) {

	// Strategy 1: Super admin bypass
	if isAdmin || userID == "super-admin" {
		logger.Debugf(ctx, "Admin access granted for user %s", userID)
		return true, nil
	}

	// Strategy 2: Check wildcard permissions from token
	for _, perm := range permissions {
		if perm == "*:*" {
			logger.Debugf(ctx, "Wildcard permission granted from token")
			return true, nil
		}
	}

	// Strategy 3: Check direct permission match from token
	requiredPerm := mapHTTPToPermission(resource, action)
	if requiredPerm != "" {
		for _, perm := range permissions {
			if perm == requiredPerm {
				logger.Debugf(ctx, "Direct permission match from token: %s", perm)
				return true, nil
			}
		}
	}

	// Strategy 4: Use Casbin enforcer with 6-parameter model
	if enforcer != nil {
		// Use default tenant if not specified
		domain := tenantID
		if domain == "" {
			domain = "*" // or use a default domain
		}

		// Check with user roles using 6-parameter model: sub, dom, obj, act, v4, v5
		for _, role := range roles {
			if ok, err := enforcer.Enforce(role, domain, resource, action, nil, nil); err == nil && ok {
				logger.Debugf(ctx, "Casbin permission granted for role %s in domain %s", role, domain)
				return true, nil
			}

			// Also try with wildcard domain
			if domain != "*" {
				if ok, err := enforcer.Enforce(role, "*", resource, action, nil, nil); err == nil && ok {
					logger.Debugf(ctx, "Casbin permission granted for role %s in wildcard domain", role)
					return true, nil
				}
			}
		}

		// Check direct user permission
		if ok, err := enforcer.Enforce(userID, domain, resource, action, nil, nil); err == nil && ok {
			logger.Debugf(ctx, "Casbin permission granted for user %s directly", userID)
			return true, nil
		}

		// Also try with wildcard domain for user
		if domain != "*" {
			if ok, err := enforcer.Enforce(userID, "*", resource, action, nil, nil); err == nil && ok {
				logger.Debugf(ctx, "Casbin permission granted for user %s in wildcard domain", userID)
				return true, nil
			}
		}
	}

	return false, nil
}

// mapHTTPToPermission maps HTTP requests to permission strings (same as before)
func mapHTTPToPermission(resource, action string) string {
	actionMap := map[string]string{
		"GET":    "read",
		"POST":   "create",
		"PUT":    "update",
		"PATCH":  "update",
		"DELETE": "delete",
	}

	permAction := actionMap[action]
	if permAction == "" {
		return ""
	}

	subject := extractSubjectFromPath(resource)
	if subject == "" {
		return ""
	}

	return permAction + ":" + subject
}

// extractSubjectFromPath extracts subject from resource path (same as before)
func extractSubjectFromPath(path string) string {
	pathToSubject := map[string]string{
		"/iam/account":            "account",
		"/iam/account/tenant":     "account",
		"/iam/account/tenants":    "account",
		"/user/users":             "user",
		"/user/employees":         "employee",
		"/sys/menus":              "menu",
		"/sys/dictionaries":       "dictionary",
		"/sys/options":            "system",
		"/access/roles":           "role",
		"/access/permissions":     "permission",
		"/tenant/tenants":         "tenant",
		"/space/groups":           "group",
		"/content/topics":         "content",
		"/content/taxonomies":     "taxonomy",
		"/resources":              "resource",
		"/workflow/processes":     "workflow",
		"/workflow/tasks":         "task",
		"/payment/orders":         "payment",
		"/realtime/notifications": "realtime",
	}

	// Check exact match first
	if subject, exists := pathToSubject[path]; exists {
		return subject
	}

	// Check prefix match for paths with parameters
	for pathPrefix, subject := range pathToSubject {
		if len(path) > len(pathPrefix) && path[:len(pathPrefix)] == pathPrefix {
			if path[len(pathPrefix)] == '/' {
				return subject
			}
		}
	}

	return ""
}
