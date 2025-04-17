package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
)

// HasPermission middleware checks if user has the required permission
func HasPermission(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for admin status first (admins have all permissions)
		isAdmin, exists := c.Get("is_admin")
		if exists && isAdmin.(bool) {
			c.Next()
			return
		}

		// Get permissions from context
		permissions, exists := c.Get("permissions")
		if !exists {
			resp.Fail(c.Writer, resp.Forbidden("Permission information not available"))
			c.Abort()
			return
		}

		permissionList, ok := permissions.([]string)
		if !ok {
			resp.Fail(c.Writer, resp.Forbidden("Invalid permission format"))
			c.Abort()
			return
		}

		// Check for exact match
		for _, perm := range permissionList {
			if perm == requiredPermission || perm == "*" {
				c.Next()
				return
			}

			// Handle wildcards in permissions (e.g., "read:*" should match "read:users")
			if strings.HasSuffix(perm, ":*") {
				action := strings.TrimSuffix(perm, ":*")
				if strings.HasPrefix(requiredPermission, action+":") {
					c.Next()
					return
				}
			}

			// Handle wildcards in resource (e.g., "*:users" should match "read:users")
			if strings.HasPrefix(perm, "*:") {
				resource := strings.TrimPrefix(perm, "*:")
				if strings.HasSuffix(requiredPermission, ":"+resource) {
					c.Next()
					return
				}
			}
		}

		logger.Warnf(c.Request.Context(), "Permission denied: %s", requiredPermission)
		resp.Fail(c.Writer, resp.Forbidden("You don't have the required permission"))
		c.Abort()
	}
}

// HasRole middleware checks if user has the required role
func HasRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get roles from context
		roles, exists := c.Get("roles")
		if !exists {
			resp.Fail(c.Writer, resp.Forbidden("Role information not available"))
			c.Abort()
			return
		}

		roleList, ok := roles.([]string)
		if !ok {
			resp.Fail(c.Writer, resp.Forbidden("Invalid role format"))
			c.Abort()
			return
		}

		// Check if user has the required role
		for _, role := range roleList {
			if role == requiredRole {
				c.Next()
				return
			}
		}

		logger.Warnf(c.Request.Context(), "Role denied: %s", requiredRole)
		resp.Fail(c.Writer, resp.Forbidden("You don't have the required role"))
		c.Abort()
	}
}

// HasAnyRole middleware checks if user has any of the required roles
func HasAnyRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get roles from context
		roles, exists := c.Get("roles")
		if !exists {
			resp.Fail(c.Writer, resp.Forbidden("Role information not available"))
			c.Abort()
			return
		}

		roleList, ok := roles.([]string)
		if !ok {
			resp.Fail(c.Writer, resp.Forbidden("Invalid role format"))
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		for _, userRole := range roleList {
			for _, requiredRole := range requiredRoles {
				if userRole == requiredRole {
					c.Next()
					return
				}
			}
		}

		logger.Warnf(c.Request.Context(), "None of required roles found: %v", requiredRoles)
		resp.Fail(c.Writer, resp.Forbidden("You don't have any of the required roles"))
		c.Abort()
	}
}

// IsAdmin middleware checks if user is an admin
func IsAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			resp.Fail(c.Writer, resp.Forbidden("Admin access required"))
			c.Abort()
			return
		}
		c.Next()
	}
}
