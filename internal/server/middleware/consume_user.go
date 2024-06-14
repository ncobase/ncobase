package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"stocms/internal/helper"
	"stocms/pkg/consts"
	"stocms/pkg/ecode"
	"stocms/pkg/jwt"
	"stocms/pkg/resp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// refreshToken TODO refresh token
func refreshToken(oldToken string) (string, error) {
	return oldToken, nil
}

// isTokenExpiring token is expiring
func isTokenExpiring(tokenData map[string]any) bool {
	exp, ok := tokenData["exp"].(int64)
	if !ok {
		return false
	}
	expirationTime := time.Unix(exp, 0)
	return time.Until(expirationTime) < 10*time.Minute // 假设如果令牌在 10 分钟内过期，则刷新
}

// ConsumeUser consumes user authentication information.
func ConsumeUser(c *gin.Context) {
	// Retrieve token from request header
	token := c.GetHeader(consts.AuthorizationKey)

	// Check if token is in the correct format (Bearer token)
	if !strings.HasPrefix(token, consts.BearerKey) {
		c.Next()
		return
	}

	// Extract token value after "Bearer "
	token = strings.TrimPrefix(token, consts.BearerKey)

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
	helper.SetUserID(c, userID)

	// Check if token is expiring soon, and refresh if necessary
	if isTokenExpiring(tokenData) {
		newToken, err := refreshToken(token)
		if err != nil {
			handleTokenError(c, fmt.Errorf("failed to refresh token: %v", err))
			return
		}
		c.Header("Authorization", "Bearer "+newToken)
	}

	// Continue to next middleware or handler
	c.Next()
}

// handleTokenError handles token decoding/validation errors.
func handleTokenError(c *gin.Context, err error) {
	exception := &resp.Exception{
		Status:  http.StatusForbidden,
		Code:    ecode.AccessDenied,
		Message: err.Error(),
	}
	resp.Fail(c.Writer, exception)
}
