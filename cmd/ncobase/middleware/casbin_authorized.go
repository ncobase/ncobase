package middleware

import (
	"github.com/ncobase/ncore/pkg/ecode"
	"github.com/ncobase/ncore/pkg/helper"
	"github.com/ncobase/ncore/pkg/logger"
	"github.com/ncobase/ncore/pkg/resp"
	"github.com/ncobase/ncore/pkg/types"
	"github.com/ncobase/ncore/pkg/validator"
	"ncobase/core/access/service"
	"ncobase/core/access/structs"
	"strings"

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
func CasbinAuthorized(enforcer *casbin.Enforcer, whiteList []string, svc *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipPath(c.Request, whiteList) {
			c.Next()
			return
		}

		ctx := helper.FromGinContext(c)

		currentUser := helper.GetUserID(ctx)
		currentTenant := helper.GetTenantID(ctx)

		// Check if user ID or tenant ID is empty
		if validator.IsEmpty(currentUser) || validator.IsEmpty(currentTenant) {
			// Respond with unauthorized error
			resp.Fail(c.Writer, resp.UnAuthorized(ecode.Text(ecode.Unauthorized)))
			c.Abort()
			return
		}

		obj := c.Request.URL.Path
		act := c.Request.Method

		// Retrieve user roles from service
		userRoles, err := svc.UserRole.GetUserRoles(ctx, currentUser)
		if err != nil {
			handleException(c, nil, err, "Error retrieving user roles")
			return
		}
		if c.IsAborted() {
			return
		}

		if len(userRoles) == 0 {
			handleException(c, resp.Forbidden("User has no roles, please contact your administrator"), nil, "")
			return
		}

		var roles []string
		for _, r := range userRoles {
			roles = append(roles, r.Slug)
		}

		// Query current user roles of the current tenant
		roleIDs, err := svc.UserTenantRole.GetUserRolesInTenant(ctx, currentUser, currentTenant)
		if err != nil {
			handleException(c, nil, err, "Error retrieving user roles in tenant")
			return
		}

		if len(roleIDs) == 0 {
			handleException(c, resp.Forbidden("User has no roles in the current tenant, please contact your administrator"), nil, "")
			return
		}

		roles = types.RemoveDuplicates(roleIDs)
		var permissions []*structs.ReadPermission

		// Retrieve role permissions from service
		for _, role := range roles {
			rolePermissions, err := svc.RolePermission.GetRolePermissions(ctx, role)
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
func checkPermission(enforcer *casbin.Enforcer, user_id string, obj string, act string, roles []string, permissions []*structs.ReadPermission, tenantID string) (bool, error) {
	// First, check role-based permissions
	for _, role := range roles {
		ok, err := enforcer.Enforce(role, tenantID, obj, act, nil, nil)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}

	// Then, check user-specific permissions
	for _, permission := range permissions {
		if types.ToValue(permission.Disabled) {
			continue // Skip disabled permissions
		}

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

	// If no matching permission is found, check if the user has a wildcard permission
	ok, err := enforcer.Enforce(user_id, tenantID, "*", "*", nil, nil)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}

	return false, nil
}
