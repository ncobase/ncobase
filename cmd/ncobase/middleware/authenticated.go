package middleware

import (
	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/gin-gonic/gin"
)

// AuthenticatedTenant checks if user is related to tenant and authenticated
func AuthenticatedTenant(c *gin.Context) {
	// Get context
	ctx := c.Request.Context()
	// Retrieve tenant ID from context
	tenantID := ctxutil.GetTenantID(ctx)

	if validator.IsEmpty(tenantID) {
		logger.Warn(ctx, "Tenant authentication failed")
		resp.Fail(c.Writer, resp.UnAuthorized(ecode.Text(ecode.Unauthorized)))
		c.Abort()
		return
	}

	c.Next()
}

// AuthenticatedUser checks if user is authenticated
func AuthenticatedUser(c *gin.Context) {
	// Get context
	ctx := c.Request.Context()
	// Retrieve user ID from context
	userID := ctxutil.GetUserID(ctx)

	if validator.IsEmpty(userID) {
		logger.Warn(ctx, "User authentication failed")
		resp.Fail(c.Writer, resp.UnAuthorized(ecode.Text(ecode.Unauthorized)))
		c.Abort()
		return
	}

	c.Next()
}
