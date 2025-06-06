package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// shouldSkipPath checks if the path should be skipped
func shouldSkipPath(request *http.Request, whiteList []string) bool {
	requestMethod := request.Method
	requestPath := request.URL.Path

	// Skip root path "/"
	if requestPath == "/" {
		return true
	}

	// Combine method and path, e.g., "GET:/path"
	fullRequest := requestMethod + ":" + requestPath

	for _, whitePath := range whiteList {
		// Support wildcard (e.g., "*keyword*", "*keyword/*", "*/keyword/*")
		if strings.Contains(whitePath, "*") {
			// Convert wildcard to regex (e.g., "*keyword*" -> ".*keyword.*")
			regexPattern := "^" + regexp.QuoteMeta(whitePath)
			regexPattern = strings.ReplaceAll(regexPattern, `\*`, ".*") + "$"

			// Precompile the regex for performance
			compiledRegex, err := regexp.Compile(regexPattern)
			if err != nil {
				continue // Skip invalid regex patterns
			}

			// Match against both full request (method + path) and just the path
			if compiledRegex.MatchString(fullRequest) || compiledRegex.MatchString(requestPath) {
				return true
			}
		} else {
			// Check for exact method:path match
			if fullRequest == whitePath {
				return true
			}
			// Check for exact path match (for paths without methods in whitelist)
			if requestPath == whitePath {
				return true
			}
			// Support prefix match for paths without methods (e.g., /static/*)
			if strings.HasPrefix(whitePath, "/") && strings.HasPrefix(requestPath, whitePath) {
				return true
			}
		}
	}
	return false
}

// WhiteList is a middleware for white list
func WhiteList(whiteList []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipPath(c.Request, whiteList) {
			c.Next()
			return
		}
		// set skip header
		c.Writer.Header().Set("X-Skip-Auth", "true")
		c.Next()
	}
}
