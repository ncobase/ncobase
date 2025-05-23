package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
)

// HasPermission middleware checks if user has the required permission
func HasPermission(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Check for admin status first (admins have all permissions)
		if ctxutil.GetUserIsAdmin(ctx) {
			c.Next()
			return
		}

		// Get permissions from context
		permissions := ctxutil.GetUserPermissions(ctx)
		if len(permissions) == 0 {
			logger.Warnf(ctx, "User has no permissions")
			resp.Fail(c.Writer, resp.Forbidden("Permission information not available"))
			c.Abort()
			return
		}

		// Check for wildcard permission first
		if hasWildcardPermission(permissions) {
			c.Next()
			return
		}

		// Check for specific permission
		if hasSpecificPermission(permissions, requiredPermission) {
			c.Next()
			return
		}

		logger.Warnf(ctx, "Permission denied: %s", requiredPermission)
		resp.Fail(c.Writer, resp.Forbidden("You don't have the required permission"))
		c.Abort()
	}
}

// HasRole middleware checks if user has the required role
func HasRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Get roles from context
		roles := ctxutil.GetUserRoles(ctx)
		if len(roles) == 0 {
			logger.Warnf(ctx, "User has no roles")
			resp.Fail(c.Writer, resp.Forbidden("Role information not available"))
			c.Abort()
			return
		}

		// Check for admin roles first
		if hasAdminRole(roles) {
			c.Next()
			return
		}

		// Check for specific role
		if hasSpecificRole(roles, requiredRole) {
			c.Next()
			return
		}

		logger.Warnf(ctx, "Role denied: %s", requiredRole)
		resp.Fail(c.Writer, resp.Forbidden("You don't have the required role"))
		c.Abort()
	}
}

// HasAnyRole middleware checks if user has any of the required roles
func HasAnyRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		roles := ctxutil.GetUserRoles(ctx)
		if len(roles) == 0 {
			resp.Fail(c.Writer, resp.Forbidden("Role information not available"))
			c.Abort()
			return
		}

		// Check for admin roles first
		if hasAdminRole(roles) {
			c.Next()
			return
		}

		// Check for any of the required roles
		for _, userRole := range roles {
			for _, requiredRole := range requiredRoles {
				if userRole == requiredRole {
					c.Next()
					return
				}
			}
		}

		logger.Warnf(ctx, "None of required roles found: %v", requiredRoles)
		resp.Fail(c.Writer, resp.Forbidden("You don't have any of the required roles"))
		c.Abort()
	}
}

// HasAnyPermission middleware checks if user has any of the required permissions
func HasAnyPermission(requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		permissions := ctxutil.GetUserPermissions(ctx)
		if len(permissions) == 0 {
			resp.Fail(c.Writer, resp.Forbidden("Permission information not available"))
			c.Abort()
			return
		}

		// Check for admin status first
		if ctxutil.GetUserIsAdmin(ctx) {
			c.Next()
			return
		}

		// Check for wildcard permission
		if hasWildcardPermission(permissions) {
			c.Next()
			return
		}

		// Check for any of the required permissions
		for _, requiredPermission := range requiredPermissions {
			if hasSpecificPermission(permissions, requiredPermission) {
				c.Next()
				return
			}
		}

		logger.Warnf(ctx, "None of required permissions found: %v", requiredPermissions)
		resp.Fail(c.Writer, resp.Forbidden("You don't have any of the required permissions"))
		c.Abort()
	}
}

// IsAdmin middleware checks if user is an admin
func IsAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		if !ctxutil.GetUserIsAdmin(ctx) {
			resp.Fail(c.Writer, resp.Forbidden("Admin access required"))
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireManagerOrAbove requires user to have manager level or above
func RequireManagerOrAbove() gin.HandlerFunc {
	managerRoles := []string{
		"super-admin", "system-admin", "enterprise-admin", "tenant-admin",
		"hr-manager", "finance-manager", "it-manager",
		"department-manager", "team-leader", "project-manager",
		"technical-lead", "qa-manager", "customer-service-manager",
		"enterprise-executive", "company-director", "department-head",
	}

	return HasAnyRole(managerRoles...)
}

// RequireEmployeeOrAbove requires user to be employee level or above (excludes external users)
func RequireEmployeeOrAbove() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		roles := ctxutil.GetUserRoles(ctx)

		// Excluded external roles
		excludedRoles := []string{"contractor", "consultant", "intern", "auditor"}

		// Check if user has any excluded role
		for _, role := range roles {
			for _, excluded := range excludedRoles {
				if role == excluded {
					resp.Fail(c.Writer, resp.Forbidden("Access denied: external users not allowed"))
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// OwnerOrManager checks if user is resource owner or has management role
func OwnerOrManager(getOwnerID func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		userID := ctxutil.GetUserID(ctx)

		// Check if user is admin
		if ctxutil.GetUserIsAdmin(ctx) {
			c.Next()
			return
		}

		// Check if user is resource owner
		ownerID := getOwnerID(c)
		if userID == ownerID {
			c.Next()
			return
		}

		// Check if user has management role
		roles := ctxutil.GetUserRoles(ctx)
		if hasManagementRole(roles) {
			c.Next()
			return
		}

		resp.Fail(c.Writer, resp.Forbidden("Access denied: not owner or manager"))
		c.Abort()
	}
}

// hasWildcardPermission checks if user has wildcard permission
func hasWildcardPermission(permissions []string) bool {
	for _, perm := range permissions {
		if perm == "*:*" {
			return true
		}
	}
	return false
}

// hasSpecificPermission checks if user has specific permission
func hasSpecificPermission(permissions []string, required string) bool {
	for _, perm := range permissions {
		if perm == required {
			return true
		}
		// Support wildcard matching
		if strings.Contains(perm, "*") && matchesWildcard(perm, required) {
			return true
		}
	}
	return false
}

// hasAdminRole checks if user has admin role
func hasAdminRole(roles []string) bool {
	adminRoles := []string{"super-admin", "system-admin", "enterprise-admin"}
	for _, role := range roles {
		for _, adminRole := range adminRoles {
			if role == adminRole {
				return true
			}
		}
	}
	return false
}

// hasSpecificRole checks if user has specific role
func hasSpecificRole(roles []string, required string) bool {
	for _, role := range roles {
		if role == required {
			return true
		}
	}
	return false
}

// hasManagementRole checks if user has management role/
func hasManagementRole(roles []string) bool {
	managementRoles := []string{
		"department-manager", "team-leader", "project-manager",
		"hr-manager", "finance-manager", "it-manager",
		"technical-lead", "qa-manager", "customer-service-manager",
	}

	for _, role := range roles {
		for _, mgmtRole := range managementRoles {
			if role == mgmtRole {
				return true
			}
		}
	}
	return false
}

// matchesWildcard checks if pattern matches target
func matchesWildcard(pattern, target string) bool {
	// Simple wildcard matching implementation
	// e.g.: "read:*" matches "read:user", "read:employee" etc.
	parts := strings.Split(pattern, ":")
	targetParts := strings.Split(target, ":")

	if len(parts) != len(targetParts) {
		return false
	}

	for i, part := range parts {
		if part != "*" && part != targetParts[i] {
			return false
		}
	}

	return true
}
