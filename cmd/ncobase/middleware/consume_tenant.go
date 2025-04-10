package middleware

import (
	"github.com/ncobase/ncore/pkg/consts"
	"github.com/ncobase/ncore/pkg/helper"
	"github.com/ncobase/ncore/pkg/logger"
	tenantService "ncobase/core/tenant/service"

	"github.com/gin-gonic/gin"
)

// ConsumeTenant consumes tenant information from the request header or user tenants.
func ConsumeTenant(ts *tenantService.Service, whiteList []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipPath(c.Request, whiteList) {
			c.Next()
			return
		}
		ctx := helper.FromGinContext(c)
		// Retrieve user ID from context
		userID := helper.GetUserID(ctx)
		// Retrieve tenant ID from request header
		tenantID := c.GetHeader(consts.TenantKey)
		// If tenant ID is not provided in the header, try to fetch from other sources
		if tenantID == "" && userID != "" {
			// Get tenant ID
			tenantID = helper.GetTenantID(ctx)
			if tenantID == "" {
				logger.Warn(ctx, "tenant not found, try to fetch from user tenants")
				// Fetch user tenants
				tenant, err := ts.UserTenant.UserBelongTenant(c, userID)
				if err != nil {
					logger.Errorf(ctx, "failed to fetch user belong tenant: %v", err.Error())
				}
				if tenant != nil {
					tenantID = tenant.ID
				}
			}
		}

		// Set tenant ID to context if it exists
		if tenantID != "" {
			helper.SetTenantID(ctx, tenantID)
		}

		// Continue to next middleware or handler
		c.Next()
	}
}
