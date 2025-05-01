package structs

import (
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
)

// FileCategory represents the category of a file
type FileCategory string

// File categories
const (
	FileCategoryImage    FileCategory = "image"
	FileCategoryDocument FileCategory = "document"
	FileCategoryVideo    FileCategory = "video"
	FileCategoryAudio    FileCategory = "audio"
	FileCategoryArchive  FileCategory = "archive"
	FileCategoryOther    FileCategory = "other"
)

// AccessLevel represents the access level for a resource
type AccessLevel string

// Access levels
const (
	AccessLevelPublic  AccessLevel = "public"  // Accessible to anyone
	AccessLevelPrivate AccessLevel = "private" // Only accessible to owner
	AccessLevelShared  AccessLevel = "shared"  // Shared with specific users/groups
)

// FileMetadata represents additional file metadata
type FileMetadata struct {
	Width        *int           `json:"width,omitempty"`         // For images
	Height       *int           `json:"height,omitempty"`        // For images
	Duration     *float64       `json:"duration,omitempty"`      // For audio/video in seconds
	Author       *string        `json:"author,omitempty"`        // Content author
	CreationDate *time.Time     `json:"creation_date,omitempty"` // Original file creation date
	ModifiedDate *time.Time     `json:"modified_date,omitempty"` // Last modification date
	Title        *string        `json:"title,omitempty"`         // Content title
	Description  *string        `json:"description,omitempty"`   // Content description
	Keywords     []string       `json:"keywords,omitempty"`      // Content keywords
	Category     FileCategory   `json:"category,omitempty"`      // File category
	CustomFields map[string]any `json:"custom_fields,omitempty"` // User-defined metadata
}

// FindFile represents the parameters for finding an file.
type FindFile struct {
	File   string `json:"file,omitempty"`
	Tenant string `json:"tenant,omitempty"`
	User   string `json:"user,omitempty"`
}

// FileBody extends FileBody with additional fields
type FileBody struct {
	File              multipart.File     `json:"-"` // For internal use only, not to be serialized
	Name              string             `json:"name,omitempty"`
	Path              string             `json:"path,omitempty"`
	Type              string             `json:"type,omitempty"`
	Size              *int               `json:"size,omitempty"`
	Storage           string             `json:"storage,omitempty"`
	Bucket            string             `json:"bucket,omitempty"`
	Endpoint          string             `json:"endpoint,omitempty"`
	FolderPath        string             `json:"folder_path,omitempty"`        // Virtual folder path
	AccessLevel       AccessLevel        `json:"access_level,omitempty"`       // Access level
	ExpiresAt         *int64             `json:"expires_at,omitempty"`         // Expiration timestamp
	Metadata          *FileMetadata      `json:"metadata,omitempty"`           //  metadata
	Tags              []string           `json:"tags,omitempty"`               // Tags for categorization
	IsPublic          bool               `json:"is_public,omitempty"`          // Publicly accessible flag
	Versions          []string           `json:"versions,omitempty"`           // Previous versions IDs
	ProcessingOptions *ProcessingOptions `json:"processing_options,omitempty"` // Processing options
	ObjectID          string             `json:"object_id,omitempty"`
	TenantID          string             `json:"tenant_id,omitempty"`
	Extras            *types.JSON        `json:"extras,omitempty"`
	CreatedBy         *string            `json:"created_by,omitempty"`
	UpdatedBy         *string            `json:"updated_by,omitempty"`
}

// ProcessingOptions represents options for processing the file
type ProcessingOptions struct {
	CreateThumbnail    bool   `json:"create_thumbnail,omitempty"`
	ResizeImage        bool   `json:"resize_image,omitempty"`
	MaxWidth           int    `json:"max_width,omitempty"`
	MaxHeight          int    `json:"max_height,omitempty"`
	CompressImage      bool   `json:"compress_image,omitempty"`
	CompressionQuality int    `json:"compression_quality,omitempty"` // 1-100
	ConvertFormat      string `json:"convert_format,omitempty"`      // Target format
}

// CreateFileBody represents the body for creating an file
type CreateFileBody struct {
	FileBody
}

// UpdateFileBody represents the body for updating an file
type UpdateFileBody struct {
	ID string `json:"id"`
	FileBody
}

