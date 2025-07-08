package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"ncobase/resource/data"
	"ncobase/resource/data/ent"
	"ncobase/resource/data/repository"
	"ncobase/resource/event"
	"ncobase/resource/structs"
	"path/filepath"
	"strings"
	"time"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"
)

type FileServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateFileBody) (*structs.ReadFile, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*structs.ReadFile, error)
	Get(ctx context.Context, slug string) (*structs.ReadFile, error)
	GetPublic(ctx context.Context, slug string) (*structs.ReadFile, error)
	GetByShareToken(ctx context.Context, token string) (*structs.ReadFile, error)
	Delete(ctx context.Context, slug string) error
	List(ctx context.Context, params *structs.ListFileParams) (paging.Result[*structs.ReadFile], error)
	GetFileStream(ctx context.Context, slug string) (io.ReadCloser, *structs.ReadFile, error)
	GetFileStreamByID(ctx context.Context, id string) (io.ReadCloser, error)
	GetThumbnail(ctx context.Context, slug string) (io.ReadCloser, error)
	SearchByTags(ctx context.Context, ownerID string, tags []string, limit int) ([]*structs.ReadFile, error)
	GeneratePublicURL(ctx context.Context, slug string, expirationHours int) (string, error)
	CreateVersion(ctx context.Context, slug string, file io.Reader, filename string) (*structs.ReadFile, error)
	GetVersions(ctx context.Context, slug string) ([]*structs.ReadFile, error)
	SetAccessLevel(ctx context.Context, slug string, accessLevel structs.AccessLevel) (*structs.ReadFile, error)
	CreateThumbnail(ctx context.Context, slug string, options *structs.ProcessingOptions) (*structs.ReadFile, error)
	GetTagsByOwner(ctx context.Context, ownerID string) ([]string, error)
}

type fileService struct {
	fileRepo       repository.FileRepositoryInterface
	imageProcessor ImageProcessorInterface
	quotaService   QuotaServiceInterface
	publisher      event.PublisherInterface
}

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
	}
}

