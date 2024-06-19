package middleware

import (
	"ncobase/internal/helper"
	"net/http"

	"github.com/ncobase/common/ecode"
	"github.com/ncobase/common/validator"

	"github.com/gin-gonic/gin"
)

// Authorized middleware verifies the existence of a user.
func Authorized(c *gin.Context) {
	// Retrieve user ID from the context
	userID := helper.GetUserID(c)
	// Retrieve tenant from the context
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

	// Continue to the next handler
	c.Next()
}
