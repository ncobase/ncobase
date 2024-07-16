package middleware

import (
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/validator"
	"ncobase/helper"

	"github.com/gin-gonic/gin"
)

// AuthenticatedTenant is a middleware that checks if the user is related to a tenant and authenticated.
func AuthenticatedTenant(c *gin.Context) {
	// Retrieve the context.Context from *gin.Context
	ctx := c.Request.Context()

	// Retrieve tenant ID from the context
	tenantID := helper.GetTenantID(ctx)

	// Check if tenant ID is empty
	if validator.IsEmpty(tenantID) {
		// Respond with unauthorized error
		resp.Fail(c.Writer, resp.UnAuthorized(ecode.Text(ecode.Unauthorized)))
		c.Abort()
	}

	// Proceed to the next handler if the user is authenticated
	c.Next()
}