// ReadFile represents the output schema for retrieving an file
type ReadFile struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Path         string        `json:"path"`
	Type         string        `json:"type"`
	Size         *int          `json:"size"`
	Storage      string        `json:"storage"`
	Bucket       string        `json:"bucket"`
	Endpoint     string        `json:"endpoint"`
	FolderPath   string        `json:"folder_path,omitempty"`
	AccessLevel  AccessLevel   `json:"access_level,omitempty"`
	ExpiresAt    *int64        `json:"expires_at,omitempty"`
	Metadata     *FileMetadata `json:"metadata,omitempty"`
	Tags         []string      `json:"tags,omitempty"`
	IsPublic     bool          `json:"is_public,omitempty"`
	Versions     []string      `json:"versions,omitempty"`
	DownloadURL  string        `json:"download_url,omitempty"`
	ThumbnailURL string        `json:"thumbnail_url,omitempty"`
	IsExpired    bool          `json:"is_expired,omitempty"`
	ObjectID     string        `json:"object_id"`
	TenantID     string        `json:"tenant_id"`
	Extras       *types.JSON   `json:"extras,omitempty"`
	CreatedBy    *string       `json:"created_by,omitempty"`
	CreatedAt    *int64        `json:"created_at,omitempty"`
	UpdatedBy    *string       `json:"updated_by,omitempty"`
	UpdatedAt    *int64        `json:"updated_at,omitempty"`
}

// GetCursorValue returns the cursor value.
func (r *ReadFile) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, convert.ToValue(r.CreatedAt))
}

// ListFileParams represents the parameters for listing files
type ListFileParams struct {
	Cursor        string       `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit         int          `form:"limit,omitempty" json:"limit,omitempty"`
	Direction     string       `form:"direction,omitempty" json:"direction,omitempty"`
	Tenant        string       `form:"tenant,omitempty" json:"tenant,omitempty" validate:"required"`
	Object        string       `form:"object,omitempty" json:"object,omitempty" validate:"required"`
	User          string       `form:"user,omitempty" json:"user,omitempty"`
	Type          string       `form:"type,omitempty" json:"type,omitempty"`
	Storage       string       `form:"storage,omitempty" json:"storage,omitempty"`
	Category      FileCategory `form:"category,omitempty" json:"category,omitempty"`
	Tags          string       `form:"tags,omitempty" json:"tags,omitempty"` // Comma-separated tags
	AccessLevel   AccessLevel  `form:"access_level,omitempty" json:"access_level,omitempty"`
	FolderPath    string       `form:"folder_path,omitempty" json:"folder_path,omitempty"`
	CreatedAfter  int64        `form:"created_after,omitempty" json:"created_after,omitempty"`
	CreatedBefore int64        `form:"created_before,omitempty" json:"created_before,omitempty"`
	SizeMin       int64        `form:"size_min,omitempty" json:"size_min,omitempty"`
	SizeMax       int64        `form:"size_max,omitempty" json:"size_max,omitempty"`
	IsPublic      *bool        `form:"is_public,omitempty" json:"is_public,omitempty"`
	SearchQuery   string       `form:"q,omitempty" json:"q,omitempty"` // Full-text search
}

// GetFileCategory determines the file category based on its extension
func GetFileCategory(extension string) FileCategory {
	ext := strings.ToLower(extension)

	// Image formats
	if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" ||
		ext == ".bmp" || ext == ".webp" || ext == ".svg" || ext == ".tiff" {
		return FileCategoryImage
	}

	// Document formats
	if ext == ".pdf" || ext == ".doc" || ext == ".docx" || ext == ".xls" ||
		ext == ".xlsx" || ext == ".ppt" || ext == ".pptx" || ext == ".txt" ||
		ext == ".rtf" || ext == ".md" || ext == ".csv" {
		return FileCategoryDocument
	}

	// Video formats
	if ext == ".mp4" || ext == ".avi" || ext == ".mov" || ext == ".wmv" ||
		ext == ".flv" || ext == ".mkv" || ext == ".webm" {
		return FileCategoryVideo
	}

	// Audio formats
	if ext == ".mp3" || ext == ".wav" || ext == ".ogg" || ext == ".flac" ||
		ext == ".aac" || ext == ".m4a" {
		return FileCategoryAudio
	}

	// Archive formats
	if ext == ".zip" || ext == ".rar" || ext == ".7z" || ext == ".tar" ||
		ext == ".gz" || ext == ".bz2" {
		return FileCategoryArchive
	}

	// Default
	return FileCategoryOther
}
