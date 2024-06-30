package helper

import (
	"context"

	"github.com/gin-gonic/gin"
)

// FromGinContext extracts the context.Context from *gin.Context.
func FromGinContext(c *gin.Context) context.Context {
	return c.Request.Context()
}

// WithGinContext returns a context.Context that embeds the *gin.Context.
func WithGinContext(ctx context.Context, c *gin.Context) context.Context {
	return context.WithValue(ctx, "ginContext", c)
}

// GetGinContext extracts *gin.Context from context.Context if it exists.
func GetGinContext(ctx context.Context) (*gin.Context, bool) {
	if c, ok := ctx.Value("ginContext").(*gin.Context); ok {
		return c, ok
	}
	return nil, false
}

// GetValue retrieves a value from the context.
func GetValue(c *gin.Context, key string) any {
	// if c, ok := GetGinContext(ctx); ok {
	if val, exists := c.Get(key); exists {
		return val
	}
	// }
	return nil
}

// SetValue sets a value to the context.
func SetValue(c *gin.Context, key string, val any) context.Context {
	// if c, ok := GetGinContext(ctx); ok {
	c.Set(key, val)
	// }
	return context.WithValue(c, key, val)
}
