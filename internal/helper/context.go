package helper

import (
	"context"
	"stocms/internal/config"

	"github.com/gin-gonic/gin"
)

// FromGinContext extracts the context.Context from *gin.Context
func FromGinContext(c *gin.Context) context.Context {
	return c.Request.Context()
}

// GetConfig gets config from gin.Context
func GetConfig(c *gin.Context) *config.Config {
	if conf, exists := c.Get("config"); exists {
		return conf.(*config.Config)
	}
	// context does not contain config, load it from config
	return config.GetConfig()
}

// SetConfig sets config to gin.Context
func SetConfig(c *gin.Context, conf *config.Config) {
	c.Set("config", conf)
}

// GetUserID gets user id from gin.Context
func GetUserID(c *gin.Context) string {
	if uid, exists := c.Get("user_id"); exists {
		return uid.(string)
	}
	return ""
}

// SetUserID sets user id to gin.Context
func SetUserID(c *gin.Context, uid string) {
	c.Set("user_id", uid)
}

// GetUsername gets username from gin.Context
func GetUsername(c *gin.Context) string {
	if username, exists := c.Get("username"); exists {
		return username.(string)
	}
	return ""
}

// SetUsername sets username to gin.Context
func SetUsername(c *gin.Context, username string) {
	c.Set("username", username)
}

// GetToken gets token from gin.Context
func GetToken(c *gin.Context) string {
	if token, exists := c.Get("token"); exists {
		return token.(string)
	}
	return ""
}

// SetToken sets token to gin.Context
func SetToken(c *gin.Context, token string) {
	c.Set("token", token)
}

// GetRole gets role from gin.Context
func GetRole(c *gin.Context) string {
	if role, exists := c.Get("role"); exists {
		return role.(string)
	}
	return ""
}

// SetRole sets role to gin.Context
func SetRole(c *gin.Context, role string) {
	c.Set("role", role)
}

// GetProvider gets provider from gin.Context
func GetProvider(c *gin.Context) string {
	if provider, exists := c.Get("provider"); exists {
		return provider.(string)
	}
	return ""
}

// SetProvider sets provider to gin.Context
func SetProvider(c *gin.Context, provider string) {
	c.Set("provider", provider)
}

// GetProfile gets profile from gin.Context
func GetProfile(c *gin.Context) any {
	if profile, exists := c.Get("profile"); exists {
		return profile
	}
	return nil
}

// SetProfile sets profile to gin.Context
func SetProfile(c *gin.Context, profile any) {
	c.Set("profile", profile)
}

// GetRequestID gets request ID from gin.Context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}
	return ""
}

// SetRequestID sets request ID to gin.Context
func SetRequestID(c *gin.Context, requestID string) {
	c.Set("request_id", requestID)
}

// GetTraceID gets trace ID from gin.Context
func GetTraceID(c *gin.Context) string {
	if traceID, exists := c.Get("trace_id"); exists {
		return traceID.(string)
	}
	return ""
}

// SetTraceID sets trace ID to gin.Context
func SetTraceID(c *gin.Context, traceID string) {
	c.Set("trace_id", traceID)
}
