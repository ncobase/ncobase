package helper

import (
	"context"
	"stocms/internal/config"
	"stocms/pkg/nanoid"

	"github.com/gin-gonic/gin"
)

// FromGinContext extracts the context.Context from *gin.Context.
func FromGinContext(c *gin.Context) context.Context {
	return c.Request.Context()
}

// WithGinContext returns a context.Context that embeds the *gin.Context.
func WithGinContext(ctx context.Context, c *gin.Context) context.Context {
	return context.WithValue(ctx, "ginContext", c)
}

// GetGinContext extracts *gin.Context from context.Context if it exists.
func GetGinContext(ctx context.Context) (*gin.Context, bool) {
	if c, ok := ctx.Value("ginContext").(*gin.Context); ok {
		return c, ok
	}
	return nil, false
}

// GetValue retrieves a value from the context.
func GetValue(c *gin.Context, key string) any {
	// if c, ok := GetGinContext(ctx); ok {
	if val, exists := c.Get(key); exists {
		return val
	}
	// }
	return nil
}

// SetValue sets a value to the context.
func SetValue(c *gin.Context, key string, val any) context.Context {
	// if c, ok := GetGinContext(ctx); ok {
	c.Set(key, val)
	// }
	return context.WithValue(c, key, val)
}

// GetConfig gets config from gin.Context or context.Context.
func GetConfig(c *gin.Context) *config.Config {
	if conf, ok := GetValue(c, "config").(*config.Config); ok {
		return conf
	}
	// Context does not contain config, load it from config.
	return config.GetConfig()
}

// SetConfig sets config to gin.Context or context.Context.
func SetConfig(c *gin.Context, conf *config.Config) context.Context {
	return SetValue(c, "config", conf)
}

// GetUserID gets user id from gin.Context
func GetUserID(c *gin.Context) string {
	if uid, ok := GetValue(c, "user_id").(string); ok {
		return uid
	}
	return ""
}

// SetUserID sets user id to gin.Context
func SetUserID(c *gin.Context, uid string) {
	SetValue(c, "user_id", uid)
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

// GetProvider gets provider from gin.Context
func GetProvider(c *gin.Context) string {
	if provider, ok := GetValue(c, "provider").(string); ok {
		return provider
	}
	return ""
}

// SetProvider sets provider to gin.Context
func SetProvider(c *gin.Context, provider string) {
	SetValue(c, "provider", provider)
}

// GetProfile gets profile from gin.Context
func GetProfile(c *gin.Context) any {
	if profile, ok := GetValue(c, "profile").(any); ok {
		return profile
	}
	return nil
}

// SetProfile sets profile to gin.Context
func SetProfile(c *gin.Context, profile any) {
	SetValue(c, "profile", profile)
}

// GetRequestID gets request id from gin.Context
func GetRequestID(c *gin.Context) string {
	if rid, ok := GetValue(c, "request_id").(string); ok {
		return rid
	}
	return ""
}

// SetRequestID sets request id to gin.Context
func SetRequestID(c *gin.Context, rid string) {
	SetValue(c, "request_id", rid)
}

// GetTraceID gets trace id from gin.Context
func GetTraceID(c *gin.Context) string {
	if traceID, ok := GetValue(c, "trace_id").(string); ok {
		return traceID
	}
	return ""
}

// SetTraceID sets trace id to gin.Context
func SetTraceID(c *gin.Context, traceID string) {
	SetValue(c, "trace_id", traceID)
}

// NewTraceID creates a new trace ID.
func NewTraceID() string {
	return nanoid.Must(16)
}

// NewRequestID creates a new request ID.
func NewRequestID() string {
	return nanoid.Must(16)
}
