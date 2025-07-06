package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/net/resp"
)

// ValidateContentType validates content type
func ValidateContentType() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut || c.Request.Method == http.MethodPatch {
			contentType := c.GetHeader("Content-Type")

			allowedTypes := []string{
				"application/json",
				"multipart/form-data",
				"application/x-www-form-urlencoded",
				"application/octet-stream",
			}

			for _, allowed := range allowedTypes {
				if strings.HasPrefix(contentType, allowed) {
					c.Next()
					return
				}
			}

			resp.Fail(c.Writer, resp.BadRequest("Unsupported Content-Type: "+contentType))
			c.Abort()
			return
		}

		c.Next()
	}
}
