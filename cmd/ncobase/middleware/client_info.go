package middleware

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/ctxutil"
)

// ClientInfo middleware extracts and sets client information to context
func ClientInfo(c *gin.Context) {
	// Get context
	ctx := c.Request.Context()
	if _, ok := ctxutil.GetGinContext(ctx); !ok {
		ctx = ctxutil.WithGinContext(ctx, c)
	}

	// Extract and set client information
	ctx = ctxutil.SetClientInfo(ctx,
		extractRealClientIP(c),
		extractUserAgent(c),
		extractSessionID(c),
	)

	// Set HTTP request for direct access
	ctx = ctxutil.SetHTTPRequest(ctx, c.Request)

	// Update request context
	c.Request = c.Request.WithContext(ctx)

	c.Next()
}

// extractRealClientIP extracts real client IP with comprehensive proxy support
func extractRealClientIP(c *gin.Context) string {
	// Check forwarded headers in order of priority
	forwardedHeaders := []struct {
		name     string
		multiple bool // whether header can contain multiple IPs
	}{
		{"X-Forwarded-For", true},
		{"X-Real-IP", false},
		{"CF-Connecting-IP", false},
		{"X-Client-IP", false},
		{"X-Cluster-Client-IP", false},
		{"Forwarded-For", false},
		{"Forwarded", false},
	}

	for _, header := range forwardedHeaders {
		value := c.GetHeader(header.name)
		if value == "" || value == "unknown" {
			continue
		}

		if header.multiple {
			// Handle comma-separated IPs (X-Forwarded-For)
			ips := strings.Split(value, ",")
			for _, ip := range ips {
				cleanIP := strings.TrimSpace(ip)
				if isValidPublicIP(cleanIP) {
					return cleanIP
				}
			}
		} else {
			if isValidPublicIP(value) {
				return value
			}
		}
	}

	// Fallback to Gin's ClientIP method
	if clientIP := c.ClientIP(); clientIP != "" {
		return clientIP
	}

	// Last resort: extract from RemoteAddr
	if c.Request != nil && c.Request.RemoteAddr != "" {
		if host, _, err := net.SplitHostPort(c.Request.RemoteAddr); err == nil {
			return host
		}
		return c.Request.RemoteAddr
	}

	return "unknown"
}

// extractUserAgent extracts User-Agent header
func extractUserAgent(c *gin.Context) string {
	if ua := c.GetHeader("User-Agent"); ua != "" {
		return ua
	}
	return "unknown"
}

// extractSessionID extracts session ID from various sources
func extractSessionID(c *gin.Context) string {
	// Try session cookie
	sessionCookieNames := []string{"session_id", "sessionid", "SESSIONID"}
	for _, cookieName := range sessionCookieNames {
		if sessionID, err := c.Cookie(cookieName); err == nil && sessionID != "" {
			return sessionID
		}
	}

	// Try session headers
	sessionHeaders := []string{"X-Session-ID", "X-Session-Id", "Session-ID"}
	for _, headerName := range sessionHeaders {
		if sessionID := c.GetHeader(headerName); sessionID != "" {
			return sessionID
		}
	}

	return ""
}

// isValidPublicIP checks if IP is valid and not private/reserved
func isValidPublicIP(ipStr string) bool {
	if ipStr == "" || ipStr == "unknown" {
		return false
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// Check for private/reserved IP ranges
	privateRanges := []string{
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"127.0.0.0/8",    // Loopback
		"169.254.0.0/16", // Link Local
		"224.0.0.0/4",    // Multicast
		"240.0.0.0/4",    // Reserved
		"::1/128",        // IPv6 loopback
		"fc00::/7",       // IPv6 unique local
		"fe80::/10",      // IPv6 link local
		"ff00::/8",       // IPv6 multicast
	}

	for _, rangeStr := range privateRanges {
		_, subnet, err := net.ParseCIDR(rangeStr)
		if err != nil {
			continue
		}
		if subnet.Contains(ip) {
			return false
		}
	}

	return true
}
