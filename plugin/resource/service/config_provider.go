package service

import (
	"context"
	"fmt"
	rConfig "ncobase/plugin/resource/config"
	"ncobase/plugin/resource/structs"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	optionResourceUpload      = "resource.upload"
	optionResourceImage       = "resource.image"
	optionResourceQuota       = "resource.quota"
	optionSystemStoragePolicy = "system.storage_policy"
)

// ResourceConfigProvider resolves resource runtime policy from system options.
type ResourceConfigProvider interface {
	Get(ctx context.Context) *rConfig.Config
	ValidateUpload(ctx context.Context, filename, contentType string, size int64) error
	DefaultProcessingOptions(ctx context.Context) *structs.ProcessingOptions
	NormalizeProcessingOptions(ctx context.Context, options *structs.ProcessingOptions) *structs.ProcessingOptions
	QuotaConfig(ctx context.Context) *QuotaConfig
	StoragePolicy(ctx context.Context) *StoragePolicyConfig
}

// OptionObjectService defines the option object lookup required by resource config.
type OptionObjectService interface {
	GetObjectByName(ctx context.Context, name string) (map[string]any, error)
}

type systemOptionConfigProvider struct {
	option OptionObjectService
}

// StoragePolicyConfig holds non-secret resource sharing policy.
type StoragePolicyConfig struct {
	DefaultProvider   string `json:"default_provider"`
	AllowPublicLinks  bool   `json:"allow_public_links"`
	RequireOwnerScope bool   `json:"require_owner_scope"`
	AuditDownloads    bool   `json:"audit_downloads"`
}

// NewSystemOptionConfigProvider creates a runtime config provider backed by /sys/options.
func NewSystemOptionConfigProvider(option OptionObjectService) ResourceConfigProvider {
	return &systemOptionConfigProvider{option: option}
}

// NewDefaultConfigProvider creates a provider backed by code defaults.
func NewDefaultConfigProvider() ResourceConfigProvider {
	return &systemOptionConfigProvider{}
}

func (p *systemOptionConfigProvider) Get(ctx context.Context) *rConfig.Config {
	cfg := rConfig.New()
	if p == nil || p.option == nil {
		return cfg
	}

	if upload, err := p.option.GetObjectByName(ctx, optionResourceUpload); err == nil {
		applyUploadConfig(cfg, upload)
	}
	if image, err := p.option.GetObjectByName(ctx, optionResourceImage); err == nil {
		applyImageConfig(cfg, image)
	}
	if quota, err := p.option.GetObjectByName(ctx, optionResourceQuota); err == nil {
		applyQuotaConfig(cfg, quota)
	}

	return normalizeResourceConfig(cfg)
}

func (p *systemOptionConfigProvider) ValidateUpload(ctx context.Context, filename, contentType string, size int64) error {
	cfg := p.Get(ctx)
	if cfg.MaxUploadSize > 0 && size > cfg.MaxUploadSize {
		return fmt.Errorf("file size exceeds maximum upload size %d bytes", cfg.MaxUploadSize)
	}
	if !isAllowedResourceType(cfg.AllowedTypes, filename, contentType) {
		return fmt.Errorf("file type %s is not allowed", resourceTypeLabel(filename, contentType))
	}
	return nil
}

func (p *systemOptionConfigProvider) DefaultProcessingOptions(ctx context.Context) *structs.ProcessingOptions {
	cfg := p.Get(ctx)
	img := cfg.ImageProcessing
	if img == nil {
		return nil
	}
	return &structs.ProcessingOptions{
		CreateThumbnail:    img.EnableThumbnails,
		ResizeImage:        false,
		MaxWidth:           img.DefaultThumbnailWidth,
		MaxHeight:          img.DefaultThumbnailHeight,
		CompressImage:      false,
		CompressionQuality: img.CompressionQuality,
	}
}

func (p *systemOptionConfigProvider) NormalizeProcessingOptions(ctx context.Context, options *structs.ProcessingOptions) *structs.ProcessingOptions {
	cfg := p.Get(ctx)
	img := cfg.ImageProcessing
	if img == nil {
		return options
	}

	if options == nil {
		return p.DefaultProcessingOptions(ctx)
	}

	if !img.EnableThumbnails {
		options.CreateThumbnail = false
	}
	if !img.EnableResizing {
		options.ResizeImage = false
	}
	if options.CreateThumbnail {
		if options.MaxWidth <= 0 {
			options.MaxWidth = img.DefaultThumbnailWidth
		}
		if options.MaxHeight <= 0 {
			options.MaxHeight = img.DefaultThumbnailHeight
		}
	}
	if options.ResizeImage {
		if options.MaxWidth <= 0 || options.MaxWidth > img.MaxImageWidth {
			options.MaxWidth = img.MaxImageWidth
		}
		if options.MaxHeight <= 0 || options.MaxHeight > img.MaxImageHeight {
			options.MaxHeight = img.MaxImageHeight
		}
	}
	if options.CompressionQuality <= 0 || options.CompressionQuality > 100 {
		options.CompressionQuality = img.CompressionQuality
	}

	return options
}