// Create creates a new file
func (s *fileService) Create(ctx context.Context, body *structs.CreateFileBody) (*structs.ReadFile, error) {
	// Get ownerID from context if not provided
	if body.OwnerID == "" {
		if userID := ctxutil.GetUserID(ctx); userID != "" {
			body.OwnerID = userID
		}
	}

	// Check quota only if ownerID is provided
	if body.OwnerID != "" && s.quotaService != nil && body.Size != nil {
		canProceed, err := s.quotaService.CheckAndUpdateQuota(ctx, body.OwnerID, *body.Size)
		if err != nil {
			logger.Warnf(ctx, "Error checking quota: %v", err)
		} else if !canProceed {
			return nil, errors.New("storage quota exceeded")
		}
	}

	// Get storage
	storageClient, storageConfig := ctxutil.GetStorage(ctx)
	if storageClient == nil || storageConfig == nil {
		return nil, errors.New("storage not configured")
	}

	// Read file content and calculate hash
	var fileBytes []byte
	var err error
	if body.File != nil {
		fileBytes, err = io.ReadAll(body.File)
		if err != nil {
			logger.Errorf(ctx, "Error reading file: %v", err)
			return nil, errors.New("failed to read file")
		}
		if closer, ok := body.File.(io.Closer); ok {
			defer closer.Close()
		}
	} else {
		return nil, errors.New("file content is required")
	}

	// Calculate file hash for deduplication
	hash := calculateFileHash(fileBytes)

	// Check for existing file with same hash (optional deduplication)
	if body.OwnerID != "" && hash != "" {
		existing, err := findFileByHash(ctx, body.OwnerID, hash)
		if err == nil && existing != nil {
			logger.Infof(ctx, "File with same hash already exists: %s", existing.ID)
			// Could return existing file or continue with new upload based on business logic
		}
	}

	// Generate storage path with optional parameters
	ext := filepath.Ext(body.Path)
	if ext == "" && body.Name != "" {
		if body.Type != "" {
			ext = s.getExtensionFromMimeType(body.Type)
		}
	}

	var ownerIDPtr, pathPrefixPtr *string
	if body.OwnerID != "" {
		ownerIDPtr = &body.OwnerID
	}
	if body.PathPrefix != "" {
		pathPrefixPtr = &body.PathPrefix
	}

	storagePath := s.generateUniqueStoragePath(body.Name, ext, ownerIDPtr, pathPrefixPtr)

	// Store file
	_, storeErr := storageClient.Put(storagePath, bytes.NewReader(fileBytes))
	if storeErr != nil {
		logger.Errorf(ctx, "Error storing file to %s: %v", storageConfig.Provider, storeErr)
		return nil, fmt.Errorf("failed to store file: %w", storeErr)
	}

	// Cleanup on error
	defer func() {
		if err != nil {
			if deleteErr := storageClient.Delete(storagePath); deleteErr != nil {
				logger.Errorf(ctx, "Failed to cleanup file after error: %v", deleteErr)
			}
		}
	}()

	// Set defaults and computed values
	if body.AccessLevel == "" {
		body.AccessLevel = structs.AccessLevelPrivate
	}

	// Set storage info
	body.Storage = storageConfig.Provider
	body.Bucket = storageConfig.Bucket
	body.Endpoint = storageConfig.Endpoint
	body.Path = storagePath

	// Set audit fields
	userID := ctxutil.GetUserID(ctx)
	if userID != "" {
		body.CreatedBy = &userID
	}

	// Process image if needed
	thumbnailPath := ""
	category := structs.GetFileCategory(filepath.Ext(storagePath))

	if category == structs.FileCategoryImage && s.imageProcessor != nil {
		if body.ProcessingOptions == nil {
			body.ProcessingOptions = &structs.ProcessingOptions{
				CreateThumbnail: true,
				MaxWidth:        300,
				MaxHeight:       300,
			}
		}

		thumbnailBytes, err := s.imageProcessor.CreateThumbnail(
			ctx,
			bytes.NewReader(fileBytes),
			body.Name,
			body.ProcessingOptions.MaxWidth,
			body.ProcessingOptions.MaxHeight,
		)

		if err == nil {
			thumbnailPath = s.generateThumbnailPath(storagePath)
			_, err = storageClient.Put(thumbnailPath, bytes.NewReader(thumbnailBytes))
			if err != nil {
				logger.Warnf(ctx, "Error storing thumbnail: %v", err)
				thumbnailPath = ""
			}
		}
	}

	// Prepare extras with all metadata
	extendedData := make(types.JSON)
	if body.Extras != nil {
		for k, v := range *body.Extras {
			extendedData[k] = v
		}
	}

	// Add computed metadata
	if thumbnailPath != "" {
		extendedData["thumbnail_path"] = thumbnailPath
	}
	if body.PathPrefix != "" {
		extendedData["path_prefix"] = body.PathPrefix
	}
	if body.OwnerID == "" {
		extendedData["anonymous"] = true
	}
	if hash != "" {
		extendedData["hash"] = hash // Also store in extras for backward compatibility
	}

	body.Extras = &extendedData

	// Create file record with retry logic for name conflicts
	maxRetries := 3
	for retry := 0; retry < maxRetries; retry++ {
		row, err := s.fileRepo.Create(ctx, body)
		if err != nil {
			if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
				body.Name = s.generateUniqueName(body.Name)
				logger.Warnf(ctx, "Name conflict, retrying with new name: %s", body.Name)
				continue
			}

			logger.Errorf(ctx, "Error creating file record: %v", err)
			if thumbnailPath != "" {
				_ = storageClient.Delete(thumbnailPath)
			}
			return nil, errors.New("failed to create file record")
		}

		// Success - publish event
		if s.publisher != nil {
			eventUserID := userID
			if eventUserID == "" {
				eventUserID = body.OwnerID
			}

			eventData := &event.FileEventData{
				ID:      row.ID,
				Name:    row.Name,
				Path:    row.Path,
				Type:    row.Type,
				Size:    row.Size,
				Storage: row.Storage,
				Bucket:  row.Bucket,
				OwnerID: row.OwnerID,
				UserID:  eventUserID,
				Extras:  &row.Extras,
			}
			s.publisher.PublishFileCreated(ctx, eventData)
		}

		return s.serialize(row), nil
	}

	return nil, errors.New("failed to create file record after retries")
}

