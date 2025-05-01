package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"ncobase/domain/resource/data"
	"ncobase/domain/resource/data/ent"
	"ncobase/domain/resource/data/repository"
	"ncobase/domain/resource/event"
	"ncobase/domain/resource/structs"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/data/search/meili"
	"github.com/ncobase/ncore/data/storage"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation/validator"
)

// FileServiceInterface represents the file service interface.
type FileServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateFileBody) (*structs.ReadFile, error)
	Update(ctx context.Context, slug string, updates map[string]any) (*structs.ReadFile, error)
	Get(ctx context.Context, slug string) (*structs.ReadFile, error)
	Delete(ctx context.Context, slug string) error
	List(ctx context.Context, params *structs.ListFileParams) (paging.Result[*structs.ReadFile], error)
	GetFileStream(ctx context.Context, slug string) (io.ReadCloser, *structs.ReadFile, error)
	SearchByTags(ctx context.Context, tenantID string, tags []string, limit int) ([]*structs.ReadFile, error)
	GeneratePublicURL(ctx context.Context, slug string, expirationHours int) (string, error)
	CreateVersion(ctx context.Context, slug string, file io.Reader, filename string) (*structs.ReadFile, error)
	GetVersions(ctx context.Context, slug string) ([]*structs.ReadFile, error)
	SetAccessLevel(ctx context.Context, slug string, accessLevel structs.AccessLevel) (*structs.ReadFile, error)
	CreateThumbnail(ctx context.Context, slug string, options *structs.ProcessingOptions) (*structs.ReadFile, error)
}

// fileService is the struct for the file service.
type fileService struct {
	fileRepo       repository.FileRepositoryInterface
	imageProcessor ImageProcessorInterface
	quotaService   QuotaServiceInterface
	publisher      event.PublisherInterface
	meili          *meili.Client
}

// NewFileService creates a new file service.
func NewFileService(
	d *data.Data,
	imageProcessor ImageProcessorInterface,
	quotaService QuotaServiceInterface,
	publisher event.PublisherInterface,
) FileServiceInterface {
	return &fileService{
		fileRepo:       repository.NewFileRepository(d),
		imageProcessor: imageProcessor,
		quotaService:   quotaService,
		publisher:      publisher,
		meili:          d.GetMeilisearch(),
	}
}

