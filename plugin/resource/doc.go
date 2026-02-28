// Package resource provides file storage and resource management for the Ncobase platform.
//
// This package implements a comprehensive file management system with support for
// uploads, versioning, thumbnails, metadata, and storage quotas. It provides a
// unified API for file operations across different storage backends.
//
// Key Features:
//   - File upload and download
//   - Multiple storage backends (local, S3, OSS)
//   - File versioning and history
//   - Thumbnail generation for images
//   - Metadata and tagging
//   - Storage quota management
//   - File organization (folders, collections)
//   - Access control and permissions
//   - Batch operations
//
// Main Components:
//   - Handler: HTTP endpoints for file operations
//   - Service: Business logic for resource management
//   - Repository: Data access layer for file metadata
//   - Storage: Backend storage adapters
//
// Supported Storage Backends:
//   - Local: Local filesystem storage
//   - S3: Amazon S3 and S3-compatible services
//   - OSS: Alibaba Cloud Object Storage Service
//
// Resource Entities:
//   - File: Individual file with metadata
//   - Version: File version history
//   - Thumbnail: Generated image thumbnails
//   - Folder: File organization structure
//   - Tag: File classification and search
//
// Usage Example:
//
//	// Upload a file
//	file, err := resourceService.Upload(ctx, &structs.UploadInput{
//	    File:        fileReader,
//	    Filename:    "document.pdf",
//	    ContentType: "application/pdf",
//	    FolderID:    folderID,
//	    Tags:        []string{"contract", "legal"},
//	})
//	if err != nil {
//	    return err
//	}
//
//	// Generate thumbnail
//	thumbnail, err := resourceService.GenerateThumbnail(ctx, file.ID, &structs.ThumbnailInput{
//	    Width:   200,
//	    Height:  200,
//	    Quality: 85,
//	})
//
//	// Create folder
//	folder, err := resourceService.CreateFolder(ctx, &structs.CreateFolderInput{
//	    Name:     "Documents",
//	    ParentID: parentFolderID,
//	})
//
//	// Get file versions
//	versions, err := resourceService.GetVersions(ctx, file.ID)
//
//	// Check storage quota
//	quota, err := resourceService.GetQuota(ctx, spaceID)
//	if quota.Used >= quota.Limit {
//	    return errors.New("storage quota exceeded")
//	}
//
// File Operations:
//   - Upload: Single and multipart uploads
//   - Download: Direct and streaming downloads
//   - Copy/Move: File organization operations
//   - Delete: Soft and hard deletion
//   - Search: Full-text and metadata search
//
// The package provides enterprise-grade file management with support for multiple
// storage backends, versioning, and comprehensive metadata management.
package resource
