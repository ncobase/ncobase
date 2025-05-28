package middleware

import (
	tenantService "ncobase/tenant/service"

	"github.com/ncobase/ncore/consts"
	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/logging/logger"

	"github.com/gin-gonic/gin"
)

// ConsumeTenant consumes tenant information from the request header or user tenants.
func ConsumeTenant(ts *tenantService.Service, whiteList []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipPath(c.Request, whiteList) {
			c.Next()
			return
		}

		// Get context
		ctx := c.Request.Context()
		// Retrieve user ID from context
		userID := ctxutil.GetUserID(ctx)

		// Retrieve tenant ID from request header
		tenantID := c.GetHeader(consts.TenantKey)

		// If tenant ID is provided in header, validate it belongs to user
		if tenantID != "" && userID != "" {
			isValid, err := ts.UserTenant.IsTenantInUser(ctx, userID, tenantID)
			if err != nil || !isValid {
				logger.Warnf(ctx, "Tenant %s does not belong to user %s", tenantID, userID)
				tenantID = ""
			}
		}

		// If tenant ID is not provided or invalid, try to fetch from other sources
		if tenantID == "" && userID != "" {
			// Get tenant ID from context
			tenantID = ctxutil.GetTenantID(ctx)

			// If still not found, try to fetch from user tenants
			if tenantID == "" {
				logger.Info(ctx, "tenant not found in header or context, trying to fetch from user tenants")

				// First try to get default tenant
				tenant, err := ts.UserTenant.UserBelongTenant(ctx, userID)
				if err != nil {
					logger.Warnf(ctx, "failed to fetch user default tenant: %v", err.Error())

					// If no default tenant, try to get any tenant user belongs to
					tenants, err := ts.UserTenant.UserBelongTenants(ctx, userID)
					if err == nil && len(tenants) > 0 {
						tenant = tenants[0]
					}
				}

				if tenant != nil {
					tenantID = tenant.ID
				}
			}
		}

		// Set tenant ID to context if it exists
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
