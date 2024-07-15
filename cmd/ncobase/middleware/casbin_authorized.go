package middleware

import (
	"ncobase/common/ecode"
	"ncobase/common/log"
	"ncobase/common/resp"
	"ncobase/common/types"
	"ncobase/common/util"
	"ncobase/common/validator"
	"ncobase/feature/access/service"
	"ncobase/feature/access/structs"
	"ncobase/helper"
	"net/http"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

// handleException is a helper function to handle exceptions and abort the request with the appropriate response
func handleException(c *gin.Context, exception *resp.Exception, err error, message string) {
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code":    ecode.ServerErr,
			"message": message,
		})
		return
	}
	if exception.Code != 0 {
		c.AbortWithStatusJSON(exception.Status, gin.H{
			"code":    exception.Code,
			"message": exception.Message,
		})
		return
	}
}

// CasbinAuthorized is a middleware that checks if the user has permission to access the resource
func CasbinAuthorized(enforcer *casbin.Enforcer, whiteList []string, svc *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if inWhiteList(c.Request.URL.Path, whiteList) {
			c.Next()
			return
		}

		ctx := c.Request.Context()

		currentUser := helper.GetUserID(ctx)
		currentTenant := helper.GetTenantID(ctx)

		// Check if user ID or tenant ID is empty
		if validator.IsEmpty(currentUser) || validator.IsEmpty(currentTenant) {
			// Respond with unauthorized error
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    ecode.Unauthorized,
				"message": ecode.Text(ecode.Unauthorized),
			})
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

		roles = util.RemoveDuplicates(roleIDs)
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

		log.Infof(c, "Checking permission for userID: %s, tenantID: %s, obj: %s, act: %s\n", currentUser, currentTenant, obj, act)
		log.Infof(c, "User roles: %v\n", roles)

		// Check if the user has permission to access the resource
		ok, err := checkPermission(enforcer, currentUser, obj, act, roles, permissions, currentTenant)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":    ecode.ServerErr,
				"message": "Error evaluating permission, please contact the administrator",
			})
			return
		}

		if !ok {
			log.Warnf(c, "Permission denied for userID: %s, tenantID: %s, obj: %s, act: %s\n", currentUser, currentTenant, obj, act)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    ecode.AccessDenied,
				"message": "You don't have permission to access this resource, please contact the administrator",
			})
			return
		}

		log.Infof(c, "Permission granted for userID: %s, tenantID: %s, obj: %s, act: %s\n", currentUser, currentTenant, obj, act)
		c.Next()
	}
}

// checkPermission checks if the user has permission to access the resource based on roles and permissions
func checkPermission(enforcer *casbin.Enforcer, userID string, obj string, act string, roles []string, permissions []*structs.ReadPermission, tenantID string) (bool, error) {
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
	ok, err := enforcer.Enforce(userID, tenantID, "*", "*", nil, nil)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}

	return false, nil
}
