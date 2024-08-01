package middleware

import (
	"context"
	"ncobase/common/consts"
	"ncobase/common/log"
	"ncobase/common/observes"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// Trace is a middleware for tracing
func Trace(c *gin.Context) {
	ctx := c.Request.Context()

	// Check for trace ID in the request header
	traceID := c.GetHeader(consts.TraceKey)

	if traceID == "" {
		ctx, traceID = log.EnsureTraceID(ctx)
	} else {
		ctx = log.SetTraceID(ctx, traceID)
	}

	// Update the request context
	c.Request = c.Request.WithContext(ctx)

	// Set trace ID in Gin's context for easy access in handlers
	c.Set(log.TraceIDKey, traceID)

	// Set trace header in the response
	c.Writer.Header().Set(consts.TraceKey, traceID)

	c.Next()
}

// OtelTrace is a middleware for OpenTelemetry trace
func OtelTrace(c *gin.Context) {
	path := c.Request.URL.Path
	if path == "" {
		path = c.FullPath()
	}
	tc := observes.NewTracingContext(c.Request.Context(), path, 100)
	defer tc.End()

	tc.SetAttributes(
		attribute.String("http.method", c.Request.Method),
		attribute.String("http.path", path),
	)

	c.Request = c.Request.WithContext(context.WithValue(tc.Context(), "tracing_context", tc))

	c.Next()

	status := c.Writer.Status()
	tc.SetAttributes(
		attribute.Int("http.status_code", status),
	)
	tc.SetStatus(codes.Code(status), http.StatusText(status))
}
