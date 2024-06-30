package middleware

import (
	"context"
	"ncobase/common/log"
	"ncobase/helper"
	"net/http"
	"strings"

	"ncobase/app/data/structs"
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/util"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

// relatedService represents the related service interface
type relatedService interface {
	GetUserRolesInTenantService(ctx context.Context, u string, t string) (*resp.Exception, error)
	GetUserRoleByUserIDService(ctx context.Context, u string) (*resp.Exception, error)
	GetRolePermissionsService(ctx context.Context, r string) (*resp.Exception, error)
}

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

// Authorized is a middleware that checks if the user is authorized to access the resource.
func Authorized(enforcer *casbin.Enforcer, whiteList []string, svc relatedService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if inWhiteList(c.Request.URL.Path, whiteList) {
			c.Next()
			return
		}

		currentUser := helper.GetUserID(c)
		currentTenant := helper.GetTenantID(c)
		obj := c.Request.URL.Path
		act := c.Request.Method

		log.Infof(c, "userID: %s, tenantID: %s, obj: %s, act: %s\n", currentUser, currentTenant, obj, act)

		// Retrieve user roles from service
		exception, err := svc.GetUserRoleByUserIDService(c, currentUser)
		handleException(c, exception, err, "Error retrieving user roles")
		if c.IsAborted() {
			return
		}

		userRoles := exception.Data.([]*structs.ReadRole)
		if len(userRoles) == 0 {
			handleException(c, resp.Forbidden("User has no roles, please contact your administrator"), nil, "")
			return
		}

		var roles []string
		for _, r := range userRoles {
			roles = append(roles, r.Slug)
		}

		// Query current user roles of the current tenant
		exception, err = svc.GetUserRolesInTenantService(c, currentUser, currentTenant)
		handleException(c, exception, err, "Error retrieving user roles in tenant")
		if c.IsAborted() {
			return
		}

		tenantRoles := exception.Data.([]*structs.ReadRole)
		for _, r := range tenantRoles {
			roles = append(roles, r.Slug)
		}

		roles = util.RemoveDuplicates(roles)
		var permissions []*structs.ReadPermission

		// Retrieve role permissions from service
		for _, role := range roles {
			exception, err = svc.GetRolePermissionsService(c, role)
			handleException(c, exception, err, "Error retrieving role permissions")
			if c.IsAborted() {
				return
			}

			rolePermissions := exception.Data.([]*structs.ReadPermission)
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
