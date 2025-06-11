package structs

import (
	"fmt"
	"mime/multipart"
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
	Path              string             `json:"path,omitempty"`
	Type              string             `json:"type,omitempty"`
	Size              *int               `json:"size,omitempty"`
	Storage           string             `json:"storage,omitempty"`
	Bucket            string             `json:"bucket,omitempty"`
	Endpoint          string             `json:"endpoint,omitempty"`
	FolderPath        string             `json:"folder_path,omitempty"`
	AccessLevel       AccessLevel        `json:"access_level,omitempty"`
	ExpiresAt         *int64             `json:"expires_at,omitempty"`
	Tags              []string           `json:"tags,omitempty"`
	IsPublic          bool               `json:"is_public,omitempty"`
	ProcessingOptions *ProcessingOptions `json:"processing_options,omitempty"`
	OwnerID           string             `json:"owner_id,omitempty"`
	SpaceID           string             `json:"space_id,omitempty"`
	Extras            *types.JSON        `json:"extras,omitempty"`
	CreatedBy         *string            `json:"created_by,omitempty"`
	UpdatedBy         *string            `json:"updated_by,omitempty"`
}

// ReadFile represents file output
type ReadFile struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Path         string       `json:"path"`
	Type         string       `json:"type"`
	Size         *int         `json:"size"`
	Storage      string       `json:"storage"`
	Bucket       string       `json:"bucket"`
	Endpoint     string       `json:"endpoint"`
	FolderPath   string       `json:"folder_path,omitempty"`
	AccessLevel  AccessLevel  `json:"access_level,omitempty"`
	ExpiresAt    *int64       `json:"expires_at,omitempty"`
	Tags         []string     `json:"tags,omitempty"`
	IsPublic     bool         `json:"is_public,omitempty"`
	Category     FileCategory `json:"category,omitempty"`
	DownloadURL  string       `json:"download_url,omitempty"`
	ThumbnailURL string       `json:"thumbnail_url,omitempty"`
	IsExpired    bool         `json:"is_expired,omitempty"`
	OwnerID      string       `json:"owner_id"`
	SpaceID      string       `json:"space_id"`
	Extras       *types.JSON  `json:"extras,omitempty"`
	CreatedBy    *string      `json:"created_by,omitempty"`
	CreatedAt    *int64       `json:"created_at,omitempty"`
	UpdatedBy    *string      `json:"updated_by,omitempty"`
	UpdatedAt    *int64       `json:"updated_at,omitempty"`
}

// GetCursorValue returns cursor value for pagination
func (r *ReadFile) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, convert.ToValue(r.CreatedAt))
}

// ToContentReference converts ReadFile to content module's ResourceFileReference
func (r *ReadFile) ToContentReference() *ContentResourceFileReference {
	return &ContentResourceFileReference{
		ID:           r.ID,
		Name:         r.Name,
		Path:         r.Path,
		Type:         r.Type,
		Size:         r.Size,
		Storage:      r.Storage,
		DownloadURL:  r.DownloadURL,
		ThumbnailURL: r.ThumbnailURL,
		IsExpired:    r.IsExpired,
	}
}

// ContentResourceFileReference represents resource file reference for content module
type ContentResourceFileReference struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	Type         string `json:"type"`
	Size         *int   `json:"size"`
	Storage      string `json:"storage"`
	DownloadURL  string `json:"download_url,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	IsExpired    bool   `json:"is_expired,omitempty"`
}

// ListFileParams for listing files
type ListFileParams struct {
	Cursor        string       `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit         int          `form:"limit,omitempty" json:"limit,omitempty"`
	Direction     string       `form:"direction,omitempty" json:"direction,omitempty"`
	SpaceID       string       `form:"space_id,omitempty" json:"space_id,omitempty" validate:"required"`
	OwnerID       string       `form:"owner_id,omitempty" json:"owner_id,omitempty" validate:"required"`
	User          string       `form:"user,omitempty" json:"user,omitempty"`
	Type          string       `form:"type,omitempty" json:"type,omitempty"`
	Storage       string       `form:"storage,omitempty" json:"storage,omitempty"`
	Category      FileCategory `form:"category,omitempty" json:"category,omitempty"`
	Tags          string       `form:"tags,omitempty" json:"tags,omitempty"`
	AccessLevel   AccessLevel  `form:"access_level,omitempty" json:"access_level,omitempty"`
	FolderPath    string       `form:"folder_path,omitempty" json:"folder_path,omitempty"`
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
	SpaceID string `json:"space_id,omitempty"`
	User    string `json:"user,omitempty"`
}

// GetFileCategory determines file category from extension
func GetFileCategory(extension string) FileCategory {
	ext := strings.ToLower(extension)

	imageExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".bmp": true, ".webp": true, ".svg": true, ".tiff": true,
	}
	documentExts := map[string]bool{
		".pdf": true, ".doc": true, ".docx": true, ".xls": true,
		".xlsx": true, ".ppt": true, ".pptx": true, ".txt": true,
		".rtf": true, ".md": true, ".csv": true,
	}
	videoExts := map[string]bool{
		".mp4": true, ".avi": true, ".mov": true, ".wmv": true,
		".flv": true, ".mkv": true, ".webm": true,
	}
	audioExts := map[string]bool{
		".mp3": true, ".wav": true, ".ogg": true, ".flac": true,
		".aac": true, ".m4a": true,
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
