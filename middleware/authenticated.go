package middleware

import (
	"ncobase/common/validator"
	"net/http"

	"ncobase/common/ecode"
	"ncobase/helper"

	"github.com/gin-gonic/gin"
)

// Authenticated is a middleware that checks if the user is authenticated.
func Authenticated(c *gin.Context) {
	// Retrieve the context.Context from *gin.Context
	ctx := helper.FromGinContext(c)

	// Retrieve user ID and tenant ID from the context
	userID := helper.GetUserID(ctx)
	tenantID := helper.GetTenantID(ctx)

	// Check if user ID or tenant ID is empty
	if validator.IsEmpty(userID) || validator.IsEmpty(tenantID) {
		// Respond with unauthorized error
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"code":    ecode.Unauthorized,
			"message": ecode.Text(ecode.Unauthorized),
		})
		return
	}

	// Proceed to the next handler if the user is authenticated
	c.Next()
}
