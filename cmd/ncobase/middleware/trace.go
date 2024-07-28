package middleware

import (
	"context"
	"ncobase/common/observes"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

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
