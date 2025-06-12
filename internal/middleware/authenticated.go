package middleware

import (
	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/gin-gonic/gin"
)

// AuthenticatedSpace checks if user is related to space and authenticated
func AuthenticatedSpace(c *gin.Context) {
	// Get context
	ctx := c.Request.Context()
	// Retrieve space ID from context
	spaceID := ctxutil.GetSpaceID(ctx)

	if validator.IsEmpty(spaceID) {
		logger.Warn(ctx, "Space authentication failed")
		resp.Fail(c.Writer, resp.UnAuthorized(ecode.Text(ecode.Unauthorized)))
		c.Abort()
		return
	}

	c.Next()
}

// AuthenticatedUser checks if user is authenticated
func AuthenticatedUser(c *gin.Context) {
	// Get context
	ctx := c.Request.Context()
	// Retrieve user ID from context
	userID := ctxutil.GetUserID(ctx)

	if validator.IsEmpty(userID) {
		logger.Warn(ctx, "User authentication failed")
		resp.Fail(c.Writer, resp.UnAuthorized(ecode.Text(ecode.Unauthorized)))
		c.Abort()
		return
	}

	c.Next()
}
