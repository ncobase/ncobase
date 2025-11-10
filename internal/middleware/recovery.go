package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"

	"github.com/gin-gonic/gin"
)

// Recovery returns a middleware that recovers from panics and logs the error
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				ctx := ctxutil.FromGinContext(c)

				// Get stack trace
				stack := string(debug.Stack())

				// Log the panic with full context
				logger.Errorf(ctx,
					"PANIC RECOVERED: %v\nRequest: %s %s\nClient IP: %s\nUser-Agent: %s\nStack Trace:\n%s",
					err,
					c.Request.Method,
					c.Request.URL.Path,
					c.ClientIP(),
					c.Request.UserAgent(),
					stack,
				)

				// Get user context if available
				userID := ctxutil.GetUserID(ctx)
				username := ctxutil.GetUsername(ctx)
				if userID != "" {
					logger.Errorf(ctx, "Panic occurred for user: %s (ID: %s)", username, userID)
				}

				// Check if response was already written
				if !c.Writer.Written() {
					// Return error response
					resp.Fail(c.Writer, resp.InternalServer("Internal server error. The issue has been logged and will be investigated."))
				}

				// Abort the request
				c.Abort()
			}
		}()

		c.Next()
	}
}

// RecoveryWithCustomHandler returns a middleware with a custom panic handler
func RecoveryWithCustomHandler(handler func(c *gin.Context, err interface{})) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				ctx := ctxutil.FromGinContext(c)

				// Log the panic
				stack := string(debug.Stack())
				logger.Errorf(ctx,
					"PANIC RECOVERED: %v\nStack Trace:\n%s",
					err,
					stack,
				)

				// Call custom handler
				if handler != nil {
					handler(c, err)
				}

				// Abort the request
				c.Abort()
			}
		}()

		c.Next()
	}
}

// safeHandler wraps a gin.HandlerFunc to catch panics within the handler itself
func safeHandler(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				ctx := ctxutil.FromGinContext(c)

				// Log panic
				logger.Errorf(ctx, "Handler panic: %v\nStack: %s", err, string(debug.Stack()))

				// Return error
				if !c.Writer.Written() {
					resp.Fail(c.Writer, resp.InternalServer(fmt.Sprintf("Handler error: %v", err)))
				}

				c.Abort()
			}
		}()

		handler(c)
	}
}

// SafeHandler is a public wrapper for safeHandler
func SafeHandler(handler gin.HandlerFunc) gin.HandlerFunc {
	return safeHandler(handler)
}
