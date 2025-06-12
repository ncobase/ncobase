package middleware

import (
	"github.com/ncobase/ncore/consts"
	"github.com/ncobase/ncore/ctxutil"
	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"

	"github.com/gin-gonic/gin"
)

// ConsumeSpace consumes space information from request header or user spaces
func ConsumeSpace(em ext.ManagerInterface, whiteList []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipPath(c.Request, whiteList) {
			c.Next()
			return
		}

		ctx := c.Request.Context()
		userID := ctxutil.GetUserID(ctx)
		spaceID := c.GetHeader(consts.SpaceKey)

		// Get service manager
		sw := GetServiceManager(em)
		// Get space wrapper
		tsw := sw.SpaceServiceWrapper()

		// Validate space ID belongs to user if both provided
		if spaceID != "" && userID != "" {
			if isValid, err := tsw.IsSpaceInUser(ctx, spaceID, userID); err != nil || !isValid {
				logger.Warnf(ctx, "Space %s does not belong to user %s", spaceID, userID)
				spaceID = ""
			}
		}

		// Get space from context or user spaces if not provided/invalid
		if spaceID == "" && userID != "" {
			spaceID = ctxutil.GetSpaceID(ctx)

			if spaceID == "" {
				logger.Info(ctx, "space not found in header or context, trying to fetch from user spaces")

				// Try to get default space first
				if space, err := tsw.GetUserDefaultSpace(ctx, userID); err == nil && space != nil {
					spaceID = space.ID
				} else {
					// Get any space user belongs to
					if spaces, err := tsw.GetUserSpaces(ctx, userID); err == nil && len(spaces) > 0 {
						spaceID = spaces[0].ID
					}
				}
			}
		}

		// Set space ID to context if exists
		if spaceID != "" {
			ctx = ctxutil.SetSpaceID(ctx, spaceID)
			c.Request = c.Request.WithContext(ctx)
		} else if userID != "" {
			logger.Warnf(ctx, "No space found for user: %s", userID)
		}

		c.Next()
	}
}
