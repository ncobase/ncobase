package helper

import (
	"context"

	"github.com/ncobase/common/config"

	"github.com/gin-gonic/gin"
)

// SetConfig sets config to gin.Context or context.Context.
func SetConfig(c *gin.Context, conf *config.Config) context.Context {
	return SetValue(c, "config", conf)
}

// GetConfig gets config from gin.Context or context.Context.
func GetConfig(c *gin.Context) *config.Config {
	if conf, ok := GetValue(c, "config").(*config.Config); ok {
		return conf
	}
	// Context does not contain config, load it from config.
	return config.GetConfig()
}
