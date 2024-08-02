package middleware

import (
	"path"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// shouldSkipPath checks if the path should be skipped
func shouldSkipPath(requestPath string, whiteList []string) bool {
	// Skip root path
	if requestPath == "/" {
		return true
	}

	for _, whitePath := range whiteList {
		// Support wildcard
		if strings.Contains(whitePath, "*") {
			matched, _ := path.Match(whitePath, requestPath)
			if matched {
				return true
			}
		} else {
			// Support regex
			matched, _ := regexp.MatchString(whitePath, requestPath)
			if matched {
				return true
			}
			// Support prefix
			if strings.HasPrefix(requestPath, whitePath) {
				return true
			}
		}
	}
	return false
}

// WhiteList is a middleware for white list
func WhiteList(whiteList []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipPath(c.Request.URL.Path, whiteList) {
			c.Next()
			return
		}
		// set skip header
		c.Writer.Header().Set("X-Skip-Auth", "true")
		c.Next()
	}
}
