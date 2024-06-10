package helper

import "github.com/gin-gonic/gin"

// SetUserID sets user id to gin.Context
func SetUserID(c *gin.Context, uid string) {
	SetValue(c, "user_id", uid)
}

// GetUserID gets user id from gin.Context
func GetUserID(c *gin.Context) string {
	if uid, ok := GetValue(c, "user_id").(string); ok {
		return uid
	}
	return ""
}

// GetToken gets token from gin.Context
func GetToken(c *gin.Context) string {
	if token, ok := GetValue(c, "token").(string); ok {
		return token
	}
	return ""
}

// SetToken sets token to gin.Context
func SetToken(c *gin.Context, token string) {
	SetValue(c, "token", token)
}

// SetProvider sets provider to gin.Context
func SetProvider(c *gin.Context, provider string) {
	SetValue(c, "provider", provider)
}

// GetProvider gets provider from gin.Context
func GetProvider(c *gin.Context) string {
	if provider, ok := GetValue(c, "provider").(string); ok {
		return provider
	}
	return ""
}

// SetProfile sets profile to gin.Context
func SetProfile(c *gin.Context, profile any) {
	SetValue(c, "profile", profile)
}

// GetProfile gets profile from gin.Context
func GetProfile(c *gin.Context) any {
	if profile, ok := GetValue(c, "profile").(any); ok {
		return profile
	}
	return nil
}
