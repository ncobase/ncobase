package helper

import (
	"ncobase/internal/data/structs"

	"github.com/gin-gonic/gin"
)

// ShouldBindAndValidateStruct binds and validates struct
func ShouldBindAndValidateStruct(c *gin.Context, obj any, lang ...string) (map[string]string, error) {
	contentType := c.GetHeader("Content-Type")
	if contentType == "" {
		contentType = "application/json;charset=utf-8"
	}

	if err := c.ShouldBind(obj); err != nil {
		return nil, err
	}

	return structs.Validate(obj, lang...), nil
}
