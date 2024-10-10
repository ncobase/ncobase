package middleware

import (
	"ncobase/common/consts"
	"ncobase/common/helper"
	"ncobase/common/log"
	"ncobase/common/tracing"
	tenantService "ncobase/feature/tenant/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ConsumeTenant consumes tenant information from the request header or user tenants.
func ConsumeTenant(ts *tenantService.Service, whiteList []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipPath(c.Request.URL.Path, whiteList) {
			c.Next()
			return
		}
		ctx := helper.FromGinContext(c)
		contextTraceID := helper.GetTraceID(ctx)
		// Retrieve user ID from context
		userID := helper.GetUserID(ctx)
		// Retrieve tenant ID from request header
		tenantID := c.GetHeader(consts.TenantKey)
		// If tenant ID is not provided in the header, try to fetch from other sources
		if tenantID == "" && userID != "" {
			// Get tenant ID
			tenantID = helper.GetTenantID(ctx)
			if tenantID == "" {
				log.EntryWithFields(ctx, logrus.Fields{
					tracing.TraceIDKey: contextTraceID,
				}).Warn("tenant not found, try to fetch from user tenants")
				// Fetch user tenants
				tenant, err := ts.UserTenant.UserBelongTenant(c, userID)
				if err != nil {
					log.EntryWithFields(ctx, logrus.Fields{
						tracing.TraceIDKey: contextTraceID,
					}).Errorf("failed to fetch user belong tenant: %v", err.Error())
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
