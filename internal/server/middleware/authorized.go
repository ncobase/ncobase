package middleware

import (
	"net/http"
	"stocms/internal/helper"
	"stocms/pkg/ecode"
	"stocms/pkg/validator"

	"github.com/gin-gonic/gin"
)

// Authorized middleware verifies the existence of a user.
func Authorized(c *gin.Context) {
	// Retrieve user ID from the context
	userID := helper.GetUserID(c)

	// Check if user ID is empty
	if validator.IsEmpty(userID) {
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
