package middleware

import (
	"ncobase/common/log"
	"ncobase/common/validator"
	"ncobase/feature/access/service"
	"ncobase/feature/access/structs"
	"ncobase/helper"
	"net/http"
	"strings"

	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/util"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

// inWhiteList checks if the given path is in the whiteList
func inWhiteList(path string, whiteList []string) bool {
	for _, whitePath := range whiteList {
		if strings.HasPrefix(path, whitePath) {
			return true
		}
	}
	return false
}

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

		ctx := helper.FromGinContext(c)

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

		log.Infof(c, "userID: %s, tenantID: %s, obj: %s, act: %s\n", currentUser, currentTenant, obj, act)

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

		// Check if the user has permission to access the resource
		ok, err := checkPermission(enforcer, currentUser, obj, act, roles, permissions, currentTenant)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":    ecode.ServerErr,
				"message": "Error enforcing policy",
			})
			return
		}

		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    ecode.AccessDenied,
				"message": "You don't have permission to access this resource, please contact the administrator",
			})
			return
		}

		c.Next()
	}
}

// checkPermission checks if the user has permission to access the resource based on roles and permissions
func checkPermission(enforcer *casbin.Enforcer, userID string, obj string, act string, roles []string, permissions []*structs.ReadPermission, tenantID string) (bool, error) {
	for _, role := range roles {
		ok, err := enforcer.Enforce(role, tenantID, obj, act, nil, nil)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	for _, permission := range permissions {
		ok, err := enforcer.Enforce(userID, tenantID, permission.Subject, permission.Action, nil, nil)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}
