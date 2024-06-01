package middleware

import (
	"net/http"
	"stocms/internal/helper"
	"stocms/pkg/ecode"
	"stocms/pkg/resp"
	"stocms/pkg/validator"

	"github.com/gin-gonic/gin"
)

// Authorized is a middleware for verifying the existence of a user.
func Authorized(c *gin.Context) {
	if userID := helper.GetUserID(c); validator.IsNil(userID) {
		exception := &resp.Exception{
			Status:  http.StatusUnauthorized,
			Code:    ecode.Unauthorized,
			Message: ecode.Text(ecode.Unauthorized),
		}
		resp.Fail(c.Writer, exception)
		c.Abort() // Abort all handlers
		return
	}
	c.Next()
}
