package middleware

import (
	"ncobase/ncore/ecode"
	"ncobase/ncore/helper"
	"ncobase/ncore/logger"
	"ncobase/ncore/resp"
	"ncobase/ncore/validator"

	"github.com/gin-gonic/gin"
)

// AuthenticatedTenant checks if the user is related to a tenant and authenticated.
func AuthenticatedTenant(c *gin.Context) {
	ctx := helper.FromGinContext(c)
	tenantID := helper.GetTenantID(ctx)

	if validator.IsEmpty(tenantID) {
		logger.Warn(ctx, "Tenant authentication failed")
		resp.Fail(c.Writer, resp.UnAuthorized(ecode.Text(ecode.Unauthorized)))
		c.Abort()
		return
	}

	c.Next()
}

// AuthenticatedUser checks if the user is authenticated.
func AuthenticatedUser(c *gin.Context) {
	ctx := helper.FromGinContext(c)
	userID := helper.GetUserID(ctx)

	if validator.IsEmpty(userID) {
		logger.Warn(ctx, "User authentication failed")
		resp.Fail(c.Writer, resp.UnAuthorized(ecode.Text(ecode.Unauthorized)))
		c.Abort()
		return
	}

	c.Next()
}
