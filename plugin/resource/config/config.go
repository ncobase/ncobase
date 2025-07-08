package config

import (
	"github.com/spf13/viper"
)

// Config holds resource module configuration
type Config struct {
	MaxUploadSize   int64        `json:"max_upload_size"`
	AllowedTypes    []string     `json:"allowed_types"`
	DefaultStorage  string       `json:"default_storage"`
	ImageProcessing *ImageConfig `json:"image_processing"`
	QuotaManagement *QuotaConfig `json:"quota_management"`
}

// ImageConfig holds image processing configuration
type ImageConfig struct {
	EnableThumbnails       bool `json:"enable_thumbnails"`
	DefaultThumbnailWidth  int  `json:"default_thumbnail_width"`
	DefaultThumbnailHeight int  `json:"default_thumbnail_height"`
	EnableResizing         bool `json:"enable_resizing"`
	MaxImageWidth          int  `json:"max_image_width"`
	MaxImageHeight         int  `json:"max_image_height"`
}

// QuotaConfig holds quota management configuration
type QuotaConfig struct {
	EnableQuotas       bool    `json:"enable_quotas"`
	DefaultQuota       int64   `json:"default_quota"`
	WarningThreshold   float64 `json:"warning_threshold"`
	QuotaCheckInterval string  `json:"quota_check_interval"`
}

// New returns a new Config instance with default values
func New() *Config {
	return &Config{
		MaxUploadSize:  5 * 1024 * 1024 * 1024, // 5GB default
		AllowedTypes:   []string{"*"},          // All types by default
		DefaultStorage: "filesystem",
		ImageProcessing: &ImageConfig{
			EnableThumbnails:       true,
			DefaultThumbnailWidth:  300,
			DefaultThumbnailHeight: 300,
			EnableResizing:         true,
			MaxImageWidth:          2048,
			MaxImageHeight:         2048,
		},
		QuotaManagement: &QuotaConfig{
			EnableQuotas:       true,
			DefaultQuota:       10 * 1024 * 1024 * 1024, // 10GB default
			WarningThreshold:   0.8,                     // 80% warning
			QuotaCheckInterval: "24h",                   // Daily check
		},
	}
}

// LoadFromViper loads configuration from Viper into the Config instance
func (c *Config) LoadFromViper(viper *viper.Viper) {
	if viper == nil {
		return
	}

	// MaxUploadSize
	if viper.IsSet("resource.max_upload_size") {
		c.MaxUploadSize = viper.GetInt64("resource.max_upload_size")
	}

	// AllowedTypes
	if viper.IsSet("resource.allowed_types") {
		c.AllowedTypes = viper.GetStringSlice("resource.allowed_types")
	}

	// DefaultStorage
	if viper.IsSet("resource.default_storage") {
		c.DefaultStorage = viper.GetString("resource.default_storage")
	}

	// Load image processing config
	if c.ImageProcessing == nil {
		c.ImageProcessing = &ImageConfig{}
	}

	if viper.IsSet("resource.image_processing.enable_thumbnails") {
		c.ImageProcessing.EnableThumbnails = viper.GetBool("resource.image_processing.enable_thumbnails")
	}

	if viper.IsSet("resource.image_processing.default_thumbnail_width") {
		c.ImageProcessing.DefaultThumbnailWidth = viper.GetInt("resource.image_processing.default_thumbnail_width")
	}

	if viper.IsSet("resource.image_processing.default_thumbnail_height") {
		c.ImageProcessing.DefaultThumbnailHeight = viper.GetInt("resource.image_processing.default_thumbnail_height")
	}

	if viper.IsSet("resource.image_processing.enable_resizing") {
		c.ImageProcessing.EnableResizing = viper.GetBool("resource.image_processing.enable_resizing")
	}

	if viper.IsSet("resource.image_processing.max_image_width") {
		c.ImageProcessing.MaxImageWidth = viper.GetInt("resource.image_processing.max_image_width")
	}

	if viper.IsSet("resource.image_processing.max_image_height") {
		c.ImageProcessing.MaxImageHeight = viper.GetInt("resource.image_processing.max_image_height")
	}

	// Load quota management config
	if c.QuotaManagement == nil {
		c.QuotaManagement = &QuotaConfig{}
	}

	if viper.IsSet("resource.quota_management.enable_quotas") {
		c.QuotaManagement.EnableQuotas = viper.GetBool("resource.quota_management.enable_quotas")
	}

	if viper.IsSet("resource.quota_management.default_quota") {
		c.QuotaManagement.DefaultQuota = viper.GetInt64("resource.quota_management.default_quota")
	}

	if viper.IsSet("resource.quota_management.warning_threshold") {
		c.QuotaManagement.WarningThreshold = viper.GetFloat64("resource.quota_management.warning_threshold")
	}

	if viper.IsSet("resource.quota_management.quota_check_interval") {
		c.QuotaManagement.QuotaCheckInterval = viper.GetString("resource.quota_management.quota_check_interval")
	}
}
