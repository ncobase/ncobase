package middleware

import (
	"net/http"
	"stocms/pkg/ecode"
	"stocms/pkg/jwt"
	"stocms/pkg/resp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// refreshToken TODO 刷新 Token
func refreshToken(oldToken string) (string, error) {
	return oldToken, nil
}

// isTokenExpiring 检查令牌是否即将过期
func isTokenExpiring(tokenData map[string]interface{}) bool {
	exp, ok := tokenData["exp"].(int64)
	if !ok {
		return false
	}
	expirationTime := time.Unix(exp, 0)
	return time.Until(expirationTime) < 10*time.Minute // 假设如果令牌在 10 分钟内过期，则刷新
}

// ConsumeUser 处理当前用户
func ConsumeUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.Request.Header.Get("Authorization")
		// Check format
		// ie Bearer: eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9
		b := "Bearer: "
		if !strings.Contains(token, b) {
			ctx.Next()
			return
		}
		t := strings.Split(token, b)
		if len(t) < 2 {
			ctx.Next()
			return
		}
		// 解密令牌
		tokenData, err := jwt.DecodeToken(signingKey, t[1])
		if err != nil {
			exception := &resp.Exception{
				Status:  http.StatusForbidden,
				Code:    ecode.AccessDenied,
				Message: err.Error(),
			}
			resp.Fail(ctx, exception)
			return
		}

		// 设置当前用户 ID
		userID := tokenData["user_id"].(string)
		ctx.Set("uid", userID)

		// 检查令牌是否即将过期并刷新令牌
		if isTokenExpiring(tokenData) {
			newToken, err := refreshToken(t[1])
			if err == nil {
				ctx.Header("Authorization", "Bearer "+newToken)
			}
		}

		ctx.Next()

	}
}