// Create creates a new file.
func (s *fileService) Create(ctx context.Context, body *structs.CreateFileBody) (*structs.ReadFile, error) {
	if validator.IsEmpty(body.ObjectID) {
		return nil, errors.New(ecode.FieldIsRequired("belongsTo object"))
	}

	if validator.IsEmpty(body.TenantID) {
		body.TenantID = ctxutil.GetTenantID(ctx)
		if body.TenantID == "" {
			return nil, errors.New(ecode.FieldIsRequired("belongsTo tenant"))
		}
	}

	// Check quota before proceeding
	if s.quotaService != nil && body.Size != nil {
		canProceed, err := s.quotaService.CheckAndUpdateQuota(ctx, body.TenantID, *body.Size)
		if err != nil {
			logger.Warnf(ctx, "Error checking quota: %v", err)
			// Continue despite error, but log it
		} else if !canProceed {
			return nil, errors.New("storage quota exceeded for tenant")
		}
	}

	// Get storage interface
	storage, storageConfig := ctxutil.GetStorage(ctx)

	// Create a copy of the file content for processing
	var fileBytes []byte
	var err error

	if body.File != nil {
		// Read the file once for multiple operations
		fileBytes, err = io.ReadAll(body.File)
		if err != nil {
			logger.Errorf(ctx, "Error reading file: %v", err)
			return nil, errors.New("failed to read file")
		}

		// Close the original reader if it implements Closer
		if closer, ok := body.File.(io.Closer); ok {
			defer closer.Close()
		}
	}

	// Handle file storage
	_, err = storage.Put(body.Path, bytes.NewReader(fileBytes))
	if err != nil {
		logger.Errorf(ctx, "Error storing file: %v", err)
		return nil, errors.New("failed to store file")
	}
	defer func() {
		if err != nil {
			_ = storage.Delete(body.Path)
		}
	}()

	// Set default values if not provided
	if body.AccessLevel == "" {
		body.AccessLevel = structs.AccessLevelPrivate
	}

	// Set file category based on extension if not provided
	if body.Metadata != nil && body.Metadata.Category == "" {
		ext := filepath.Ext(body.Path)
		body.Metadata.Category = structs.GetFileCategory(ext)
	}

	// Assign storage provider info
	body.Storage = storageConfig.Provider
	body.Bucket = storageConfig.Bucket
	body.Endpoint = storageConfig.Endpoint

	// Set created by
	userID := ctxutil.GetUserID(ctx)
	body.CreatedBy = &userID

	// Process image if needed and it's an image file
	thumbnailPath := ""
	if len(fileBytes) > 0 && validator.IsImageFile(body.Path) && s.imageProcessor != nil && body.ProcessingOptions != nil {
		// Create a reader for the file
		fileReader := bytes.NewReader(fileBytes)

		// Create thumbnail if requested
		if body.ProcessingOptions.CreateThumbnail {
			maxWidth := body.ProcessingOptions.MaxWidth
			if maxWidth <= 0 {
				maxWidth = 300 // Default thumbnail width
			}

			maxHeight := body.ProcessingOptions.MaxHeight
			if maxHeight <= 0 {
				maxHeight = 300 // Default thumbnail height
			}

			thumbnailBytes, err := s.imageProcessor.CreateThumbnail(ctx, fileReader, body.Name, maxWidth, maxHeight)
			if err != nil {
				logger.Warnf(ctx, "Error creating thumbnail: %v", err)
			} else {
				// Store the thumbnail
				thumbnailPath = fmt.Sprintf("thumbnails/%s", body.Path)
				_, err = storage.Put(thumbnailPath, bytes.NewReader(thumbnailBytes))
				if err != nil {
					logger.Warnf(ctx, "Error storing thumbnail: %v", err)
					thumbnailPath = ""
				}

				// Update metadata
				if body.Metadata == nil {
					body.Metadata = &structs.FileMetadata{
						Category: structs.GetFileCategory(filepath.Ext(body.Path)),
					}
				}

				if body.Metadata.CustomFields == nil {
					body.Metadata.CustomFields = make(map[string]any)
				}
				body.Metadata.CustomFields["has_thumbnail"] = true
				body.Metadata.CustomFields["thumbnail_path"] = thumbnailPath
			}
		}
	}

	// Prepare extras with extended data
	extendedData := make(types.JSON)

	if body.FolderPath != "" {
		extendedData["folder_path"] = body.FolderPath
	}
	if body.AccessLevel != "" {
		extendedData["access_level"] = string(body.AccessLevel)
	}
	if len(body.Tags) > 0 {
		extendedData["tags"] = body.Tags
	}
	if body.Metadata != nil {
		extendedData["metadata"] = body.Metadata
	}
	if body.IsPublic {
		extendedData["is_public"] = body.IsPublic
	}
	if thumbnailPath != "" {
		extendedData["thumbnail_path"] = thumbnailPath
	}
	if body.ExpiresAt != nil {
		extendedData["expires_at"] = *body.ExpiresAt
	}

	// Merge with existing extras
	if body.Extras != nil {
		for k, v := range *body.Extras {
			extendedData[k] = v
		}
	}

	body.Extras = &extendedData

	// Create the file using the repository
	row, err := s.fileRepo.Create(ctx, body)
	if err != nil {
		logger.Errorf(ctx, "Error creating file: %v", err)
		// Clean up thumbnail if creation failed
		if thumbnailPath != "" {
			_ = storage.Delete(thumbnailPath)
		}
		return nil, errors.New("failed to create file")
	}

	// Publish event
	if s.publisher != nil {
		eventData := &event.FileEventData{
			ID:       row.ID,
			Name:     row.Name,
			Path:     row.Path,
			Type:     row.Type,
			Size:     row.Size,
			Storage:  row.Storage,
			Bucket:   row.Bucket,
			ObjectID: row.ObjectID,
			TenantID: row.TenantID,
			UserID:   userID,
			Extras:   &row.Extras,
		}
		s.publisher.PublishFileCreated(ctx, eventData)
	}

	// Extract extended properties
	folderPath, accessLevel, expiresAt, fileMetadata, tags, isPublic, _ := extractExtendedProperties(row)

	// Return file with extended properties
	return s.Serialize(row, folderPath, accessLevel, expiresAt, fileMetadata, tags, isPublic, thumbnailPath), nil
}

