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
func ConsumeTenant(ts *tenantService.Service, whiteList []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if inWhiteList(c.Request.URL.Path, whiteList) {
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
				// Fetch user tenants
				tenant, err := ts.UserTenant.UserBelongTenant(c, userID)
				if err != nil {
					log.Errorf(context.Background(), "failed to fetch user belong tenant: %v", err.Error())
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
