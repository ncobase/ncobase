package middleware

import (
	"context"
	"ncobase/common/log"
	"ncobase/common/observes"
	"ncobase/helper"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func Trace(c *gin.Context) {
	ctx := c.Request.Context()
	// Get the trace ID from the request
	traceID := helper.GetTraceID(ctx)

	// If trace ID is not present in the request, generate a new one
	if traceID == "" {
		traceID = helper.NewTraceID()
		// Set the trace ID in the request context
		c.Request = c.Request.WithContext(log.NewTraceIDContext(ctx, traceID))
	}

	// Set trace header in the response
	c.Writer.Header().Set("X-Trace-ID", traceID)

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