func (p *systemOptionConfigProvider) QuotaConfig(ctx context.Context) *QuotaConfig {
	cfg := p.Get(ctx)
	quota := cfg.QuotaManagement
	if quota == nil {
		quota = rConfig.New().QuotaManagement
	}

	interval, err := parseDuration(quota.QuotaCheckInterval)
	if err != nil || interval <= 0 {
		interval = 24 * time.Hour
	}

	return &QuotaConfig{
		DefaultQuota:      quota.DefaultQuota,
		WarningThreshold:  quota.WarningThreshold,
		CheckInterval:     interval,
		EnableEnforcement: quota.EnableEnforcement && quota.EnableQuotas,
		EnableQuotas:      quota.EnableQuotas,
	}
}

func (p *systemOptionConfigProvider) StoragePolicy(ctx context.Context) *StoragePolicyConfig {
	policy := &StoragePolicyConfig{
		DefaultProvider:   "configured",
		AllowPublicLinks:  true,
		RequireOwnerScope: true,
		AuditDownloads:    true,
	}
	if p == nil || p.option == nil {
		return policy
	}

	values, err := p.option.GetObjectByName(ctx, optionSystemStoragePolicy)
	if err != nil {
		return policy
	}
	if v, ok := stringFromAny(values["default_provider"]); ok && strings.TrimSpace(v) != "" {
		policy.DefaultProvider = strings.TrimSpace(v)
	}
	if v, ok := boolFromAny(values["allow_public_links"]); ok {
		policy.AllowPublicLinks = v
	}
	if v, ok := boolFromAny(values["require_owner_scope"]); ok {
		policy.RequireOwnerScope = v
	}
	if v, ok := boolFromAny(values["audit_downloads"]); ok {
		policy.AuditDownloads = v
	}
	return policy
}

func applyUploadConfig(cfg *rConfig.Config, values map[string]any) {
	if v, ok := int64FromAny(values["max_upload_size"]); ok && v > 0 {
		cfg.MaxUploadSize = v
	}
	if v, ok := stringSliceFromAny(values["allowed_types"]); ok && len(v) > 0 {
		cfg.AllowedTypes = v
	}
	if v, ok := stringFromAny(values["default_storage"]); ok && v != "" {
		cfg.DefaultStorage = v
	}
}

func applyImageConfig(cfg *rConfig.Config, values map[string]any) {
	if cfg.ImageProcessing == nil {
		cfg.ImageProcessing = rConfig.New().ImageProcessing
	}
	if v, ok := boolFromAny(values["enable_thumbnails"]); ok {
		cfg.ImageProcessing.EnableThumbnails = v
	}
	if v, ok := intFromAny(values["default_thumbnail_width"]); ok && v > 0 {
		cfg.ImageProcessing.DefaultThumbnailWidth = v
	}
	if v, ok := intFromAny(values["default_thumbnail_height"]); ok && v > 0 {
		cfg.ImageProcessing.DefaultThumbnailHeight = v
	}
	if v, ok := boolFromAny(values["enable_resizing"]); ok {
		cfg.ImageProcessing.EnableResizing = v
	}
	if v, ok := intFromAny(values["max_image_width"]); ok && v > 0 {
		cfg.ImageProcessing.MaxImageWidth = v
	}
	if v, ok := intFromAny(values["max_image_height"]); ok && v > 0 {
		cfg.ImageProcessing.MaxImageHeight = v
	}
	if v, ok := intFromAny(values["compression_quality"]); ok && v > 0 && v <= 100 {
		cfg.ImageProcessing.CompressionQuality = v
	}
}

func applyQuotaConfig(cfg *rConfig.Config, values map[string]any) {
	if cfg.QuotaManagement == nil {
		cfg.QuotaManagement = rConfig.New().QuotaManagement
	}
	if v, ok := boolFromAny(values["enable_quotas"]); ok {
		cfg.QuotaManagement.EnableQuotas = v
	}
	if v, ok := boolFromAny(values["enable_enforcement"]); ok {
		cfg.QuotaManagement.EnableEnforcement = v
	}
	if v, ok := int64FromAny(values["default_quota"]); ok && v > 0 {
		cfg.QuotaManagement.DefaultQuota = v
	}
	if v, ok := floatFromAny(values["warning_threshold"]); ok && v > 0 && v <= 1 {
		cfg.QuotaManagement.WarningThreshold = v
	}
	if v, ok := stringFromAny(values["quota_check_interval"]); ok && v != "" {
		cfg.QuotaManagement.QuotaCheckInterval = v
	}
}

