package middleware

import (
	accessService "ncobase/access/service"
	"ncobase/access/structs"
	tenantService "ncobase/tenant/service"
	"strings"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/utils"
	"github.com/ncobase/ncore/utils/convert"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// handleException is a helper function to handle exceptions and abort the request with the appropriate response
func handleException(c *gin.Context, exception *resp.Exception, err error, message string) {
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(message))
		c.Abort()
	}
	if exception.Code != 0 {
		resp.Fail(c.Writer, exception)
		c.Abort()
	}
}

// CasbinAuthorized is a middleware that checks if the user has permission to access the resource
func CasbinAuthorized(enforcer *casbin.Enforcer, whiteList []string, as *accessService.Service, ts *tenantService.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipPath(c.Request, whiteList) {
			c.Next()
			return
		}

		ctx := ctxutil.FromGinContext(c)

		currentUser := ctxutil.GetUserID(ctx)
		currentTenant := ctxutil.GetTenantID(ctx)

		// Check if user ID is empty
		if validator.IsEmpty(currentUser) {
			// Respond with unauthorized error
			resp.Fail(c.Writer, resp.UnAuthorized(ecode.Text(ecode.Unauthorized)))
			c.Abort()
			return
		}

		obj := c.Request.URL.Path
		act := c.Request.Method

		// If tenant ID is empty but user is authenticated, try to get a default tenant
		if validator.IsEmpty(currentTenant) {
			tenant, err := ts.UserTenant.UserBelongTenant(ctx, currentUser)
			if err == nil && tenant != nil {
				currentTenant = tenant.ID
				// Update context with the resolved tenant
				ctx = ctxutil.SetTenantID(ctx, currentTenant)
				c.Request = c.Request.WithContext(ctx)
			}
			// Continue even if no tenant is found - we'll check permissions without tenant context
		}

		// Retrieve user roles - both global and tenant-specific
		var roles []string

		// Get global roles
		userRoles, err := as.UserRole.GetUserRoles(ctx, currentUser)
		if err != nil {
			handleException(c, nil, err, "Error retrieving user roles")
			return
		}

		if len(userRoles) > 0 {
			for _, r := range userRoles {
				roles = append(roles, r.Slug)
			}
		}

		// Get tenant-specific roles if a tenant is available
		if currentTenant != "" {
			roleIDs, err := as.UserTenantRole.GetUserRolesInTenant(ctx, currentUser, currentTenant)
			if err != nil {
				handleException(c, nil, err, "Error retrieving user roles in tenant")
				return
			}

			if len(roleIDs) > 0 {
				roles = append(roles, roleIDs...)
			}
		}

		// If user has no roles at all, deny access
		if len(roles) == 0 {
			handleException(c, resp.Forbidden("User has no roles, please contact your administrator"), nil, "")
			return
		}

		roles = utils.RemoveDuplicates(roles)
		var permissions []*structs.ReadPermission

		// Retrieve role permissions from service
		for _, role := range roles {
			rolePermissions, err := as.RolePermission.GetRolePermissions(ctx, role)
			if err != nil {
				handleException(c, nil, err, "Error retrieving role permissions")
				return
			}
			if c.IsAborted() {
				return
			}

			permissions = append(permissions, rolePermissions...)
		}

		logger.WithFields(ctx,
			logrus.Fields{
				"user_id":     currentUser,
				"tenant_id":   currentTenant,
				"object":      obj,
				"action":      act,
				"roles":       roles,
				"permissions": permissions,
			},
		).Info("Checking permission")

		// Check if the user has permission to access the resource
		ok, err := checkPermission(enforcer, currentUser, obj, act, roles, permissions, currentTenant)
		if err != nil {
			resp.Fail(c.Writer, resp.InternalServer("Error evaluating permission, please contact the administrator"))
			c.Abort()
			return
		}

		if !ok {
			logger.WithFields(ctx, logrus.Fields{
				"user_id":   currentUser,
				"tenant_id": currentTenant,
				"object":    obj,
				"action":    act,
			}).Warn("Permission denied")

			resp.Fail(c.Writer, resp.Forbidden("You don't have permission to access this resource, please contact the administrator"))

			c.Abort()
			return
		}
		logger.WithFields(ctx, logrus.Fields{
			"user_id":   currentUser,
			"tenant_id": currentTenant,
			"object":    obj,
			"action":    act,
		}).Info("Permission granted")

		c.Next()
	}
}

// checkPermission checks if the user has permission to access the resource based on roles and permissions
func checkPermission(enforcer *casbin.Enforcer, userID string, obj string, act string, roles []string, permissions []*structs.ReadPermission, tenantID string) (bool, error) {
	// Special case: if no tenant is specified, try to enforce with empty tenant
	if tenantID == "" {
		// Try to enforce with no tenant context (global permissions)
		for _, role := range roles {
			ok, err := enforcer.Enforce(role, "", obj, act, nil, nil)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
	} else {
		// First, check tenant-specific role permissions
		for _, role := range roles {
			// Try exact tenant match
			ok, err := enforcer.Enforce(role, tenantID, obj, act, tenantID, nil)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}

			// Try wildcard tenant (policies that apply to all tenants)
			ok, err = enforcer.Enforce(role, "*", obj, act, tenantID, nil)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
	}

	// Then, check user-specific permissions
	for _, permission := range permissions {
		if convert.ToValue(permission.Disabled) {
			continue // Skip disabled permissions
		}

		// Check if this permission applies to current tenant
		isTenantSpecific := false
		if extras := convert.ToValue(permission.Extras); extras != nil {
			if tenantIDs, ok := extras["tenant_ids"].([]string); ok {
				isTenantSpecific = true
				if tenantID == "" || !utils.Contains(tenantIDs, tenantID) {
					continue // Skip if permission doesn't apply to this tenant
				}
			}
		}

		// If permission is not tenant-specific or it applies to current tenant
		if !isTenantSpecific || len(tenantID) == 0 {
			// Check for exact match
			if permission.Subject == obj && (permission.Action == act || permission.Action == "*") {
				return true, nil
			}

			// Check for wildcard permissions
			if permission.Subject == "*" && (permission.Action == act || permission.Action == "*") {
				return true, nil
			}

			// Check for partial wildcard matches (e.g., /v1/tenants/* should match /v1/tenants/{slug}/users)
			if strings.HasSuffix(permission.Subject, "*") {
				prefix := strings.TrimSuffix(permission.Subject, "*")
				if strings.HasPrefix(obj, prefix) && (permission.Action == act || permission.Action == "*") {
					return true, nil
				}
			}
		}
	}

	// If no matching permission is found, check if the user has a wildcard permission
	ok, err := enforcer.Enforce(userID, tenantID, "*", "*", tenantID, nil)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}

	// Check global wildcard as last resort
	ok, err = enforcer.Enforce(userID, "*", "*", "*", "*", nil)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}

	return false, nil
}