// Update updates file
func (s *fileService) Update(ctx context.Context, slug string, updates types.JSON) (*structs.ReadFile, error) {
	if validator.IsEmpty(slug) {
		return nil, errors.New(ecode.FieldIsRequired("slug"))
	}

	if len(updates) == 0 {
		return nil, errors.New(ecode.FieldIsEmpty("updates fields"))
	}

	// Get existing file
	existing, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		return nil, handleEntError(ctx, "File", err)
	}

	// Handle file update with hash calculation
	if fileReader, ok := updates["file"].(io.Reader); ok {
		storageClient, storageConfig := ctxutil.GetStorage(ctx)
		if storageClient == nil || storageConfig == nil {
			return nil, errors.New("storage not configured")
		}

		fileBytes, err := io.ReadAll(fileReader)
		if err != nil {
			return nil, errors.New("error reading uploaded file")
		}

		if closer, ok := fileReader.(io.Closer); ok {
			closer.Close()
		}

		// Calculate new hash
		newHash := calculateFileHash(fileBytes)

		// Generate new storage path
		fileName := existing.Name
		if name, ok := updates["name"].(string); ok {
			fileName = name
		}

		ext := filepath.Ext(existing.Path)
		extras := getExtrasFromFile(existing)

		var ownerIDPtr, pathPrefixPtr *string
		if existing.OwnerID != "" {
			ownerIDPtr = &existing.OwnerID
		}
		if pathPrefix, hasPrefix := extras["path_prefix"].(string); hasPrefix && pathPrefix != "" {
			pathPrefixPtr = &pathPrefix
		}

		newStoragePath := s.generateUniqueStoragePath(fileName, ext, ownerIDPtr, pathPrefixPtr)

		// Store new file
		if _, err := storageClient.Put(newStoragePath, bytes.NewReader(fileBytes)); err != nil {
			logger.Errorf(ctx, "Error updating file in storage: %v", err)
			return nil, errors.New("error updating file")
		}

		// Delete old file
		if err := storageClient.Delete(existing.Path); err != nil {
			logger.Warnf(ctx, "Error deleting old file: %v", err)
		}

		// Update file metadata
		updates["path"] = newStoragePath
		updates["storage"] = storageConfig.Provider
		updates["bucket"] = storageConfig.Bucket
		updates["endpoint"] = storageConfig.Endpoint
		updates["size"] = len(fileBytes)
		updates["hash"] = newHash

		// Update category if file type changed
		newCategory := structs.GetFileCategory(ext)
		updates["category"] = newCategory

		delete(updates, "file")
	}

	// Process extras updates - merge with existing
	if extrasUpdate, ok := updates["extras"].(types.JSON); ok {
		existingExtras := getExtrasFromFile(existing)

		// Merge extras
		for k, v := range extrasUpdate {
			existingExtras[k] = v
		}

		updates["extras"] = existingExtras
	}

	// Set updated by
	userID := ctxutil.GetUserID(ctx)
	if userID != "" {
		updates["updated_by"] = userID
	}

	// Update file
	row, err := s.fileRepo.Update(ctx, slug, updates)
	if err != nil {
		return nil, handleEntError(ctx, "File", err)
	}

	// Publish event
	if s.publisher != nil {
		eventData := &event.FileEventData{
			ID:      row.ID,
			Name:    row.Name,
			Path:    row.Path,
			Type:    row.Type,
			Size:    row.Size,
			Storage: row.Storage,
			Bucket:  row.Bucket,
			OwnerID: row.OwnerID,
			UserID:  userID,
			Extras:  &row.Extras,
		}
		s.publisher.PublishFileUpdated(ctx, eventData)
	}

	return s.serialize(row), nil
}

// Get retrieves file by ID
func (s *fileService) Get(ctx context.Context, slug string) (*structs.ReadFile, error) {
	if validator.IsEmpty(slug) {
		return nil, errors.New(ecode.FieldIsRequired("slug"))
	}

	row, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New(ecode.NotExist(fmt.Sprintf("File %s", slug)))
		}
		return nil, errors.New("error retrieving file")
	}

	return s.serialize(row), nil
}

// GetPublic retrieves public file
func (s *fileService) GetPublic(ctx context.Context, slug string) (*structs.ReadFile, error) {
	file, err := s.Get(ctx, slug)
	if err != nil {
		return nil, err
	}

	if !file.IsPublic {
		return nil, errors.New("file is not public")
	}

	// Check expiration
	if file.ExpiresAt != nil && time.Now().UnixMilli() > *file.ExpiresAt {
		return nil, errors.New("file access has expired")
	}

	return file, nil
}