// Update updates an existing file.
func (s *fileService) Update(ctx context.Context, slug string, updates map[string]any) (*structs.ReadFile, error) {
	// Check if ID is empty
	if validator.IsEmpty(slug) {
		return nil, errors.New(ecode.FieldIsRequired("slug"))
	}

	// Check if updates map is empty
	if len(updates) == 0 {
		return nil, errors.New(ecode.FieldIsEmpty("updates fields"))
	}

	// Get existing file to merge changes
	existing, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		return nil, handleEntError(ctx, "File", err)
	}

	// Get storage interface
	storage, storageConfig := ctxutil.GetStorage(ctx)

	// Handle file update if path is included in updates
	if path, ok := updates["path"].(string); ok {
		// Check if the file content is included in the updates
		if fileReader, ok := updates["file"].(io.Reader); ok {
			// Read the entire file content
			fileBytes, err := io.ReadAll(fileReader)
			if err != nil {
				logger.Errorf(ctx, "Error reading file: %v", err)
				return nil, errors.New("error reading uploaded file")
			}

			// Close the original reader if it implements Closer
			if closer, ok := fileReader.(io.Closer); ok {
				closer.Close()
			}

			// Check quota before proceeding with file update
			if s.quotaService != nil && existing.Size > 0 {
				newSize := 0
				if sizeVal, ok := updates["size"].(int); ok {
					newSize = sizeVal
				} else {
					newSize = len(fileBytes) // Use actual size of read content
				}

				// Only check quota if new file is larger
				if newSize > existing.Size {
					sizeDiff := newSize - existing.Size
					canProceed, err := s.quotaService.CheckAndUpdateQuota(ctx, existing.TenantID, sizeDiff)
					if err != nil {
						logger.Warnf(ctx, "Error checking quota: %v", err)
						// Continue despite error, but log it
					} else if !canProceed {
						return nil, errors.New("storage quota exceeded for tenant")
					}
				}
			}

			// Store the file
			if _, err := storage.Put(path, bytes.NewReader(fileBytes)); err != nil {
				logger.Errorf(ctx, "Error updating file: %v", err)
				return nil, errors.New("error updating file")
			}

			// Update storage
			updates["storage"] = storageConfig.Provider
			// Update bucket
			updates["bucket"] = storageConfig.Bucket
			// Update endpoint
			updates["endpoint"] = storageConfig.Endpoint

			// Process image if it's an image file
			if validator.IsImageFile(path) && s.imageProcessor != nil {
				// Get processing options if provided
				var options *structs.ProcessingOptions
				if optionsVal, ok := updates["processing_options"].(*structs.ProcessingOptions); ok {
					options = optionsVal
				} else {
					// Default options
					options = &structs.ProcessingOptions{
						CreateThumbnail: true,
						MaxWidth:        300,
						MaxHeight:       300,
					}
				}

				// Create thumbnail
				if options.CreateThumbnail {
					thumbnailBytes, err := s.imageProcessor.CreateThumbnail(ctx, bytes.NewReader(fileBytes), path, options.MaxWidth, options.MaxHeight)
					if err != nil {
						logger.Warnf(ctx, "Error creating thumbnail: %v", err)
					} else {
						// Store the thumbnail
						thumbnailPath := fmt.Sprintf("thumbnails/%s", path)
						_, err = storage.Put(thumbnailPath, bytes.NewReader(thumbnailBytes))
						if err != nil {
							logger.Warnf(ctx, "Error storing thumbnail: %v", err)
						} else {
							// Update extras with thumbnail path
							extras := getExtrasFromFile(existing)
							extras["thumbnail_path"] = thumbnailPath
							updates["extras"] = extras
						}
					}
				}
			}

			// Remove file from updates after storing to avoid saving the file object itself in DB
			delete(updates, "file")

			// Update size if not explicitly provided
			if _, ok := updates["size"]; !ok {
				updates["size"] = len(fileBytes)
			}
		}
	}

	// Update extended properties
	extras := getExtrasFromFile(existing)

	// Update folder path if provided
	if folderPath, ok := updates["folder_path"].(string); ok {
		extras["folder_path"] = folderPath
		delete(updates, "folder_path")
	}

	// Update access level if provided
	if accessLevel, ok := updates["access_level"].(structs.AccessLevel); ok {
		extras["access_level"] = string(accessLevel)
		delete(updates, "access_level")
	}

	// Update expires_at if provided
	if expiresAt, ok := updates["expires_at"].(*int64); ok {
		extras["expires_at"] = *expiresAt
		delete(updates, "expires_at")
	}

	// Update metadata if provided
	if metadata, ok := updates["metadata"].(*structs.FileMetadata); ok {
		extras["metadata"] = metadata
		delete(updates, "metadata")
	}

	// Update tags if provided
	if tags, ok := updates["tags"].([]string); ok {
		extras["tags"] = tags
		delete(updates, "tags")
	}

	// Update isPublic if provided
	if isPublic, ok := updates["is_public"].(bool); ok {
		extras["is_public"] = isPublic
		delete(updates, "is_public")
	}

	// Add extras back to updates
	updates["extras"] = extras

	// Set updated by
	userID := ctxutil.GetUserID(ctx)
	updates["updated_by"] = userID

	// Update file
	row, err := s.fileRepo.Update(ctx, slug, updates)
	if err != nil {
		return nil, handleEntError(ctx, "File", err)
	}

	// Publish event
	if s.publisher != nil {
		eventData := &event.FileEventData{
			ID:       row.ID,
			Name:     row.Name,
			Path:     row.Path,
			Type:     row.Type,
			Size:     row.Size,
			Storage:  row.Storage,
			Bucket:   row.Bucket,
			ObjectID: row.ObjectID,
			TenantID: row.TenantID,
			UserID:   userID,
			Extras:   &row.Extras,
		}
		s.publisher.PublishFileUpdated(ctx, eventData)
	}

	// Extract extended properties
	folderPath, accessLevel, expiresAt, fileMetadata, tags, isPublic, thumbnailPath := extractExtendedProperties(row)

	// Return extended file
	return s.Serialize(row, folderPath, accessLevel, expiresAt, fileMetadata, tags, isPublic, thumbnailPath), nil
}

