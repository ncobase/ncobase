package resource

import "github.com/spf13/viper"

// Config holds configuration for the resource module
type Config struct {
	MaxUploadSize   int64        `json:"max_upload_size"`
	AllowedTypes    []string     `json:"allowed_types"`
	DefaultStorage  string       `json:"default_storage"`
	ImageProcessing *ImageConfig `json:"image_processing"`
	QuotaManagement *QuotaConfig `json:"quota_management"`
}

// ImageConfig holds configuration for image processing
type ImageConfig struct {
	EnableThumbnails       bool `json:"enable_thumbnails"`
	DefaultThumbnailWidth  int  `json:"default_thumbnail_width"`
	DefaultThumbnailHeight int  `json:"default_thumbnail_height"`
	EnableResizing         bool `json:"enable_resizing"`
	MaxImageWidth          int  `json:"max_image_width"`
	MaxImageHeight         int  `json:"max_image_height"`
}

// QuotaConfig holds configuration for storage quota management
type QuotaConfig struct {
	EnableQuotas       bool    `json:"enable_quotas"`
	DefaultQuota       int64   `json:"default_quota"`
	WarningThreshold   float64 `json:"warning_threshold"`
	QuotaCheckInterval string  `json:"quota_check_interval"`
}

// GetDefaultConfig returns the default configuration for the resource module
func (m *Module) GetDefaultConfig() *Config {
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

// GetConfigFromFile loads configuration from Viper into the resource configuration
func (m *Module) GetConfigFromFile(viper *viper.Viper) {
	if viper == nil {
		return
	}

	// MaxUploadSize
	if viper.IsSet("resource.max_upload_size") {
		m.config.MaxUploadSize = viper.GetInt64("resource.max_upload_size")
	}

	// AllowedTypes
	if viper.IsSet("resource.allowed_types") {
		m.config.AllowedTypes = viper.GetStringSlice("resource.allowed_types")
	}

	// DefaultStorage
	if viper.IsSet("resource.default_storage") {
		m.config.DefaultStorage = viper.GetString("resource.default_storage")
	}

	// Load image processing config
	if m.config.ImageProcessing == nil {
		m.config.ImageProcessing = &ImageConfig{}
	}

	if viper.IsSet("resource.image_processing.enable_thumbnails") {
		m.config.ImageProcessing.EnableThumbnails = viper.GetBool("resource.image_processing.enable_thumbnails")
	}

	if viper.IsSet("resource.image_processing.default_thumbnail_width") {
		m.config.ImageProcessing.DefaultThumbnailWidth = viper.GetInt("resource.image_processing.default_thumbnail_width")
	}

	if viper.IsSet("resource.image_processing.default_thumbnail_height") {
		m.config.ImageProcessing.DefaultThumbnailHeight = viper.GetInt("resource.image_processing.default_thumbnail_height")
	}

	if viper.IsSet("resource.image_processing.enable_resizing") {
		m.config.ImageProcessing.EnableResizing = viper.GetBool("resource.image_processing.enable_resizing")
	}

	if viper.IsSet("resource.image_processing.max_image_width") {
		m.config.ImageProcessing.MaxImageWidth = viper.GetInt("resource.image_processing.max_image_width")
	}

	if viper.IsSet("resource.image_processing.max_image_height") {
		m.config.ImageProcessing.MaxImageHeight = viper.GetInt("resource.image_processing.max_image_height")
	}

	// Load quota management config
	if m.config.QuotaManagement == nil {
		m.config.QuotaManagement = &QuotaConfig{}
	}

	if viper.IsSet("resource.quota_management.enable_quotas") {
		m.config.QuotaManagement.EnableQuotas = viper.GetBool("resource.quota_management.enable_quotas")
	}

	if viper.IsSet("resource.quota_management.default_quota") {
		m.config.QuotaManagement.DefaultQuota = viper.GetInt64("resource.quota_management.default_quota")
	}

	if viper.IsSet("resource.quota_management.warning_threshold") {
		m.config.QuotaManagement.WarningThreshold = viper.GetFloat64("resource.quota_management.warning_threshold")
	}

	if viper.IsSet("resource.quota_management.quota_check_interval") {
		m.config.QuotaManagement.QuotaCheckInterval = viper.GetString("resource.quota_management.quota_check_interval")
	}
}