// GetByShareToken retrieves file by share token
func (s *fileService) GetByShareToken(ctx context.Context, token string) (*structs.ReadFile, error) {
	if len(token) < 10 {
		return nil, errors.New("invalid share token")
	}

	// Extract file ID from token (simplified)
	fileID := token[:len(token)-10] // Remove timestamp suffix

	file, err := s.Get(ctx, fileID)
	if err != nil {
		return nil, err
	}

	// Verify token validity
	if file.AccessLevel != structs.AccessLevelShared {
		return nil, errors.New("file is not shared")
	}

	return file, nil
}

// Delete deletes file
func (s *fileService) Delete(ctx context.Context, slug string) error {
	if validator.IsEmpty(slug) {
		return errors.New(ecode.FieldIsRequired("slug"))
	}

	storageClient, _ := ctxutil.GetStorage(ctx)

	// Get file details
	row, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		return errors.New("error retrieving file")
	}

	// Get thumbnail path
	thumbnailPath := ""
	extras := getExtrasFromFile(row)
	if tp, ok := extras["thumbnail_path"].(string); ok {
		thumbnailPath = tp
	}

	// Delete from database first
	err = s.fileRepo.Delete(ctx, slug)
	if err != nil {
		return errors.New("error deleting file record")
	}

	// Delete from storage (don't fail if storage deletion fails)
	if storageClient != nil {
		if err := storageClient.Delete(row.Path); err != nil {
			logger.Errorf(ctx, "Error deleting file from storage: %v", err)
		}

		// Delete thumbnail
		if thumbnailPath != "" {
			if err := storageClient.Delete(thumbnailPath); err != nil {
				logger.Warnf(ctx, "Error deleting thumbnail: %v", err)
			}
		}
	}

	// Publish event
	if s.publisher != nil {
		userID := ctxutil.GetUserID(ctx)
		eventData := &event.FileEventData{
			ID:      row.ID,
			Name:    row.Name,
			Path:    row.Path,
			Type:    row.Type,
			Size:    row.Size,
			Storage: row.Storage,
			Bucket:  row.Bucket,
			OwnerID: row.OwnerID,
			UserID:  userID,
		}
		s.publisher.PublishFileDeleted(ctx, eventData)
	}

	return nil
}

// List lists files with pagination
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
		if err != nil {
			return nil, 0, err
		}

		total := s.fileRepo.CountX(ctx, &lp)

		results := make([]*structs.ReadFile, 0, len(rows))
		for _, row := range rows {
			results = append(results, s.serialize(row))
		}

		return results, total, nil
	})
}

// GetFileStream gets file stream
func (s *fileService) GetFileStream(ctx context.Context, slug string) (io.ReadCloser, *structs.ReadFile, error) {
	if validator.IsEmpty(slug) {
		return nil, nil, errors.New(ecode.FieldIsRequired("slug"))
	}

	storageClient, _ := ctxutil.GetStorage(ctx)
	if storageClient == nil {
		return nil, nil, errors.New("storage not configured")
	}

	row, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil, errors.New(ecode.NotExist(fmt.Sprintf("File %s", slug)))
		}
		return nil, nil, errors.New("error retrieving file")
	}

	// Check expiration
	extras := getExtrasFromFile(row)
	if exp, ok := extras["expires_at"].(int64); ok {
		if time.Now().UnixMilli() > exp {
			return nil, nil, errors.New("file access has expired")
		}
	}

	fileStream, err := storageClient.GetStream(row.Path)
	if err != nil {
		logger.Errorf(ctx, "Error retrieving file stream: %v", err)
		return nil, nil, errors.New("error retrieving file stream")
	}

	return fileStream, s.serialize(row), nil
}

// GetFileStreamByID gets file stream by ID
func (s *fileService) GetFileStreamByID(ctx context.Context, id string) (io.ReadCloser, error) {
	storageClient, _ := ctxutil.GetStorage(ctx)
	if storageClient == nil {
		return nil, errors.New("storage not configured")
	}

	row, err := s.fileRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("error retrieving file")
	}

	return storageClient.GetStream(row.Path)
}

