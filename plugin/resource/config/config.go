package config

// Config holds resource runtime configuration loaded from system options.
type Config struct {
	MaxUploadSize   int64        `json:"max_upload_size"`
	AllowedTypes    []string     `json:"allowed_types"`
	DefaultStorage  string       `json:"default_storage"`
	ImageProcessing *ImageConfig `json:"image_processing"`
	QuotaManagement *QuotaConfig `json:"quota_management"`
}

// ImageConfig holds image processing configuration.
type ImageConfig struct {
	EnableThumbnails       bool `json:"enable_thumbnails"`
	DefaultThumbnailWidth  int  `json:"default_thumbnail_width"`
	DefaultThumbnailHeight int  `json:"default_thumbnail_height"`
	EnableResizing         bool `json:"enable_resizing"`
	MaxImageWidth          int  `json:"max_image_width"`
	MaxImageHeight         int  `json:"max_image_height"`
	CompressionQuality     int  `json:"compression_quality"`
}

// QuotaConfig holds quota management configuration.
type QuotaConfig struct {
	EnableQuotas       bool    `json:"enable_quotas"`
	EnableEnforcement  bool    `json:"enable_enforcement"`
	DefaultQuota       int64   `json:"default_quota"`
	WarningThreshold   float64 `json:"warning_threshold"`
	QuotaCheckInterval string  `json:"quota_check_interval"`
}

// New returns default resource runtime configuration.
func New() *Config {
	return &Config{
		MaxUploadSize:  5 * 1024 * 1024 * 1024,
		AllowedTypes:   []string{"*"},
		DefaultStorage: "configured",
		ImageProcessing: &ImageConfig{
			EnableThumbnails:       true,
			DefaultThumbnailWidth:  300,
			DefaultThumbnailHeight: 300,
			EnableResizing:         true,
			MaxImageWidth:          2048,
			MaxImageHeight:         2048,
			CompressionQuality:     85,
		},
		QuotaManagement: &QuotaConfig{
			EnableQuotas:       true,
			EnableEnforcement:  true,
			DefaultQuota:       10 * 1024 * 1024 * 1024,
			WarningThreshold:   0.8,
			QuotaCheckInterval: "24h",
		},
	}
}
