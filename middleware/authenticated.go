package middleware

import (
	"ncobase/common/validator"
	"ncobase/helper"
	"net/http"

	"ncobase/common/ecode"

	"github.com/gin-gonic/gin"
)

// Authenticated is a middleware that checks if the user is authenticated.
func Authenticated(c *gin.Context) {
	// Retrieve user ID and tenant ID from the context
	userID := helper.GetUserID(c)
	tenantID := helper.GetTenantID(c)

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
