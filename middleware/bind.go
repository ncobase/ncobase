package middleware

import (
	"ncobase/common/config"
	"ncobase/helper"

	"github.com/gin-gonic/gin"
)

// BindConfig binds the application configuration to the Gin context.
func BindConfig(c *gin.Context) {
	// Get the application configuration
	appConfig := config.GetConfig()

	// Set the configuration in the context
	ctx := helper.SetValue(c.Request.Context(), "config", appConfig)

	// Update the request context with the new context
	c.Request = c.Request.WithContext(ctx)

	// Call the next handler
	c.Next()
}

// BindGinContext binds the Gin context to the standard context.Context.
func BindGinContext(c *gin.Context) {
	// Wrap the Gin context with a standard context
	ctx := helper.WithGinContext(c.Request.Context(), c)

	// Update the request context with the new context
	c.Request = c.Request.WithContext(ctx)

	// Call the next handler
	c.Next()
}
