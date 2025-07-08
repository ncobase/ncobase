package structs

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
)

// FileCategory represents file category
type FileCategory string

const (
	FileCategoryImage    FileCategory = "image"
	FileCategoryDocument FileCategory = "document"
	FileCategoryVideo    FileCategory = "video"
	FileCategoryAudio    FileCategory = "audio"
	FileCategoryArchive  FileCategory = "archive"
	FileCategoryOther    FileCategory = "other"
)

// AccessLevel represents file access level
type AccessLevel string

const (
	AccessLevelPublic  AccessLevel = "public"
	AccessLevelPrivate AccessLevel = "private"
	AccessLevelShared  AccessLevel = "shared"
)

// ProcessingOptions for file processing
type ProcessingOptions struct {
	CreateThumbnail    bool   `json:"create_thumbnail,omitempty"`
	ResizeImage        bool   `json:"resize_image,omitempty"`
	MaxWidth           int    `json:"max_width,omitempty"`
	MaxHeight          int    `json:"max_height,omitempty"`
	CompressImage      bool   `json:"compress_image,omitempty"`
	CompressionQuality int    `json:"compression_quality,omitempty"` // 1-100
	ConvertFormat      string `json:"convert_format,omitempty"`
}

// CreateFileBody for creating files
type CreateFileBody struct {
	File              multipart.File     `json:"-"`
	Name              string             `json:"name,omitempty"`
	OriginalName      string             `json:"original_name,omitempty"`
	Path              string             `json:"path,omitempty"`
	PathPrefix        string             `json:"path_prefix,omitempty"`
	Type              string             `json:"type,omitempty"`
	Size              *int               `json:"size,omitempty"`
	Storage           string             `json:"storage,omitempty"`
	Bucket            string             `json:"bucket,omitempty"`
	Endpoint          string             `json:"endpoint,omitempty"`
	AccessLevel       AccessLevel        `json:"access_level,omitempty"`
	ExpiresAt         *int64             `json:"expires_at,omitempty"`
	Tags              []string           `json:"tags,omitempty"`
	IsPublic          bool               `json:"is_public,omitempty"`
	ProcessingOptions *ProcessingOptions `json:"processing_options,omitempty"`
	OwnerID           string             `json:"owner_id,omitempty"`
	Extras            *types.JSON        `json:"extras,omitempty"`
	CreatedBy         *string            `json:"created_by,omitempty"`
	UpdatedBy         *string            `json:"updated_by,omitempty"`
}

// Validate validates the create file body
func (c *CreateFileBody) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if c.Size == nil || *c.Size <= 0 {
		return fmt.Errorf("valid size is required")
	}
	if c.File == nil {
		return fmt.Errorf("file content is required")
	}
	return nil
}

// GetFolderPath returns folder path
func (c *CreateFileBody) GetFolderPath() string {
	if c.Path == "" {
		return ""
	}
	dir := filepath.Dir(c.Path)
	if dir == "." || dir == "/" {
		return ""
	}
	return filepath.ToSlash(dir)
}

// ReadFile represents file output with context-aware serialization
type ReadFile struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	OriginalName string `json:"original_name,omitempty"`
	Path         string `json:"path"`
	Type         string `json:"type"`
	Size         *int   `json:"size"`

	// Storage fields
	Storage  string `json:"storage,omitempty"`
	Bucket   string `json:"bucket,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`

	AccessLevel  AccessLevel  `json:"access_level,omitempty"`
	ExpiresAt    *int64       `json:"expires_at,omitempty"`
	Tags         []string     `json:"tags,omitempty"`
	IsPublic     bool         `json:"is_public,omitempty"`
	Category     FileCategory `json:"category,omitempty"`
	DownloadURL  string       `json:"download_url,omitempty"`
	ThumbnailURL string       `json:"thumbnail_url,omitempty"`
	IsExpired    bool         `json:"is_expired,omitempty"`

	// Sensitive fields
	Hash      string      `json:"hash,omitempty"`
	OwnerID   string      `json:"owner_id,omitempty"`
	Extras    *types.JSON `json:"extras,omitempty"`
	CreatedBy *string     `json:"created_by,omitempty"`
	UpdatedBy *string     `json:"updated_by,omitempty"`

	CreatedAt *int64 `json:"created_at,omitempty"`
	UpdatedAt *int64 `json:"updated_at,omitempty"`

	// Virtual fields
	FullPath string `json:"full_path,omitempty"`

	// Context flag
	isPublicContext bool `json:"-"`
}

