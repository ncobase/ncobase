package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ncobase/ncore/consts"
	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/logging/observes"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// Trace middleware for request tracing and context setup
func Trace(c *gin.Context) {
	// Create context with Gin context embedded
	ctx := ctxutil.WithGinContext(c.Request.Context(), c)

	// Handle trace ID
	traceID := c.GetHeader(consts.TraceKey)
	if traceID == "" {
		ctx, traceID = ctxutil.EnsureTraceID(ctx)
	} else {
		ctx = ctxutil.SetTraceID(ctx, traceID)
	}

	// Set client information in context
	ctx = setClientInfoToContext(ctx, c)

	// Update request context - this is crucial!
	c.Request = c.Request.WithContext(ctx)

	// Set trace ID in Gin's context for easy access
	c.Set(ctxutil.TraceIDKey, traceID)

	// Set trace header in response
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
		attribute.String("client.ip", c.ClientIP()),
		attribute.String("user.agent", c.GetHeader("User-Agent")),
	)

	ctx = context.WithValue(tc.Context(), "tracing_context", tc)
	c.Request = c.Request.WithContext(ctx)

	c.Next()

	// Update span with response status
	status := c.Writer.Status()
	tc.SetAttributes(
		attribute.Int("http.status_code", status),
	)
	tc.SetStatus(codes.Code(status), http.StatusText(status))
}

// setClientInfoToContext sets client information to context
func setClientInfoToContext(ctx context.Context, c *gin.Context) context.Context {
	// Extract client information
	ip := extractClientIP(c)
	userAgent := c.GetHeader("User-Agent")
	referer := c.GetHeader("Referer")
	sessionID := extractSessionID(c)

	// Set to context
	ctx = ctxutil.SetClientIP(ctx, ip)
	ctx = ctxutil.SetUserAgent(ctx, userAgent)
	ctx = ctxutil.SetSessionID(ctx, sessionID)
	ctx = ctxutil.SetReferer(ctx, referer)

	// Also set HTTP request for direct access
	ctx = ctxutil.SetHTTPRequest(ctx, c.Request)

	return ctx
}

// extractClientIP extracts real client IP from various headers
func extractClientIP(c *gin.Context) string {
	// Priority order for IP extraction
	headers := []string{
		"X-Forwarded-For",
		"X-Real-IP",
		"CF-Connecting-IP",
		"X-Client-IP",
		"X-Cluster-Client-IP",
	}

	for _, header := range headers {
		ip := c.GetHeader(header)
		if ip != "" && ip != "unknown" {
			// Handle X-Forwarded-For which may contain multiple IPs
			if header == "X-Forwarded-For" {
				ips := strings.Split(ip, ",")
				if len(ips) > 0 {
					cleanIP := strings.TrimSpace(ips[0])
					if cleanIP != "" && cleanIP != "unknown" {
						return cleanIP
					}
				}
			} else {
				return ip
			}
		}
	}

	// Fallback to Gin's ClientIP method
	return c.ClientIP()
}

// OtelTrace middleware for OpenTelemetry trace (updated)
func OtelTrace(c *gin.Context) {
	// Use the context from previous middleware (should contain Gin context)
	ctx := c.Request.Context()

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
		attribute.String("client.ip", ctxutil.GetClientIP(ctx)),
		attribute.String("user.agent", ctxutil.GetUserAgent(ctx)),
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
