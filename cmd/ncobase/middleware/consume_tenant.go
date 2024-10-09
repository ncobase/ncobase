package middleware

import (
	"ncobase/common/consts"
	"ncobase/common/helper"
	"ncobase/common/log"
	tenantService "ncobase/feature/tenant/service"

	"github.com/gin-gonic/gin"
)

// ConsumeTenant consumes tenant information from the request header or user tenants.
func ConsumeTenant(ts *tenantService.Service, whiteList []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipPath(c.Request.URL.Path, whiteList) {
			c.Next()
			return
		}
		ctx := c.Request.Context()
		// Retrieve user ID from context
		userID := helper.GetUserID(ctx)
		// Retrieve tenant ID from request header
		tenantID := c.GetHeader(consts.TenantKey)
		// If tenant ID is not provided in the header, try to fetch from other sources
		if tenantID == "" && userID != "" {
			// Get tenant ID
			tenantID = helper.GetTenantID(ctx)
			if tenantID == "" {
				log.Warn(ctx, "tenant not found, try to fetch from user tenants")
				// Fetch user tenants
				tenant, err := ts.UserTenant.UserBelongTenant(c, userID)
				if err != nil {
					log.Errorf(ctx, "failed to fetch user belong tenant: %v", err.Error())
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