// PublicView returns copy configured for public API
func (r *ReadFile) PublicView() *ReadFile {
	return &ReadFile{
		ID:              r.ID,
		Name:            r.Name,
		OriginalName:    r.OriginalName,
		Path:            r.GetSafePath(),
		Type:            r.Type,
		Size:            r.Size,
		Category:        r.Category,
		DownloadURL:     r.DownloadURL,
		ThumbnailURL:    r.ThumbnailURL,
		IsExpired:       r.IsExpired,
		Tags:            r.Tags,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
		FullPath:        r.GetSafePath(),
		isPublicContext: true,
	}
}

// InternalView returns copy configured for internal API
func (r *ReadFile) InternalView() *ReadFile {
	internal := *r
	internal.isPublicContext = false
	internal.FullPath = r.GetLogicalPath()
	return &internal
}

// GetFullPath returns complete storage path including endpoint
func (r *ReadFile) GetFullPath() string {
	if r.Path == "" {
		return ""
	}
	if r.Endpoint != "" && r.Bucket != "" {
		return fmt.Sprintf("%s/%s/%s", strings.TrimSuffix(r.Endpoint, "/"), r.Bucket, r.Path)
	}
	if r.Storage == "filesystem" {
		return r.Path
	}
	return r.Path
}

// GetSafePath returns safe path for public display
func (r *ReadFile) GetSafePath() string {
	if r.Path == "" {
		return ""
	}
	parts := strings.Split(r.Path, "/")
	if len(parts) > 1 {
		if len(parts) >= 2 {
			return strings.Join(parts[len(parts)-2:], "/")
		}
	}
	return filepath.Base(r.Path)
}

// GetLogicalPath returns user-friendly logical path
func (r *ReadFile) GetLogicalPath() string {
	if r.Path == "" {
		return ""
	}

	// Check for path prefix in extras
	if r.Extras != nil {
		if pathPrefix, ok := (*r.Extras)["path_prefix"].(string); ok && pathPrefix != "" {
			return r.buildFriendlyPath(pathPrefix)
		}
	}

	return r.Path
}

// GetDisplayPath returns path based on context
func (r *ReadFile) GetDisplayPath() string {
	if r.isPublicContext {
		return r.GetSafePath()
	}
	return r.GetLogicalPath()
}

// buildFriendlyPath constructs user-friendly path display
func (r *ReadFile) buildFriendlyPath(pathPrefix string) string {
	if r.OriginalName != "" {
		return fmt.Sprintf("%s/%s", pathPrefix, r.OriginalName)
	}

	filename := filepath.Base(r.Path)
	cleanFilename := r.extractCleanFilename(filename)
	return fmt.Sprintf("%s/%s", pathPrefix, cleanFilename)
}

// extractCleanFilename extracts clean filename from timestamped filename
func (r *ReadFile) extractCleanFilename(filename string) string {
	parts := strings.Split(filename, "_")
	if len(parts) >= 3 {
		// Skip timestamp and random ID parts
		cleanParts := parts[2:]
		return strings.Join(cleanParts, "_")
	}
	return filename
}

// GetDirectory returns directory part of path
func (r *ReadFile) GetDirectory() string {
	if r.Path == "" {
		return ""
	}
	dir := filepath.Dir(r.Path)
	if dir == "." || dir == "/" {
		return ""
	}
	return filepath.ToSlash(dir)
}

// GetFilename returns filename with extension
func (r *ReadFile) GetFilename() string {
	if r.OriginalName != "" {
		return r.OriginalName
	}
	if r.Path == "" {
		return r.Name
	}
	return filepath.Base(r.Path)
}

// GetBasename returns filename without extension
func (r *ReadFile) GetBasename() string {
	filename := r.GetFilename()
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
}

// GetExtension returns file extension
func (r *ReadFile) GetExtension() string {
	if r.Path != "" {
		return filepath.Ext(r.Path)
	}
	if r.OriginalName != "" {
		return filepath.Ext(r.OriginalName)
	}
	return ""
}

