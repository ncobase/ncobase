package middleware

import (
	"context"
	"ncobase/common/config"
	"ncobase/helper"

	"github.com/gin-gonic/gin"
)

// BindConfig binds the application configuration to the Gin context.
func BindConfig(c *gin.Context) {
	// Get the application configuration
	appConfig := config.GetConfig()

	// Set the configuration in the Gin context
	ctx := helper.SetConfig(c, appConfig)

	// Update the request context with the new context
	c.Request = c.Request.WithContext(ctx)

	// call the next handler
	c.Next()
}

// BindGinContext binds the Gin context to the standard context.Context.
func BindGinContext(c *gin.Context) {
	// Wrap the Gin context with a standard context
	ctx := helper.WithGinContext(context.Background(), c)

	// Update the request context with the new context
	c.Request = c.Request.WithContext(ctx)

	// call the next handler
	c.Next()
}