// GetThumbnail gets thumbnail stream
func (s *fileService) GetThumbnail(ctx context.Context, slug string) (io.ReadCloser, error) {
	storageClient, _ := ctxutil.GetStorage(ctx)
	if storageClient == nil {
		return nil, errors.New("storage not configured")
	}

	row, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		return nil, errors.New("error retrieving file")
	}

	extras := getExtrasFromFile(row)
	thumbnailPath, ok := extras["thumbnail_path"].(string)
	if !ok || thumbnailPath == "" {
		return nil, errors.New("thumbnail not found")
	}

	return storageClient.GetStream(thumbnailPath)
}

// SearchByTags searches files by tags
func (s *fileService) SearchByTags(ctx context.Context, ownerID string, tags []string, limit int) ([]*structs.ReadFile, error) {
	if len(tags) == 0 {
		return nil, errors.New("at least one tag is required")
	}

	files, err := s.fileRepo.SearchByTags(ctx, ownerID, tags, limit)
	if err != nil {
		return nil, err
	}

	results := make([]*structs.ReadFile, 0, len(files))
	for _, row := range files {
		results = append(results, s.serialize(row))
	}

	return results, nil
}

// GeneratePublicURL generates public URL
func (s *fileService) GeneratePublicURL(ctx context.Context, slug string, expirationHours int) (string, error) {
	row, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		return "", handleEntError(ctx, "File", err)
	}

	if expirationHours <= 0 {
		expirationHours = 24
	}
	expiresAt := time.Now().Add(time.Duration(expirationHours) * time.Hour).UnixMilli()

	extras := getExtrasFromFile(row)
	extras["is_public"] = true
	extras["expires_at"] = expiresAt

	_, err = s.fileRepo.Update(ctx, slug, types.JSON{
		"extras": extras,
	})
	if err != nil {
		return "", handleEntError(ctx, "File", err)
	}

	// Generate share token (simplified)
	shareToken := fmt.Sprintf("%s%d", row.ID, expiresAt)
	downloadURL := fmt.Sprintf("/res/share/%s", shareToken)
	return downloadURL, nil
}

// CreateVersion creates file version
func (s *fileService) CreateVersion(ctx context.Context, slug string, file io.Reader, filename string) (*structs.ReadFile, error) {
	existing, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		return nil, handleEntError(ctx, "File", err)
	}

	storageClient, storageConfig := ctxutil.GetStorage(ctx)
	if storageClient == nil || storageConfig == nil {
		return nil, errors.New("storage not configured")
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if closer, ok := file.(io.Closer); ok {
		closer.Close()
	}

	// Generate version path using existing path prefix if available
	ext := filepath.Ext(filename)
	extras := getExtrasFromFile(existing)

	var versionPath string
	if pathPrefix, hasPrefix := extras["path_prefix"].(string); hasPrefix && pathPrefix != "" {
		versionPath = s.generateVersionPathWithPrefix(existing.OwnerID, slug, filename, pathPrefix)
	} else {
		versionPath = s.generateVersionPath(existing.OwnerID, slug, filename)
	}

	_, err = storageClient.Put(versionPath, bytes.NewReader(fileBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to store file version: %w", err)
	}

	// Extract current extras
	var versions []string
	if v, ok := extras["versions"].([]string); ok {
		versions = v
	}
	versions = append(versions, existing.ID)

	createBody := &structs.CreateFileBody{
		Name:         strings.TrimSuffix(filename, ext),
		OriginalName: filename,
		Path:         versionPath,
		Type:         existing.Type,
		Size:         &[]int{len(fileBytes)}[0],
		Storage:      storageConfig.Provider,
		Bucket:       storageConfig.Bucket,
		Endpoint:     storageConfig.Endpoint,
		OwnerID:      existing.OwnerID,
	}

	// Copy extended properties
	if accessLevel, ok := extras["access_level"].(string); ok {
		createBody.AccessLevel = structs.AccessLevel(accessLevel)
	}
	if tags, ok := extras["tags"].([]string); ok {
		createBody.Tags = tags
	}
	if isPublic, ok := extras["is_public"].(bool); ok {
		createBody.IsPublic = isPublic
	}

	return s.Create(ctx, createBody)
}

// GetVersions gets file versions
func (s *fileService) GetVersions(ctx context.Context, slug string) ([]*structs.ReadFile, error) {
	current, err := s.Get(ctx, slug)
	if err != nil {
		return nil, err
	}

	extras := getExtrasFromFile(&ent.File{Extras: *current.Extras})
	versions, ok := extras["versions"].([]string)
	if !ok || len(versions) == 0 {
		return []*structs.ReadFile{current}, nil
	}

	result := make([]*structs.ReadFile, 0, len(versions)+1)
	result = append(result, current)

	for _, versionID := range versions {
		version, err := s.Get(ctx, versionID)
		if err != nil {
			logger.Warnf(ctx, "Error retrieving version %s: %v", versionID, err)
			continue
		}
		result = append(result, version)
	}

	return result, nil
}

// SetAccessLevel sets file access level
func (s *fileService) SetAccessLevel(ctx context.Context, slug string, accessLevel structs.AccessLevel) (*structs.ReadFile, error) {
	if accessLevel != structs.AccessLevelPublic &&
		accessLevel != structs.AccessLevelPrivate &&
		accessLevel != structs.AccessLevelShared {
		return nil, fmt.Errorf("invalid access level: %s", accessLevel)
	}

	row, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		return nil, handleEntError(ctx, "File", err)
	}

	extras := getExtrasFromFile(row)
	extras["access_level"] = string(accessLevel)
	extras["is_public"] = accessLevel == structs.AccessLevelPublic

	updated, err := s.fileRepo.Update(ctx, slug, types.JSON{
		"extras": extras,
	})
	if err != nil {
		return nil, handleEntError(ctx, "File", err)
	}

	return s.serialize(updated), nil
}