// IsImage checks if file is an image
func (r *ReadFile) IsImage() bool {
	return r.Category == FileCategoryImage
}

// GetCursorValue returns cursor value for pagination
func (r *ReadFile) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, convert.ToValue(r.CreatedAt))
}

// ToContentReference converts to content module reference
func (r *ReadFile) ToContentReference() *ContentResourceFileReference {
	return &ContentResourceFileReference{
		ID:           r.ID,
		Name:         r.Name,
		OriginalName: r.OriginalName,
		Path:         r.GetSafePath(),
		Type:         r.Type,
		Size:         r.Size,
		Storage:      "resource",
		DownloadURL:  r.DownloadURL,
		ThumbnailURL: r.ThumbnailURL,
		IsExpired:    r.IsExpired,
	}
}

// ContentResourceFileReference for content module integration
type ContentResourceFileReference struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	OriginalName string `json:"original_name,omitempty"`
	Path         string `json:"path"`
	Type         string `json:"type"`
	Size         *int   `json:"size"`
	Storage      string `json:"storage"`
	DownloadURL  string `json:"download_url,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	IsExpired    bool   `json:"is_expired,omitempty"`
}

// ListFileParams for file listing with filtering
type ListFileParams struct {
	Cursor        string       `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit         int          `form:"limit,omitempty" json:"limit,omitempty"`
	Direction     string       `form:"direction,omitempty" json:"direction,omitempty"`
	OwnerID       string       `form:"owner_id,omitempty" json:"owner_id,omitempty" validate:"required"`
	User          string       `form:"user,omitempty" json:"user,omitempty"`
	Type          string       `form:"type,omitempty" json:"type,omitempty"`
	Storage       string       `form:"storage,omitempty" json:"storage,omitempty"`
	Category      FileCategory `form:"category,omitempty" json:"category,omitempty"`
	Tags          string       `form:"tags,omitempty" json:"tags,omitempty"`
	AccessLevel   AccessLevel  `form:"access_level,omitempty" json:"access_level,omitempty"`
	PathPrefix    string       `form:"path_prefix,omitempty" json:"path_prefix,omitempty"`
	CreatedAfter  int64        `form:"created_after,omitempty" json:"created_after,omitempty"`
	CreatedBefore int64        `form:"created_before,omitempty" json:"created_before,omitempty"`
	SizeMin       int64        `form:"size_min,omitempty" json:"size_min,omitempty"`
	SizeMax       int64        `form:"size_max,omitempty" json:"size_max,omitempty"`
	IsPublic      *bool        `form:"is_public,omitempty" json:"is_public,omitempty"`
	SearchQuery   string       `form:"q,omitempty" json:"q,omitempty"`
}

// FindFile for finding files
type FindFile struct {
	File    string `json:"file,omitempty"`
	OwnerID string `json:"owner_id,omitempty"`
	User    string `json:"user,omitempty"`
}

// GetFileCategory determines file category from extension
func GetFileCategory(extension string) FileCategory {
	ext := strings.ToLower(extension)

	imageExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".bmp": true, ".webp": true, ".svg": true, ".tiff": true,
		".ico": true,
	}
	documentExts := map[string]bool{
		".pdf": true, ".doc": true, ".docx": true, ".xls": true,
		".xlsx": true, ".ppt": true, ".pptx": true, ".txt": true,
		".rtf": true, ".md": true, ".csv": true,
	}
	videoExts := map[string]bool{
		".mp4": true, ".avi": true, ".mov": true, ".wmv": true,
		".flv": true, ".mkv": true, ".webm": true, ".m4v": true,
	}
	audioExts := map[string]bool{
		".mp3": true, ".wav": true, ".ogg": true, ".flac": true,
		".aac": true, ".m4a": true, ".wma": true,
	}
	archiveExts := map[string]bool{
		".zip": true, ".rar": true, ".7z": true, ".tar": true,
		".gz": true, ".bz2": true,
	}

	if imageExts[ext] {
		return FileCategoryImage
	}
	if documentExts[ext] {
		return FileCategoryDocument
	}
	if videoExts[ext] {
		return FileCategoryVideo
	}
	if audioExts[ext] {
		return FileCategoryAudio
	}
	if archiveExts[ext] {
		return FileCategoryArchive
	}

	return FileCategoryOther
}