// Get retrieves an file by ID.
func (s *fileService) Get(ctx context.Context, slug string) (*structs.ReadFile, error) {
	// Check if ID is empty
	if validator.IsEmpty(slug) {
		return nil, errors.New(ecode.FieldIsRequired("slug"))
	}

	// Get file from repository
	row, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New(ecode.NotExist(fmt.Sprintf("File %s", slug)))
		}
		logger.Errorf(ctx, "Error retrieving file: %v", err)
		return nil, errors.New("error retrieving file")
	}

	// Extract extended properties from extras
	folderPath, accessLevel, expiresAt, fileMetadata, tags, isPublic, thumbnailPath := extractExtendedProperties(row)

	// Generate URLs if public
	downloadURL := ""
	thumbnailURL := ""
	if isPublic {
		// Generate URLs based on storage provider and configuration
		downloadURL = fmt.Sprintf("/res/files/%s?type=download", row.ID)
		if thumbnailPath != "" {
			thumbnailURL = fmt.Sprintf("/res/files/%s?type=thumbnail", row.ID)
		}
	}

	// Publish access event
	if s.publisher != nil {
		userID := ctxutil.GetUserID(ctx)
		eventData := &event.FileEventData{
			ID:       row.ID,
			Name:     row.Name,
			Path:     row.Path,
			Type:     row.Type,
			Size:     row.Size,
			Storage:  row.Storage,
			Bucket:   row.Bucket,
			ObjectID: row.ObjectID,
			TenantID: row.TenantID,
			UserID:   userID,
			Extras:   &row.Extras,
		}
		s.publisher.PublishFileAccessed(ctx, eventData)
	}

	// Create extended file
	result := s.Serialize(row, folderPath, accessLevel, expiresAt, fileMetadata, tags, isPublic, thumbnailPath)

	// Add URLs if public
	if isPublic {
		result.DownloadURL = downloadURL
		result.ThumbnailURL = thumbnailURL
	}

	// Check if expired
	if expiresAt != nil && *expiresAt > 0 {
		now := time.Now().Unix()
		if now > *expiresAt {
			result.IsExpired = true
		}
	}

	return result, nil
}

// Delete deletes an file by ID.
func (s *fileService) Delete(ctx context.Context, slug string) error {
	// Check if ID is empty
	if validator.IsEmpty(slug) {
		return errors.New(ecode.FieldIsRequired("slug"))
	}

	// Get storage interface
	storage, _ := ctxutil.GetStorage(ctx)

	// Get file details before deletion
	row, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		logger.Errorf(ctx, "Error retrieving file: %v", err)
		return errors.New("error retrieving file")
	}

	// Get thumbnail path if exists
	thumbnailPath := ""
	extras := getExtrasFromFile(row)
	if tp, ok := extras["thumbnail_path"].(string); ok {
		thumbnailPath = tp
	}

	// Delete the file from database
	err = s.fileRepo.Delete(ctx, slug)
	if err != nil {
		logger.Errorf(ctx, "Error deleting file: %v", err)
		return errors.New("error deleting file")
	}

	// Delete the file from storage
	if err := storage.Delete(row.Path); err != nil {
		logger.Errorf(ctx, "Error deleting file: %v", err)
		// Continue with deletion process even if file deletion fails
	}

	// Delete thumbnail if exists
	if thumbnailPath != "" {
		if err := storage.Delete(thumbnailPath); err != nil {
			logger.Warnf(ctx, "Error deleting thumbnail: %v", err)
			// Continue even if thumbnail deletion fails
		}
	}

	// Publish delete event
	if s.publisher != nil {
		userID := ctxutil.GetUserID(ctx)
		eventData := &event.FileEventData{
			ID:       row.ID,
			Name:     row.Name,
			Path:     row.Path,
			Type:     row.Type,
			Size:     row.Size,
			Storage:  row.Storage,
			Bucket:   row.Bucket,
			ObjectID: row.ObjectID,
			TenantID: row.TenantID,
			UserID:   userID,
		}
		s.publisher.PublishFileDeleted(ctx, eventData)
	}

	// Remove from Meilisearch if available
	if s.meili != nil {
		if err := s.meili.DeleteDocuments("files", row.ID); err != nil {
			logger.Warnf(ctx, "Error removing file from Meilisearch: %v", err)
			// Continue even if index deletion fails
		}
	}

	return nil
}

// List lists files.
func (s *fileService) List(ctx context.Context, params *structs.ListFileParams) (paging.Result[*structs.ReadFile], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadFile, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, err := s.fileRepo.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			logger.Errorf(ctx, "Error listing files: %v", err)
			return nil, 0, err
		}

		// Filter results based on extended params
		filteredRows := filterFiles(rows, &lp)

		total := s.fileRepo.CountX(ctx, &lp)

		// Convert to enhanced files
		results := make([]*structs.ReadFile, 0, len(filteredRows))
		for _, row := range filteredRows {
			folderPath, accessLevel, expiresAt, fileMetadata, tags, isPublic, thumbnailPath := extractExtendedProperties(row)
			file := s.Serialize(row, folderPath, accessLevel, expiresAt, fileMetadata, tags, isPublic, thumbnailPath)
			results = append(results, file)
		}

		return results, total, nil
	})
}