// CreateThumbnail creates thumbnail
func (s *fileService) CreateThumbnail(ctx context.Context, slug string, options *structs.ProcessingOptions) (*structs.ReadFile, error) {
	row, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		return nil, handleEntError(ctx, "File", err)
	}

	if !validator.IsImageFile(row.Path) {
		return nil, fmt.Errorf("file is not an image")
	}

	if options == nil {
		options = &structs.ProcessingOptions{
			CreateThumbnail: true,
			MaxWidth:        300,
			MaxHeight:       300,
		}
	}

	storageClient, _ := ctxutil.GetStorage(ctx)
	if storageClient == nil {
		return nil, errors.New("storage not configured")
	}

	file, err := storageClient.GetStream(row.Path)
	if err != nil {
		return nil, fmt.Errorf("error retrieving file: %w", err)
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

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

	thumbnailPath := s.generateThumbnailPath(row.Path)
	_, err = storageClient.Put(thumbnailPath, bytes.NewReader(thumbnailBytes))
	if err != nil {
		return nil, fmt.Errorf("error storing thumbnail: %w", err)
	}

	extras := getExtrasFromFile(row)
	extras["thumbnail_path"] = thumbnailPath

	updated, err := s.fileRepo.Update(ctx, slug, types.JSON{
		"extras": extras,
	})
	if err != nil {
		return nil, handleEntError(ctx, "File", err)
	}

	return s.serialize(updated), nil
}

// GetTagsByOwner gets tags by owner
func (s *fileService) GetTagsByOwner(ctx context.Context, ownerID string) ([]string, error) {
	return s.fileRepo.GetTagsByOwner(ctx, ownerID)
}

// Helper methods

// generateUniqueStoragePathWithPrefix generates storage path with custom prefix
func (s *fileService) generateUniqueStoragePathWithPrefix(ownerID, fileName, ext, prefix string) string {
	return s.generateUniqueStoragePath(fileName, ext, &ownerID, &prefix)
}

// generateUniqueStoragePath generates default unique storage path
func (s *fileService) generateUniqueStoragePath(fileName, ext string, ownerID, pathPrefix *string) string {
	timestamp := time.Now().Unix()
	randomID := nanoid.Number(8)

	// Clean filename
	cleanName := strings.ReplaceAll(fileName, " ", "_")
	cleanName = strings.ReplaceAll(cleanName, "/", "_")

	pathParts := []string{}

	// Add pathPrefix if provided
	if pathPrefix != nil && *pathPrefix != "" {
		pathParts = append(pathParts, *pathPrefix)
	}

	// Add ownerID if provided
	if ownerID != nil && *ownerID != "" {
		pathParts = append(pathParts, *ownerID)
	}

	// Add filename with timestamp and random ID
	filename := fmt.Sprintf("%d_%s_%s%s", timestamp, randomID, cleanName, ext)
	pathParts = append(pathParts, filename)

	return strings.Join(pathParts, "/")
}

