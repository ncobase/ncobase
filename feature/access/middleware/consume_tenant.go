package middleware

import (
	"context"
	"ncobase/common/consts"
	"ncobase/common/log"
	"ncobase/feature/tenant/service"
	"ncobase/feature/tenant/structs"
	"ncobase/helper"

	"github.com/gin-gonic/gin"
)

// ConsumeTenant consumes tenant information from the request header or user tenants.
func ConsumeTenant(svc *service.Service) gin.HandlerFunc {
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
				if result, _ := svc.UserTenant.UserBelongTenantService(c, userID); result.Code != 0 {
					log.Errorf(context.Background(), "failed to fetch user belong tenant: %v", result)
				} else if readTenant, ok := result.Data.(*structs.ReadTenant); ok {
					tenantID = readTenant.ID
				} else {
					log.Errorf(context.Background(), "failed to parse user belong tenant: %v", result)
				}
			}
		}

		// Set tenant ID to context
		helper.SetTenantID(ctx, tenantID)

		// Continue to next middleware or handler
		c.Next()
	}
}
