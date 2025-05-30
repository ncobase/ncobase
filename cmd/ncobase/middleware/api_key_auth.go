package middleware

import (
	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/net/resp"

	"github.com/gin-gonic/gin"
)

// ApiKeyAuth middleware for API key authentication
func ApiKeyAuth(sm *ServiceManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if already authenticated
		if ctxutil.GetUserID(c.Request.Context()) != "" {
			c.Next()
			return
		}

		// Get API key from header or query parameter
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		// If no API key, proceed to next middleware
		if apiKey == "" {
			c.Next()
			return
		}

		// Get user service
		usw := sm.UserServiceWrapper()
		// Validate API key
		key, err := usw.ValidateApiKey(c.Request.Context(), apiKey)
		if err != nil {
			resp.Fail(c.Writer, resp.UnAuthorized("Invalid API key"))
			c.Abort()
			return
		}

		// Set user ID in context
		ctx := ctxutil.SetUserID(c.Request.Context(), key.UserID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
