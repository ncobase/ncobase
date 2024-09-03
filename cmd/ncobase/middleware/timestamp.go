package middleware

import (
	"ncobase/common/resp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Timestamp middleware for checking timestamp
func Timestamp(whiteList []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipPath(c.Request.URL.Path, whiteList) {
			c.Next()
			return
		}

		_t := c.Query("_t")
		if _t == "" {
			resp.Fail(c.Writer, resp.BadRequest("Missing timestamp parameter '_t'"))
			c.Abort()
			return
		}

		timestamp, err := strconv.ParseInt(_t, 10, 64)
		if err != nil {
			resp.Fail(c.Writer, resp.BadRequest("Invalid timestamp format"))
			c.Abort()
			return
		}

		now := time.Now().UnixMilli()

		diff := abs(now - timestamp)

		// 15 minutes
		if diff > 900000 {
			resp.Fail(c.Writer, resp.BadRequest("Timestamp is more than 15 minutes off"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// abs returns the absolute value of x.
func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