// generateUniqueName generates unique name for database
func (s *fileService) generateUniqueName(originalName string) string {
	timestamp := time.Now().Unix()
	randomID := nanoid.Number(6)
	return fmt.Sprintf("%s_%d_%s", originalName, timestamp, randomID)
}

// generateThumbnailPath generates thumbnail storage path
func (s *fileService) generateThumbnailPath(originalPath string) string {
	dir := filepath.Dir(originalPath)
	fileName := filepath.Base(originalPath)
	ext := filepath.Ext(fileName)
	nameWithoutExt := strings.TrimSuffix(fileName, ext)

	return fmt.Sprintf("%s/thumbnails/%s_thumb.jpg", dir, nameWithoutExt)
}

// generateVersionPath generates version storage path
func (s *fileService) generateVersionPath(ownerID, parentSlug, fileName string) string {
	timestamp := time.Now().Unix()
	randomID := nanoid.Number(6)
	ext := filepath.Ext(fileName)
	nameWithoutExt := strings.TrimSuffix(fileName, ext)

	return fmt.Sprintf("files/%s/versions/%s/%d_%s_%s%s",
		ownerID, parentSlug, timestamp, randomID, nameWithoutExt, ext)
}

// generateVersionPathWithPrefix generates version storage path with custom prefix
func (s *fileService) generateVersionPathWithPrefix(ownerID, parentSlug, fileName, prefix string) string {
	timestamp := time.Now().Unix()
	randomID := nanoid.Number(6)
	ext := filepath.Ext(fileName)
	nameWithoutExt := strings.TrimSuffix(fileName, ext)

	return fmt.Sprintf("%s/%s/versions/%s/%d_%s_%s%s",
		prefix, ownerID, parentSlug, timestamp, randomID, nameWithoutExt, ext)
}

// getExtensionFromMimeType attempts to get file extension from MIME type
func (s *fileService) getExtensionFromMimeType(mimeType string) string {
	mimeToExt := map[string]string{
		"image/jpeg":       ".jpg",
		"image/png":        ".png",
		"image/gif":        ".gif",
		"image/webp":       ".webp",
		"text/plain":       ".txt",
		"application/pdf":  ".pdf",
		"application/json": ".json",
		"video/mp4":        ".mp4",
		"video/avi":        ".avi",
		"audio/mp3":        ".mp3",
		"audio/wav":        ".wav",
	}

	if ext, ok := mimeToExt[mimeType]; ok {
		return ext
	}
	return ""
}

// serialize converts ent.File to structs.ReadFile
func (s *fileService) serialize(row *ent.File) *structs.ReadFile {
	extras := getExtrasFromFile(row)

	file := &structs.ReadFile{
		ID:           row.ID,
		Name:         row.Name,
		OriginalName: row.OriginalName,
		Path:         row.Path,
		Type:         row.Type,
		Size:         &row.Size,
		Storage:      row.Storage,
		Bucket:       row.Bucket,
		Endpoint:     row.Endpoint,
		AccessLevel:  structs.AccessLevel(row.AccessLevel),
		ExpiresAt:    row.ExpiresAt,
		Tags:         row.Tags,
		IsPublic:     row.IsPublic,
		Category:     structs.FileCategory(row.Category),
		Hash:         row.Hash,
		OwnerID:      row.OwnerID,
		Extras:       &row.Extras,
		CreatedBy:    &row.CreatedBy,
		CreatedAt:    &row.CreatedAt,
		UpdatedBy:    &row.UpdatedBy,
		UpdatedAt:    &row.UpdatedAt,
	}

	// Check expiration
	if file.ExpiresAt != nil {
		file.IsExpired = time.Now().UnixMilli() > *file.ExpiresAt
	}

	// Set URLs if public
	if file.IsPublic {
		file.DownloadURL = fmt.Sprintf("/res/dl/%s", row.ID)
		if thumbnailPath, ok := extras["thumbnail_path"].(string); ok && thumbnailPath != "" {
			file.ThumbnailURL = fmt.Sprintf("/res/thumb/%s", row.ID)
		}
	}

	return file
}
