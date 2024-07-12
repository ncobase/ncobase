package middleware

import (
	"context"
	"ncobase/common/consts"
	"ncobase/common/log"
	tenantService "ncobase/feature/tenant/service"
	"ncobase/helper"

	"github.com/gin-gonic/gin"
)

// ConsumeTenant consumes tenant information from the request header or user tenants.
func ConsumeTenant(ts *tenantService.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := helper.FromGinContext(c)
		// Retrieve tenant ID from request header
		tenantID := c.GetHeader(consts.XMdTenantKey)
		// If tenant ID is not provided in the header, try to fetch from other sources
		if tenantID == "" {
			// Get tenant ID
			tenantID = helper.GetTenantID(ctx)
			if tenantID == "" {
				// Get user ID
				userID := helper.GetUserID(ctx)
				// Fetch user tenants
				if tenant, err := ts.UserTenant.UserBelongTenant(c, userID); err != nil {
					log.Errorf(context.Background(), "failed to fetch user belong tenant: %v", err.Error())
				} else if tenant != nil {
					tenantID = tenant.ID
				} else {
					log.Errorf(context.Background(), "failed to parse user belong tenant: %v", err.Error())
				}
			}
		}

		// Set tenant ID to context
		helper.SetTenantID(ctx, tenantID)

		// Continue to next middleware or handler
		c.Next()
	}
}
