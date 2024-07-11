package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSHeaders contains the CORS headers.
var CORSHeaders = map[string]string{
	"Access-Control-Allow-Origin":      "*",
	"Access-Control-Allow-Credentials": "true",
	"Access-Control-Allow-Headers":     "content-type, content-length, accept-encoding, x-csrf-token, authorization, accept, origin, cache-control, x-requested-with",
	"Access-Control-Allow-Methods":     "POST, GET, OPTIONS, PUT, DELETE",
}

// CORSHandler is a middleware for handling CORS.
func CORSHandler(c *gin.Context) {
	// Set CORS headers
	for key, value := range CORSHeaders {
		c.Writer.Header().Set(key, value)
	}

	// Handle preflight requests
	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.Next()
}