// GetFileStream retrieves an file's file stream.
func (s *fileService) GetFileStream(ctx context.Context, slug string) (io.ReadCloser, *structs.ReadFile, error) {
	// Check if ID is empty
	if validator.IsEmpty(slug) {
		return nil, nil, errors.New(ecode.FieldIsRequired("slug"))
	}

	// Get storage interface
	storage, _ := ctxutil.GetStorage(ctx)

	// Retrieve file by ID
	row, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil, errors.New(ecode.NotExist(fmt.Sprintf("File %s", slug)))
		}
		logger.Errorf(ctx, "Error retrieving file: %v", err)
		return nil, nil, errors.New("error retrieving file")
	}

	// Check if file is expired
	extras := getExtrasFromFile(row)
	isExpired := false
	if exp, ok := extras["expires_at"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			isExpired = true
		}
	} else if exp, ok := extras["expires_at"].(int64); ok {
		if time.Now().Unix() > exp {
			isExpired = true
		}
	}

	// Don't allow access to expired files
	if isExpired {
		return nil, nil, errors.New("file access has expired")
	}

	// Fetch file stream from storage
	fileStream, err := storage.GetStream(row.Path)
	if err != nil {
		logger.Errorf(ctx, "Error retrieving file stream: %v", err)
		return nil, nil, errors.New("error retrieving file stream")
	}

	// Extract extended properties
	folderPath, accessLevel, expiresAt, fileMetadata, tags, isPublic, thumbnailPath := extractExtendedProperties(row)

	// Return file stream along with file information
	file := s.Serialize(row, folderPath, accessLevel, expiresAt, fileMetadata, tags, isPublic, thumbnailPath)

	return fileStream, file, nil
}

// SearchByTags searches for files by tags
func (s *fileService) SearchByTags(
	ctx context.Context,
	tenantID string,
	tags []string,
	limit int,
) ([]*structs.ReadFile, error) {
	if len(tags) == 0 {
		return nil, errors.New("at least one tag is required")
	}

	// If Meilisearch is available, use it for tag search
	if s.meili != nil {
		query := strings.Join(tags, " ")
		searchParams := &meili.SearchParams{
			Query:  query,
			Limit:  int64(limit),
			Filter: fmt.Sprintf("tenant_id = %s", tenantID),
			Sort:   []string{"created_at:desc"},
			Facets: []string{"tags"},
		}

		results, err := s.meili.Search("files", query, searchParams)
		if err != nil {
			logger.Errorf(ctx, "Error searching Meilisearch: %v", err)
			// Fall back to database search
		} else {
			// Convert search results to files
			files := make([]*structs.ReadFile, 0, len(results.Hits))
			for _, hit := range results.Hits {
				if hitMap, ok := hit.(map[string]any); ok {
					id, _ := hitMap["id"].(string)
					if id != "" {
						file, err := s.Get(ctx, id)
						if err == nil {
							files = append(files, file)
						}
					}
				}
			}
			return files, nil
		}
	}

	// Fall back to database search
	// This is simplified; real implementation would query the database directly
	files, err := s.fileRepo.SearchByTags(ctx, tenantID, tags, limit)
	if err != nil {
		return nil, err
	}

	// Convert to enhanced files
	results := make([]*structs.ReadFile, 0, len(files))
	for _, row := range files {
		folderPath, accessLevel, expiresAt, fileMetadata, tags, isPublic, thumbnailPath := extractExtendedProperties(row)
		file := s.Serialize(row, folderPath, accessLevel, expiresAt, fileMetadata, tags, isPublic, thumbnailPath)
		results = append(results, file)
	}

	return results, nil
}

// GeneratePublicURL generates a temporary public URL for an file
func (s *fileService) GeneratePublicURL(
	ctx context.Context,
	slug string,
	expirationHours int,
) (string, error) {
	// Get file
	row, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		return "", handleEntError(ctx, "File", err)
	}

	// Set expiration time
	if expirationHours <= 0 {
		expirationHours = 24 // Default 24 hours
	}
	expiresAt := time.Now().Add(time.Duration(expirationHours) * time.Hour).Unix()

	// Update file to be public with expiration
	extras := getExtrasFromFile(row)
	extras["is_public"] = true
	extras["expires_at"] = expiresAt

	// Update in database
	_, err = s.fileRepo.Update(ctx, slug, map[string]any{
		"extras": extras,
	})
	if err != nil {
		return "", handleEntError(ctx, "File", err)
	}

	// Generate URL
	// In a real implementation, this might include a signed token
	downloadURL := fmt.Sprintf("/res/files/%s?type=download&token=%d", row.ID, expiresAt)

	return downloadURL, nil
}

