package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig defines the config for CORS middleware.
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	ExposeHeaders    []string
	MaxAge           int
}

// defaultCORSConfig is the default config for CORS middleware.
var defaultCORSConfig = CORSConfig{
	AllowOrigins:     []string{"*"},
	AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
	AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
	AllowCredentials: false,
	ExposeHeaders:    []string{},
	MaxAge:           12 * 60 * 60, // 12 hours
}

// CORSHandler is a middleware for handling CORS.
func CORSHandler(c *gin.Context) {
	config := defaultCORSConfig
	c.Writer.Header().Set("Access-Control-Allow-Origin", strings.Join(config.AllowOrigins, ","))
	c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ","))
	c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ","))
	c.Writer.Header().Set("Access-Control-Allow-Credentials", strconv.FormatBool(config.AllowCredentials))
	c.Writer.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ","))
	c.Writer.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))

	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.Next()
}
