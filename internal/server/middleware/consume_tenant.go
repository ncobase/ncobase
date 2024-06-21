package middleware

import (
	"context"
	"ncobase/internal/data/structs"
	"ncobase/internal/helper"

	"ncobase/common/consts"
	"ncobase/common/log"
	"ncobase/common/resp"

	"github.com/gin-gonic/gin"
)

type TenantFetcher interface {
	UserTenantService(c *gin.Context, username string) (*resp.Exception, error)
}

// ConsumeTenant consumes tenant information from the request header or user tenants.
func ConsumeTenant(svc TenantFetcher) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve tenant ID from request header
		tenantID := c.GetHeader(consts.XMdTenantKey)
		// If tenant ID is not provided in the header, try to fetch from other sources
		if tenantID == "" {
			// Get tenant ID
			tenantID = helper.GetTenantID(c)
			if tenantID == "" {
				// Get user ID
				userID := helper.GetUserID(c)
				// Fetch user tenants
				result, _ := svc.UserTenantService(c, userID)
				if result.Code != 0 {
					log.Errorf(context.Background(), "Failed to fetch user tenants: %v", result)
				} else if readTenant, ok := result.Data.(*structs.ReadTenant); ok {
					tenantID = readTenant.ID
				} else {
					log.Errorf(context.Background(), "Failed to fetch user tenants: %v", result)
				}
			}
		}

		// Set tenant ID to context
		helper.SetTenantID(c, tenantID)

		// Continue to next middleware or handler
		c.Next()
	}
}