// CreateVersion creates a new version of an existing file
func (s *fileService) CreateVersion(
	ctx context.Context,
	slug string,
	file io.Reader,
	filename string,
) (*structs.ReadFile, error) {
	// Get existing file
	existing, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		return nil, handleEntError(ctx, "File", err)
	}

	// Create file header for the new version
	fileHeader := &storage.FileHeader{
		Name: filename,
		Size: 0, // Will be determined after reading the file
		Path: fmt.Sprintf("versions/%s/%s", slug, filename),
		Type: "", // Will be determined by content
	}

	// Get storage interface
	storage, storageConfig := ctxutil.GetStorage(ctx)

	// Read file content to determine size and type
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Close the original reader if it implements Closer
	if closer, ok := file.(io.Closer); ok {
		closer.Close()
	}

	fileHeader.Size = len(fileBytes)
	fileHeader.Type = http.DetectContentType(fileBytes)

	// Create new version of the file
	_, err = storage.Put(fileHeader.Path, bytes.NewReader(fileBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to store file version: %w", err)
	}

	// Extract current versions from extras
	extras := getExtrasFromFile(existing)
	var versions []string
	if v, ok := extras["versions"].([]string); ok {
		versions = v
	} else if v, ok := extras["versions"].([]any); ok {
		versions = make([]string, 0, len(v))
		for _, ver := range v {
			if verStr, ok := ver.(string); ok {
				versions = append(versions, verStr)
			}
		}
	}

	// Add current file ID to versions
	versions = append(versions, existing.ID)

	// Create new file with same properties but new file
	createBody := &structs.CreateFileBody{
		FileBody: structs.FileBody{
			Name:     fileHeader.Name,
			Path:     fileHeader.Path,
			Type:     fileHeader.Type,
			Size:     &fileHeader.Size,
			Storage:  storageConfig.Provider,
			Bucket:   storageConfig.Bucket,
			Endpoint: storageConfig.Endpoint,
			ObjectID: existing.ObjectID,
			TenantID: existing.TenantID,
		},
	}

	// Extract extended properties
	folderPath, accessLevel, expiresAt, fileMetadata, tags, isPublic, _ := extractExtendedProperties(existing)

	// Copy extended properties
	createBody.FolderPath = folderPath
	createBody.AccessLevel = accessLevel
	createBody.ExpiresAt = expiresAt
	createBody.Metadata = fileMetadata
	createBody.Tags = tags
	createBody.IsPublic = isPublic
	createBody.Versions = versions

	// Create new file
	return s.Create(ctx, createBody)
}

// GetVersions retrieves all versions of an file
func (s *fileService) GetVersions(
	ctx context.Context,
	slug string,
) ([]*structs.ReadFile, error) {
	// Get current file
	current, err := s.Get(ctx, slug)
	if err != nil {
		return nil, err
	}

	// Extract versions
	if current.Versions == nil || len(current.Versions) == 0 {
		// No versions, return only current
		return []*structs.ReadFile{current}, nil
	}

	// Get all versions
	versions := make([]*structs.ReadFile, 0, len(current.Versions)+1)

	// Add current version
	versions = append(versions, current)

	// Get previous versions
	for _, versionID := range current.Versions {
		version, err := s.Get(ctx, versionID)
		if err != nil {
			logger.Warnf(ctx, "Error retrieving version %s: %v", versionID, err)
			continue
		}

		versions = append(versions, version)
	}

	return versions, nil
}

// SetAccessLevel sets the access level for an file
func (s *fileService) SetAccessLevel(
	ctx context.Context,
	slug string,
	accessLevel structs.AccessLevel,
) (*structs.ReadFile, error) {
	// Validate access level
	if accessLevel != structs.AccessLevelPublic &&
		accessLevel != structs.AccessLevelPrivate &&
		accessLevel != structs.AccessLevelShared {
		return nil, fmt.Errorf("invalid access level: %s", accessLevel)
	}

	// Get existing file
	row, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		return nil, handleEntError(ctx, "File", err)
	}

	// Update extras
	extras := getExtrasFromFile(row)
	extras["access_level"] = string(accessLevel)

	// Set is_public based on access level
	isPublic := accessLevel == structs.AccessLevelPublic
	extras["is_public"] = isPublic

	// Update file
	updates := map[string]any{
		"extras": extras,
	}

	updated, err := s.fileRepo.Update(ctx, slug, updates)
	if err != nil {
		return nil, handleEntError(ctx, "File", err)
	}

	// Extract extended properties
	folderPath, accessLevel, expiresAt, fileMetadata, tags, isPublic, thumbnailPath := extractExtendedProperties(updated)

	// Return extended file
	return s.Serialize(updated, folderPath, accessLevel, expiresAt, fileMetadata, tags, isPublic, thumbnailPath), nil
}

