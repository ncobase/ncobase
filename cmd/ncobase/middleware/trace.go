package middleware

import (
	"context"
	"net/http"

	"github.com/ncobase/ncore/consts"
	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/logging/observes"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// Trace is a middleware for tracing
func Trace(c *gin.Context) {
	ctx := ctxutil.FromGinContext(c)

	// Check for trace ID in the request header
	traceID := c.GetHeader(consts.TraceKey)

	if traceID == "" {
		ctx, traceID = ctxutil.EnsureTraceID(ctx)
	} else {
		ctx = ctxutil.SetTraceID(ctx, traceID)
	}

	// Update the request context
	c.Request = c.Request.WithContext(ctx)

	// Set trace ID in Gin's context for easy access in handlers
	c.Set(ctxutil.TraceIDKey, traceID)

	// Set trace header in the response
	c.Writer.Header().Set(consts.TraceKey, traceID)

	// Create OpenTelemetry tracing context
	path := c.Request.URL.Path
	if path == "" {
		path = c.FullPath()
	}
	tc := observes.NewTracingContext(ctx, path, 100)
	defer tc.End()

	tc.SetAttributes(
		attribute.String("http.method", c.Request.Method),
		attribute.String("http.path", path),
		attribute.String("trace.id", traceID),
	)

	ctx = context.WithValue(tc.Context(), "tracing_context", tc)
	c.Request = c.Request.WithContext(ctx)

	c.Next()

	// Update OpenTelemetry span with response status
	status := c.Writer.Status()
	tc.SetAttributes(
		attribute.Int("http.status_code", status),
	)
	tc.SetStatus(codes.Code(status), http.StatusText(status))
}

// OtelTrace is a middleware for OpenTelemetry trace
func OtelTrace(c *gin.Context) {
	ctx := ctxutil.FromGinContext(c)
	path := c.Request.URL.Path
	if path == "" {
		path = c.FullPath()
	}

	traceID := ctxutil.GetTraceID(ctx)
	tc := observes.NewTracingContext(ctx, path, 100)
	defer tc.End()

	tc.SetAttributes(
		attribute.String("http.method", c.Request.Method),
		attribute.String("http.path", path),
		attribute.String("trace.id", traceID),
	)

	ctx = context.WithValue(tc.Context(), "tracing_context", tc)
	c.Request = c.Request.WithContext(ctx)

	c.Next()

	status := c.Writer.Status()
	tc.SetAttributes(
		attribute.Int("http.status_code", status),
	)
	tc.SetStatus(codes.Code(status), http.StatusText(status))
}