func normalizeResourceConfig(cfg *rConfig.Config) *rConfig.Config {
	defaults := rConfig.New()
	if cfg.MaxUploadSize <= 0 {
		cfg.MaxUploadSize = defaults.MaxUploadSize
	}
	if len(cfg.AllowedTypes) == 0 {
		cfg.AllowedTypes = defaults.AllowedTypes
	}
	if cfg.DefaultStorage == "" {
		cfg.DefaultStorage = defaults.DefaultStorage
	}
	if cfg.ImageProcessing == nil {
		cfg.ImageProcessing = defaults.ImageProcessing
	}
	if cfg.ImageProcessing.DefaultThumbnailWidth <= 0 {
		cfg.ImageProcessing.DefaultThumbnailWidth = defaults.ImageProcessing.DefaultThumbnailWidth
	}
	if cfg.ImageProcessing.DefaultThumbnailHeight <= 0 {
		cfg.ImageProcessing.DefaultThumbnailHeight = defaults.ImageProcessing.DefaultThumbnailHeight
	}
	if cfg.ImageProcessing.MaxImageWidth <= 0 {
		cfg.ImageProcessing.MaxImageWidth = defaults.ImageProcessing.MaxImageWidth
	}
	if cfg.ImageProcessing.MaxImageHeight <= 0 {
		cfg.ImageProcessing.MaxImageHeight = defaults.ImageProcessing.MaxImageHeight
	}
	if cfg.ImageProcessing.CompressionQuality <= 0 || cfg.ImageProcessing.CompressionQuality > 100 {
		cfg.ImageProcessing.CompressionQuality = defaults.ImageProcessing.CompressionQuality
	}
	if cfg.QuotaManagement == nil {
		cfg.QuotaManagement = defaults.QuotaManagement
	}
	if cfg.QuotaManagement.DefaultQuota <= 0 {
		cfg.QuotaManagement.DefaultQuota = defaults.QuotaManagement.DefaultQuota
	}
	if cfg.QuotaManagement.WarningThreshold <= 0 || cfg.QuotaManagement.WarningThreshold > 1 {
		cfg.QuotaManagement.WarningThreshold = defaults.QuotaManagement.WarningThreshold
	}
	if cfg.QuotaManagement.QuotaCheckInterval == "" {
		cfg.QuotaManagement.QuotaCheckInterval = defaults.QuotaManagement.QuotaCheckInterval
	}
	return cfg
}

func isAllowedResourceType(allowed []string, filename, contentType string) bool {
	if len(allowed) == 0 {
		return true
	}

	normalizedContentType := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	ext := strings.ToLower(filepath.Ext(filename))
	category := string(structs.GetFileCategory(ext))

	for _, item := range allowed {
		rule := strings.ToLower(strings.TrimSpace(item))
		if rule == "" {
			continue
		}
		if rule == "*" || rule == "*/*" {
			return true
		}
		if strings.HasPrefix(rule, ".") && rule == ext {
			return true
		}
		if !strings.Contains(rule, "/") && rule == strings.TrimPrefix(ext, ".") {
			return true
		}
		if !strings.Contains(rule, "/") && rule == category {
			return true
		}
		if normalizedContentType != "" && rule == normalizedContentType {
			return true
		}
		if strings.HasSuffix(rule, "/*") && normalizedContentType != "" {
			prefix := strings.TrimSuffix(rule, "/*")
			if strings.HasPrefix(normalizedContentType, prefix+"/") {
				return true
			}
		}
	}

	return false
}

func resourceTypeLabel(filename, contentType string) string {
	if contentType != "" {
		return contentType
	}
	if ext := filepath.Ext(filename); ext != "" {
		return ext
	}
	return "unknown"
}

func parseDuration(value string) (time.Duration, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("empty duration")
	}
	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil {
		return time.Duration(seconds) * time.Second, nil
	}
	if strings.HasSuffix(value, "d") {
		days, err := strconv.ParseInt(strings.TrimSuffix(value, "d"), 10, 64)
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	if strings.HasSuffix(value, "w") {
		weeks, err := strconv.ParseInt(strings.TrimSuffix(value, "w"), 10, 64)
		if err != nil {
			return 0, err
		}
		return time.Duration(weeks) * 7 * 24 * time.Hour, nil
	}
	return time.ParseDuration(value)
}

func stringFromAny(value any) (string, bool) {
	v, ok := value.(string)
	return v, ok
}

func boolFromAny(value any) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		parsed, err := strconv.ParseBool(v)
		return parsed, err == nil
	default:
		return false, false
	}
}

func intFromAny(value any) (int, bool) {
	v, ok := int64FromAny(value)
	return int(v), ok
}

func int64FromAny(value any) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int64:
		return v, true
	case float64:
		return int64(v), true
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func floatFromAny(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func stringSliceFromAny(value any) ([]string, bool) {
	switch v := value.(type) {
	case []string:
		return v, true
	case []any:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				result = append(result, strings.TrimSpace(s))
			}
		}
		return result, len(result) > 0
	case string:
		if strings.TrimSpace(v) == "" {
			return nil, false
		}
		parts := strings.Split(v, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result, len(result) > 0
	default:
		return nil, false
	}
}