// CreateThumbnail creates a thumbnail for an image file
func (s *fileService) CreateThumbnail(
	ctx context.Context,
	slug string,
	options *structs.ProcessingOptions,
) (*structs.ReadFile, error) {
	// Get existing file
	row, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		return nil, handleEntError(ctx, "File", err)
	}

	// Check if it's an image
	if !validator.IsImageFile(row.Path) {
		return nil, fmt.Errorf("file is not an image")
	}

	// Validate options
	if options == nil {
		options = &structs.ProcessingOptions{
			CreateThumbnail: true,
			MaxWidth:        300,
			MaxHeight:       300,
		}
	}

	// Get storage interface
	storage, _ := ctxutil.GetStorage(ctx)

	// Get the file
	file, err := storage.Get(row.Path)
	if err != nil {
		return nil, fmt.Errorf("error retrieving file: %w", err)
	}
	defer file.Close()

	// Read file content for processing
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Create thumbnail
	thumbnailBytes, err := s.imageProcessor.CreateThumbnail(
		ctx,
		bytes.NewReader(fileBytes),
		row.Name,
		options.MaxWidth,
		options.MaxHeight,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating thumbnail: %w", err)
	}

	// Store the thumbnail
	thumbnailPath := fmt.Sprintf("thumbnails/%s", row.Path)
	_, err = storage.Put(thumbnailPath, bytes.NewReader(thumbnailBytes))
	if err != nil {
		return nil, fmt.Errorf("error storing thumbnail: %w", err)
	}

	// Update extras
	extras := getExtrasFromFile(row)
	extras["thumbnail_path"] = thumbnailPath

	// Update metadata
	var metadata *structs.FileMetadata
	if meta, ok := extras["metadata"].(*structs.FileMetadata); ok {
		metadata = meta
	} else if metaMap, ok := extras["metadata"].(map[string]any); ok {
		// Convert map to FileMetadata
		metadata = &structs.FileMetadata{
			Category: structs.FileCategoryOther,
		}

		if cat, ok := metaMap["category"].(string); ok {
			metadata.Category = structs.FileCategory(cat)
		}

		if width, ok := metaMap["width"].(int); ok {
			metadata.Width = &width
		}

		if height, ok := metaMap["height"].(int); ok {
			metadata.Height = &height
		}

		if duration, ok := metaMap["duration"].(float64); ok {
			metadata.Duration = &duration
		}

		if custom, ok := metaMap["custom_fields"].(map[string]any); ok {
			metadata.CustomFields = custom
		} else {
			metadata.CustomFields = make(map[string]any)
		}
	} else {
		metadata = &structs.FileMetadata{
			Category:     structs.GetFileCategory(filepath.Ext(row.Path)),
			CustomFields: make(map[string]any),
		}
	}

	metadata.CustomFields["has_thumbnail"] = true
	metadata.CustomFields["thumbnail_path"] = thumbnailPath

	extras["metadata"] = metadata

	// Update file
	updates := map[string]any{
		"extras": extras,
	}

	updated, err := s.fileRepo.Update(ctx, slug, updates)
	if err != nil {
		return nil, handleEntError(ctx, "File", err)
	}

	// Extract extended properties
	folderPath, accessLevel, expiresAt, fileMetadata, tags, isPublic, _ := extractExtendedProperties(updated)

	// Return extended file
	return s.Serialize(updated, folderPath, accessLevel, expiresAt, fileMetadata, tags, isPublic, thumbnailPath), nil
}

// Serialize serializes an file with all properties
func (s *fileService) Serialize(
	row *ent.File,
	folderPath string,
	accessLevel structs.AccessLevel,
	expiresAt *int64,
	metadata *structs.FileMetadata,
	tags []string,
	isPublic bool,
	thumbnailPath string,
) *structs.ReadFile {
	file := &structs.ReadFile{
		ID:          row.ID,
		Name:        row.Name,
		Path:        row.Path,
		Type:        row.Type,
		Size:        &row.Size,
		Storage:     row.Storage,
		Bucket:      row.Bucket,
		Endpoint:    row.Endpoint,
		FolderPath:  folderPath,
		AccessLevel: accessLevel,
		ExpiresAt:   expiresAt,
		Metadata:    metadata,
		Tags:        tags,
		IsPublic:    isPublic,
		ObjectID:    row.ObjectID,
		TenantID:    row.TenantID,
		Extras:      &row.Extras,
		CreatedBy:   &row.CreatedBy,
		CreatedAt:   &row.CreatedAt,
		UpdatedBy:   &row.UpdatedBy,
		UpdatedAt:   &row.UpdatedAt,
	}

	// Add URLs if public
	if isPublic {
		file.DownloadURL = fmt.Sprintf("/res/files/%s?type=download", row.ID)
		if thumbnailPath != "" {
			file.ThumbnailURL = fmt.Sprintf("/res/files/%s?type=thumbnail", row.ID)
		}
	}

	// Check if expired
	if expiresAt != nil && *expiresAt > 0 {
		now := time.Now().Unix()
		if now > *expiresAt {
			file.IsExpired = true
		}
	}

	// Extract versions from extras if available
	extras := getExtrasFromFile(row)
	if versions, ok := extras["versions"].([]string); ok {
		file.Versions = versions
	} else if versions, ok := extras["versions"].([]any); ok {
		stringVersions := make([]string, 0, len(versions))
		for _, v := range versions {
			if vs, ok := v.(string); ok {
				stringVersions = append(stringVersions, vs)
			}
		}
		file.Versions = stringVersions
	}

	return file
}

