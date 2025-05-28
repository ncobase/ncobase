package middleware

import (
	"github.com/ncobase/ncore/consts"
	"github.com/ncobase/ncore/ctxutil"
	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"

	"github.com/gin-gonic/gin"
)

// ConsumeTenant consumes tenant information from request header or user tenants
func ConsumeTenant(em ext.ManagerInterface, whiteList []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipPath(c.Request, whiteList) {
			c.Next()
			return
		}

		ctx := c.Request.Context()
		userID := ctxutil.GetUserID(ctx)
		tenantID := c.GetHeader(consts.TenantKey)

		// Get service manager
		sw := GetServiceManager(em)
		// Get tenant wrapper
		tsw := sw.Tenant()

		// Validate tenant ID belongs to user if both provided
		if tenantID != "" && userID != "" {
			if isValid, err := tsw.IsTenantInUser(ctx, userID, tenantID); err != nil || !isValid {
				logger.Warnf(ctx, "Tenant %s does not belong to user %s", tenantID, userID)
				tenantID = ""
			}
		}

		// Get tenant from context or user tenants if not provided/invalid
		if tenantID == "" && userID != "" {
			tenantID = ctxutil.GetTenantID(ctx)

			if tenantID == "" {
				logger.Info(ctx, "tenant not found in header or context, trying to fetch from user tenants")

				// Try to get default tenant first
				if tenant, err := tsw.GetUserDefaultTenant(ctx, userID); err == nil && tenant != nil {
					tenantID = tenant.ID
				} else {
					// Get any tenant user belongs to
					if tenants, err := tsw.GetUserTenants(ctx, userID); err == nil && len(tenants) > 0 {
						tenantID = tenants[0].ID
					}
				}
			}
		}

		// Set tenant ID to context if exists
		if tenantID != "" {
			logger.Infof(ctx, "Setting tenant ID: %s for user: %s", tenantID, userID)
			ctx = ctxutil.SetTenantID(ctx, tenantID)
			c.Request = c.Request.WithContext(ctx)
		} else if userID != "" {
			logger.Warnf(ctx, "No tenant found for user: %s", userID)
		}

		c.Next()
	}
}
