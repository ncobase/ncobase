package middleware

import (
	"context"
	"errors"
	"fmt"
	"ncobase/common/consts"
	"ncobase/common/cookie"
	"ncobase/common/ecode"
	"ncobase/common/jwt"
	"ncobase/common/resp"
	"ncobase/helper"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// refreshToken TODO refresh token
func refreshToken(oldToken string) (string, error) {
	return oldToken, nil
}

// isTokenExpiring checks if the token is expiring soon.
func isTokenExpiring(tokenData map[string]any) bool {
	exp, ok := tokenData["exp"].(int64)
	if !ok {
		return false
	}
	expirationTime := time.Unix(exp, 0)
	return time.Until(expirationTime) < 10*time.Minute // Assumes token should be refreshed if expiring within 10 minutes
}

// ConsumeUser consumes user authentication information.
func ConsumeUser(signingKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// Retrieve token from request header, query, or cookie
		if authHeader := c.GetHeader("Authorization"); authHeader != "" {
			// Token from header
			if strings.HasPrefix(authHeader, consts.BearerKey) {
				token = strings.TrimPrefix(authHeader, consts.BearerKey)
			}
		} else if queryToken := c.Query("ak"); queryToken != "" {
			// Token from query
			token = queryToken
		} else if cookieToken, err := c.Cookie("access_token"); err == nil {
			// Token from cookie
			token = cookieToken
		}

		// If token is still empty, proceed to the next middleware
		if token == "" {
			c.Next()
			return
		}

		// Decode token
		tokenData, err := jwt.DecodeToken(signingKey, token)
		if err != nil {
			handleTokenError(c, fmt.Errorf("failed to decode token: %v", err))
			return
		}

		// Extract payload from token data
		payload, ok := tokenData["payload"].(map[string]any)
		if !ok {
			handleTokenError(c, errors.New("invalid token payload format"))
			return
		}

		// Set user ID to context
		userID, ok := payload["user_id"].(string)
		if !ok || userID == "" {
			handleTokenError(c, errors.New("user_id not found in token payload"))
			return
		}
		ctx := helper.WithGinContext(context.Background(), c)
		helper.SetUserID(ctx, userID)
		c.Request = c.Request.WithContext(ctx)

		// Check if token is expiring soon, and refresh if necessary
		if isTokenExpiring(tokenData) {
			newToken, err := refreshToken(token)
			if err != nil {
				handleTokenError(c, fmt.Errorf("failed to refresh token: %v", err))
				return
			}
			c.Header("Authorization", consts.BearerKey+newToken)
			cookie.SetAccessToken(c.Writer, newToken, "")
		}

		// Continue to next middleware or handler
		c.Next()
	}
}

// handleTokenError handles token decoding/validation errors.
func handleTokenError(c *gin.Context, err error) {
	exception := &resp.Exception{
		Status:  http.StatusForbidden,
		Code:    ecode.AccessDenied,
		Message: err.Error(),
	}
	resp.Fail(c.Writer, exception)
	c.Abort()
}