// getExtrasFromFile extracts extras from an file as a map
func getExtrasFromFile(file *ent.File) map[string]any {
	if file == nil || file.Extras == nil {
		return make(map[string]any)
	}

	// Create new map and copy values
	extras := make(map[string]any)
	for k, v := range file.Extras {
		extras[k] = v
	}

	return extras
}

// extractExtendedProperties extracts extended properties from an file
func extractExtendedProperties(file *ent.File) (
	folderPath string,
	accessLevel structs.AccessLevel,
	expiresAt *int64,
	metadata *structs.FileMetadata,
	tags []string,
	isPublic bool,
	thumbnailPath string,
) {
	// Default values
	accessLevel = structs.AccessLevelPrivate

	// Extract from extras
	extras := getExtrasFromFile(file)

	// Get folder path
	if fp, ok := extras["folder_path"].(string); ok {
		folderPath = fp
	}

	// Get access level
	if al, ok := extras["access_level"].(string); ok {
		accessLevel = structs.AccessLevel(al)
	}

	// Get expires_at
	if exp, ok := extras["expires_at"].(int64); ok {
		expiresAt = &exp
	} else if exp, ok := extras["expires_at"].(float64); ok {
		expInt := int64(exp)
		expiresAt = &expInt
	}

	// Get metadata
	if meta, ok := extras["metadata"].(*structs.FileMetadata); ok {
		metadata = meta
	} else if metaMap, ok := extras["metadata"].(map[string]any); ok {
		// Convert map to FileMetadata
		metadata = &structs.FileMetadata{
			Category: structs.FileCategoryOther,
		}

		if cat, ok := metaMap["category"].(string); ok {
			metadata.Category = structs.FileCategory(cat)
		}

		if width, ok := metaMap["width"].(int); ok {
			metadata.Width = &width
		} else if width, ok := metaMap["width"].(float64); ok {
			widthInt := int(width)
			metadata.Width = &widthInt
		}

		if height, ok := metaMap["height"].(int); ok {
			metadata.Height = &height
		} else if height, ok := metaMap["height"].(float64); ok {
			heightInt := int(height)
			metadata.Height = &heightInt
		}

		if duration, ok := metaMap["duration"].(float64); ok {
			metadata.Duration = &duration
		}

		if custom, ok := metaMap["custom_fields"].(map[string]any); ok {
			metadata.CustomFields = custom
		}
	}

	// Default metadata if not present
	if metadata == nil {
		metadata = &structs.FileMetadata{
			Category: structs.GetFileCategory(filepath.Ext(file.Path)),
		}
	}

	// Get tags
	if t, ok := extras["tags"].([]string); ok {
		tags = t
	} else if tArray, ok := extras["tags"].([]any); ok {
		tags = make([]string, len(tArray))
		for i, tag := range tArray {
			if tagStr, ok := tag.(string); ok {
				tags[i] = tagStr
			}
		}
	}

	// Get is_public
	if ip, ok := extras["is_public"].(bool); ok {
		isPublic = ip
	}

	// Get thumbnail path
	if tp, ok := extras["thumbnail_path"].(string); ok {
		thumbnailPath = tp
	} else if metadata != nil && metadata.CustomFields != nil {
		if tp, ok := metadata.CustomFields["thumbnail_path"].(string); ok {
			thumbnailPath = tp
		}
	}

	return
}

// filterFiles filters files based on extended parameters
func filterFiles(
	files []*ent.File,
	params *structs.ListFileParams,
) []*ent.File {
	if params == nil {
		return files
	}

	filtered := make([]*ent.File, 0, len(files))

	for _, file := range files {
		// Extract extended properties
		folderPath, _, _, metadata, tags, isPublic, _ := extractExtendedProperties(file)

		// Filter by folder path
		if params.FolderPath != "" && folderPath != params.FolderPath {
			continue
		}

		// Filter by category
		if params.Category != "" && (metadata == nil || metadata.Category != params.Category) {
			continue
		}

		// Filter by tags
		if params.Tags != "" {
			requestedTags := strings.Split(params.Tags, ",")
			tagMatch := false

			for _, reqTag := range requestedTags {
				for _, tag := range tags {
					if strings.TrimSpace(reqTag) == tag {
						tagMatch = true
						break
					}
				}
				if tagMatch {
					break
				}
			}

			if !tagMatch {
				continue
			}
		}

		// Filter by is_public
		if params.IsPublic != nil && isPublic != *params.IsPublic {
			continue
		}

		// Filter by created_after
		if params.CreatedAfter > 0 && file.CreatedAt != 0 && file.CreatedAt < params.CreatedAfter {
			continue
		}

		// Filter by created_before
		if params.CreatedBefore > 0 && file.CreatedAt != 0 && file.CreatedAt > params.CreatedBefore {
			continue
		}

		// Filter by size range
		if params.SizeMin > 0 && file.Size < int(params.SizeMin) {
			continue
		}
		if params.SizeMax > 0 && file.Size > int(params.SizeMax) {
			continue
		}

		// Add to filtered results
		filtered = append(filtered, file)
	}

	return filtered
}
