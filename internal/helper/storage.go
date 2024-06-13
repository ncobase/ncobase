package helper

import (
	"stocms/pkg/log"
	"stocms/pkg/storage"

	"github.com/casdoor/oss"
	"github.com/gin-gonic/gin"
)

// SetStorage sets storage to gin.Context
func SetStorage(c *gin.Context, s oss.StorageInterface) {
	SetValue(c, "storage", s)
}

// GetStorage gets storage from gin.Context
func GetStorage(c *gin.Context) (oss.StorageInterface, *storage.Config) {
	if s, ok := GetValue(c, "storage").(oss.StorageInterface); ok {
		return s, &GetConfig(c).Storage
	}

	// Get config
	storageConfig := GetConfig(c).Storage

	// Initialize storage
	s, err := storage.NewStorage(&storageConfig)
	if err != nil {
		log.Errorf(c, "Error creating storage: %v\n", err)
		return nil, nil
	}
	// Set storage to gin.Context
	// SetStorage(c, s)
	return s, &storageConfig
}

// // GetCaptchaImage get captcha image
// func GetCaptchaImage(c *gin.Context) (string, error) {
// 	return
// }
