package middleware

import (
	"net/http"
	"stocms/internal/helper"
	"stocms/pkg/ecode"
	"stocms/pkg/jwt"
	"stocms/pkg/resp"
	"stocms/pkg/types"
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

// ConsumeUser consume user
func ConsumeUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		// Check format
		// ie Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9
		b := "Bearer "
		if !strings.Contains(token, b) {
			c.Next()
			return
		}
		t := strings.Split(token, b)
		if len(t) < 2 {
			c.Next()
			return
		}
		// decode token
		tokenData, err := jwt.DecodeToken(signingKey, t[1])
		if err != nil {
			exception := &resp.Exception{
				Status:  http.StatusForbidden,
				Code:    ecode.AccessDenied,
				Message: err.Error(),
			}
			resp.Fail(c.Writer, exception)
			return
		}

		payload := tokenData["payload"].(types.JSON)

		// set uid
		userID := payload["user_id"].(string)
		helper.SetUserID(c, userID)

		// if token is expiring, refresh token
		if isTokenExpiring(tokenData) {
			newToken, err := refreshToken(t[1])
			if err == nil {
				c.Header("Authorization", "Bearer "+newToken)
			}
		}

		c.Next()

	}
}
