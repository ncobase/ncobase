package middleware

import (
	"context"
	"stocms/internal/config"
	"stocms/internal/helper"

	"github.com/gin-gonic/gin"
)

// BindConfig binds config to gin.Context
func BindConfig(c *gin.Context) {
	ctx := helper.SetConfig(c, config.GetConfig())
	c.Request = c.Request.WithContext(ctx)
	c.Next()
}

// BindGinContext binds gin.Context to context.Context
func BindGinContext(c *gin.Context) {
	ctx := helper.WithGinContext(context.Background(), c)
	c.Request = c.Request.WithContext(ctx)
	c.Next()
}
