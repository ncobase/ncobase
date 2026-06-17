package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"ncobase/plugin/resource/data"
	"ncobase/plugin/resource/data/repository"
	"ncobase/plugin/resource/event"
	"ncobase/plugin/resource/structs"
	"path"
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
	GenerateShareURL(ctx context.Context, slug string, accessLevel structs.AccessLevel, expirationHours int) (string, int64, error)
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
	configProvider ResourceConfigProvider
}

type bufferedMultipartFile struct {
	*bytes.Reader
}

func (f bufferedMultipartFile) Close() error {
	return nil
}

func NewFileService(
	d *data.Data,
	imageProcessor ImageProcessorInterface,
	quotaService QuotaServiceInterface,
	publisher event.PublisherInterface,
	configProvider ResourceConfigProvider,
) FileServiceInterface {
	if configProvider == nil {
		configProvider = NewDefaultConfigProvider()
	}
	return &fileService{
		fileRepo:       repository.NewFileRepository(d),
		imageProcessor: imageProcessor,
		quotaService:   quotaService,
		publisher:      publisher,
		configProvider: configProvider,
	}
}

func (s *fileService) findFileByHash(ctx context.Context, ownerID, hash string) (*structs.ReadFile, error) {
	file, err := s.fileRepo.GetByHash(ctx, ownerID, hash)
	if repository.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return repository.SerializeFile(file), nil
}

func (s *fileService) validateUpload(ctx context.Context, filename, contentType string, size int64) error {
	if s.configProvider == nil {
		return nil
	}
	return s.configProvider.ValidateUpload(ctx, filename, contentType, size)
}

func fileBodyFilename(body *structs.CreateFileBody) string {
	if body.OriginalName != "" {
		return body.OriginalName
	}
	if body.Path != "" {
		return body.Path
	}
	return body.Name
}

func (s *fileService) storagePolicy(ctx context.Context) *StoragePolicyConfig {
	if s.configProvider != nil {
		return s.configProvider.StoragePolicy(ctx)
	}
	return NewDefaultConfigProvider().StoragePolicy(ctx)
}

func fileSpaceID(ctx context.Context, extras *types.JSON) string {
	if extras != nil {
		if value, ok := (*extras)["space_id"].(string); ok && strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ctxutil.GetSpaceID(ctx)
}

func (s *fileService) validateSharePolicy(ctx context.Context, file *structs.ReadFile, accessLevel structs.AccessLevel) error {
	if accessLevel != structs.AccessLevelPublic && accessLevel != structs.AccessLevelShared {
		return nil
	}

	policy := s.storagePolicy(ctx)
	if policy == nil {
		return nil
	}
	if !policy.AllowPublicLinks {
		return errors.New("public file links are disabled")
	}
	if policy.RequireOwnerScope && file != nil && file.OwnerID == "" {
		return errors.New("owner scope is required for public file links")
	}
	return nil
}

func (s *fileService) publishFileAccessed(ctx context.Context, file *structs.ReadFile, accessType string) {
	if s.publisher == nil || file == nil {
		return
	}
	policy := s.storagePolicy(ctx)
	if policy != nil && !policy.AuditDownloads {
		return
	}

	extras := repository.CloneExtrasPtr(file.Extras)
	if accessType != "" {
		extras["access_type"] = accessType
	}

	size := 0
	if file.Size != nil {
		size = *file.Size
	}

	s.publisher.PublishFileAccessed(ctx, &event.FileEventData{
		ID:      file.ID,
		Name:    file.Name,
		Path:    file.Path,
		Type:    file.Type,
		Size:    size,
		Storage: file.Storage,
		Bucket:  file.Bucket,
		OwnerID: file.OwnerID,
		SpaceID: fileSpaceID(ctx, file.Extras),
		UserID:  ctxutil.GetUserID(ctx),
		Extras:  &extras,
	})
}

// Create creates a new file
func (s *fileService) Create(ctx context.Context, body *structs.CreateFileBody) (*structs.ReadFile, error) {
	// Get ownerID from context if not provided
	if body.OwnerID == "" {
		if userID := ctxutil.GetUserID(ctx); userID != "" {
			body.OwnerID = userID
		}
	}

	if body.Size != nil {
		if err := s.validateUpload(ctx, fileBodyFilename(body), body.Type, int64(*body.Size)); err != nil {
			return nil, err
		}
	} else if err := s.validateUpload(ctx, fileBodyFilename(body), body.Type, 0); err != nil {
		return nil, err
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

	actualSize := len(fileBytes)
	if err := s.validateUpload(ctx, fileBodyFilename(body), body.Type, int64(actualSize)); err != nil {
		return nil, err
	}
	if body.Size == nil || *body.Size != actualSize {
		body.Size = &actualSize
	}

	quotaReserved := false
	storagePath := ""
	thumbnailPath := ""
	created := false
	defer func() {
		if created {
			return
		}
		if thumbnailPath != "" {
			if deleteErr := storageClient.Delete(thumbnailPath); deleteErr != nil {
				logger.Warnf(ctx, "Failed to cleanup thumbnail after create failure: %v", deleteErr)
			}
		}
		if storagePath != "" {
			if deleteErr := storageClient.Delete(storagePath); deleteErr != nil {
				logger.Errorf(ctx, "Failed to cleanup file after create failure: %v", deleteErr)
			}
		}
		if quotaReserved && body.OwnerID != "" && s.quotaService != nil {
			if quotaErr := s.quotaService.UpdateUsage(ctx, body.OwnerID, "storage", -int64(actualSize)); quotaErr != nil {
				logger.Warnf(ctx, "Failed to compensate quota after create failure: %v", quotaErr)
			}
		}
	}()

	// Check quota only after the actual stream size is known.
	if body.OwnerID != "" && s.quotaService != nil {
		canProceed, err := s.quotaService.CheckAndUpdateQuota(ctx, body.OwnerID, actualSize)
		if err != nil {
			logger.Warnf(ctx, "Error checking quota: %v", err)
		} else if !canProceed {
			return nil, errors.New("storage quota exceeded")
		} else {
			quotaReserved = true
		}
	}

	// Calculate file hash for deduplication
	hash := calculateFileHash(fileBytes)

	// Check for existing file with same hash (optional deduplication).
	if body.OwnerID != "" && hash != "" {
		existing, err := s.findFileByHash(ctx, body.OwnerID, hash)
		if err == nil && existing != nil {
			logger.Infof(ctx, "File with same hash already exists: %s", existing.ID)
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

	storagePath = s.generateUniqueStoragePath(body.Name, ext, ownerIDPtr, pathPrefixPtr)

	// Store file
	_, storeErr := storageClient.Put(storagePath, bytes.NewReader(fileBytes))
	if storeErr != nil {
		logger.Errorf(ctx, "Error storing file to %s: %v", storageConfig.Provider, storeErr)
		return nil, fmt.Errorf("failed to store file: %w", storeErr)
	}

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
	category := structs.GetFileCategory(filepath.Ext(storagePath))

	if category == structs.FileCategoryImage && s.imageProcessor != nil {
		if s.configProvider != nil {
			body.ProcessingOptions = s.configProvider.NormalizeProcessingOptions(ctx, body.ProcessingOptions)
		}

		if body.ProcessingOptions != nil && body.ProcessingOptions.CreateThumbnail {
			thumbnailBytes, err := s.imageProcessor.CreateThumbnail(
				ctx,
				bytes.NewReader(fileBytes),
				body.Name,
				body.ProcessingOptions.MaxWidth,
				body.ProcessingOptions.MaxHeight,
				body.ProcessingOptions.CompressionQuality,
			)

			if err != nil {
				logger.Warnf(ctx, "Error creating thumbnail: %v", err)
			} else {
				thumbnailPath = s.generateThumbnailPath(storagePath)
				_, err = storageClient.Put(thumbnailPath, bytes.NewReader(thumbnailBytes))
				if err != nil {
					logger.Warnf(ctx, "Error storing thumbnail: %v", err)
					thumbnailPath = ""
				}
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
	if spaceID := ctxutil.GetSpaceID(ctx); spaceID != "" {
		extendedData["space_id"] = spaceID
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
				SpaceID: fileSpaceID(ctx, &row.Extras),
				UserID:  eventUserID,
				Extras:  &row.Extras,
			}
			s.publisher.PublishFileCreated(ctx, eventData)
		}

		created = true
		return repository.SerializeFile(row), nil
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

	var newStoredPath string
	var oldStoredPath string
	var quotaReservedDelta int64
	var quotaReleaseDelta int64

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

		uploadFileName := existing.OriginalName
		if name, ok := updates["original_name"].(string); ok && name != "" {
			uploadFileName = name
		} else if name, ok := updates["name"].(string); ok && name != "" {
			uploadFileName = name
		}
		contentType := existing.Type
		if value, ok := updates["type"].(string); ok && value != "" {
			contentType = value
		}
		if err := s.validateUpload(ctx, uploadFileName, contentType, int64(len(fileBytes))); err != nil {
			return nil, err
		}

		sizeDelta := int64(len(fileBytes) - existing.Size)
		if sizeDelta > 0 && existing.OwnerID != "" && s.quotaService != nil {
			canProceed, quotaErr := s.quotaService.CheckAndUpdateQuota(ctx, existing.OwnerID, int(sizeDelta))
			if quotaErr != nil {
				logger.Warnf(ctx, "Error checking quota for file update: %v", quotaErr)
			} else if !canProceed {
				return nil, errors.New("storage quota exceeded")
			} else {
				quotaReservedDelta = sizeDelta
			}
		} else if sizeDelta < 0 {
			quotaReleaseDelta = sizeDelta
		}

		// Calculate new hash
		newHash := calculateFileHash(fileBytes)

		// Generate new storage path
		fileName := existing.Name
		if name, ok := updates["name"].(string); ok {
			fileName = name
		}

		ext := filepath.Ext(existing.Path)
		extras := repository.CloneExtras(existing.Extras)

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
			if quotaReservedDelta > 0 && existing.OwnerID != "" && s.quotaService != nil {
				_ = s.quotaService.UpdateUsage(ctx, existing.OwnerID, "storage", -quotaReservedDelta)
			}
			return nil, errors.New("error updating file")
		}

		newStoredPath = newStoragePath
		oldStoredPath = existing.Path

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
		existingExtras := repository.CloneExtras(existing.Extras)

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
		if newStoredPath != "" {
			if storageClient, _ := ctxutil.GetStorage(ctx); storageClient != nil {
				if deleteErr := storageClient.Delete(newStoredPath); deleteErr != nil {
					logger.Warnf(ctx, "Error deleting new file after update failure: %v", deleteErr)
				}
			}
		}
		if quotaReservedDelta > 0 && existing.OwnerID != "" && s.quotaService != nil {
			_ = s.quotaService.UpdateUsage(ctx, existing.OwnerID, "storage", -quotaReservedDelta)
		}
		return nil, handleEntError(ctx, "File", err)
	}

	if newStoredPath != "" {
		if storageClient, _ := ctxutil.GetStorage(ctx); storageClient != nil && oldStoredPath != "" {
			if err := storageClient.Delete(oldStoredPath); err != nil {
				logger.Warnf(ctx, "Error deleting old file: %v", err)
			}
		}
		if quotaReleaseDelta < 0 && existing.OwnerID != "" && s.quotaService != nil {
			if _, quotaErr := s.quotaService.RefreshUsage(ctx, existing.OwnerID); quotaErr != nil {
				logger.Warnf(ctx, "Failed to release quota after file update: %v", quotaErr)
			}
			if spaceID := fileSpaceID(ctx, &row.Extras); spaceID != "" {
				if _, quotaErr := s.quotaService.RefreshSpaceUsage(ctx, spaceID); quotaErr != nil {
					logger.Warnf(ctx, "Failed to refresh space quota after file update: %v", quotaErr)
				}
			}
		}
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
			SpaceID: fileSpaceID(ctx, &row.Extras),
			UserID:  userID,
			Extras:  &row.Extras,
		}
		s.publisher.PublishFileUpdated(ctx, eventData)
	}

	return repository.SerializeFile(row), nil
}

// Get retrieves file by ID
func (s *fileService) Get(ctx context.Context, slug string) (*structs.ReadFile, error) {
	if validator.IsEmpty(slug) {
		return nil, errors.New(ecode.FieldIsRequired("slug"))
	}

	row, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, errors.New(ecode.NotExist(fmt.Sprintf("File %s", slug)))
		}
		return nil, errors.New("error retrieving file")
	}

	return repository.SerializeFile(row), nil
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

	if file.ExpiresAt != nil && time.Now().UnixMilli() > *file.ExpiresAt {
		return nil, errors.New("file access has expired")
	}
	extras := repository.CloneExtrasPtr(file.Extras)
	if shareExpiresAt, ok := jsonInt64(extras["share_expires_at"]); ok && time.Now().UnixMilli() > shareExpiresAt {
		return nil, errors.New("public link has expired")
	}

	return file, nil
}

// GetByShareToken retrieves file by share token
func (s *fileService) GetByShareToken(ctx context.Context, token string) (*structs.ReadFile, error) {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, errors.New("invalid share token")
	}

	file, err := s.Get(ctx, parts[0])
	if err != nil {
		return nil, err
	}

	if file.AccessLevel != structs.AccessLevelShared {
		return nil, errors.New("file is not shared")
	}

	extras := repository.CloneExtrasPtr(file.Extras)
	storedToken, _ := extras["share_token"].(string)
	if storedToken == "" || storedToken != token {
		return nil, errors.New("invalid share token")
	}
	if shareExpiresAt, ok := jsonInt64(extras["share_expires_at"]); ok && time.Now().UnixMilli() > shareExpiresAt {
		return nil, errors.New("share token has expired")
	}

	return file, nil
}

func jsonInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case float64:
		return int64(v), true
	case json.Number:
		n, err := v.Int64()
		return n, err == nil
	default:
		return 0, false
	}
}

func stringListFromAny(value any) []string {
	switch v := value.(type) {
	case []string:
		return append([]string{}, v...)
	case []any:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				result = append(result, s)
			}
		}
		return result
	default:
		return []string{}
	}
}

func appendUniqueString(values []string, value string) []string {
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
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
	extras := repository.CloneExtras(row.Extras)
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

	spaceID := fileSpaceID(ctx, &row.Extras)
	if s.quotaService != nil {
		if row.OwnerID != "" {
			if _, quotaErr := s.quotaService.RefreshUsage(ctx, row.OwnerID); quotaErr != nil {
				logger.Warnf(ctx, "Failed to refresh owner quota after file deletion: %v", quotaErr)
			}
		}
		if spaceID != "" {
			if _, quotaErr := s.quotaService.RefreshSpaceUsage(ctx, spaceID); quotaErr != nil {
				logger.Warnf(ctx, "Failed to refresh space quota after file deletion: %v", quotaErr)
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
			SpaceID: spaceID,
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
			results = append(results, repository.SerializeFile(row))
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
		if repository.IsNotFound(err) {
			return nil, nil, errors.New(ecode.NotExist(fmt.Sprintf("File %s", slug)))
		}
		return nil, nil, errors.New("error retrieving file")
	}

	if row.ExpiresAt != nil && time.Now().UnixMilli() > *row.ExpiresAt {
		return nil, nil, errors.New("file access has expired")
	}

	fileStream, err := storageClient.GetStream(row.Path)
	if err != nil {
		logger.Errorf(ctx, "Error retrieving file stream: %v", err)
		return nil, nil, errors.New("error retrieving file stream")
	}

	readFile := repository.SerializeFile(row)
	s.publishFileAccessed(ctx, readFile, "stream")
	return fileStream, readFile, nil
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

	stream, err := storageClient.GetStream(row.Path)
	if err != nil {
		return nil, err
	}
	s.publishFileAccessed(ctx, repository.SerializeFile(row), "stream_by_id")
	return stream, nil
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

	extras := repository.CloneExtras(row.Extras)
	thumbnailPath, ok := extras["thumbnail_path"].(string)
	if !ok || thumbnailPath == "" {
		return nil, errors.New("thumbnail not found")
	}

	stream, err := storageClient.GetStream(thumbnailPath)
	if err != nil {
		return nil, err
	}
	s.publishFileAccessed(ctx, repository.SerializeFile(row), "thumbnail")
	return stream, nil
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
		results = append(results, repository.SerializeFile(row))
	}

	return results, nil
}

// GenerateShareURL configures a file for public or token-based shared access.
func (s *fileService) GenerateShareURL(ctx context.Context, slug string, accessLevel structs.AccessLevel, expirationHours int) (string, int64, error) {
	if accessLevel == "" {
		accessLevel = structs.AccessLevelShared
	}
	if accessLevel != structs.AccessLevelPublic && accessLevel != structs.AccessLevelShared {
		return "", 0, fmt.Errorf("invalid share access level: %s", accessLevel)
	}

	row, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		return "", 0, handleEntError(ctx, "File", err)
	}
	if err := s.validateSharePolicy(ctx, repository.SerializeFile(row), accessLevel); err != nil {
		return "", 0, err
	}

	if expirationHours <= 0 {
		expirationHours = 24
	}
	expiresAt := time.Now().Add(time.Duration(expirationHours) * time.Hour).UnixMilli()

	extras := repository.CloneExtras(row.Extras)
	extras["share_access_level"] = string(accessLevel)
	extras["share_expires_at"] = expiresAt

	isPublic := accessLevel == structs.AccessLevelPublic
	shareURL := fmt.Sprintf("/res/dl/%s", row.ID)
	if !isPublic {
		token := fmt.Sprintf("%s.%s", row.ID, nanoid.String(32))
		extras["share_token"] = token
		shareURL = fmt.Sprintf("/res/share/%s", token)
	} else {
		delete(extras, "share_token")
	}

	_, err = s.fileRepo.Update(ctx, slug, types.JSON{
		"access_level": accessLevel,
		"is_public":    isPublic,
		"extras":       extras,
	})
	if err != nil {
		return "", 0, handleEntError(ctx, "File", err)
	}

	return shareURL, expiresAt, nil
}

// CreateVersion creates file version
func (s *fileService) CreateVersion(ctx context.Context, slug string, file io.Reader, filename string) (*structs.ReadFile, error) {
	existing, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		return nil, handleEntError(ctx, "File", err)
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if closer, ok := file.(io.Closer); ok {
		closer.Close()
	}

	ext := filepath.Ext(filename)
	contentType := existing.Type
	if detected := mime.TypeByExtension(ext); detected != "" {
		contentType = detected
	}

	if err := s.validateUpload(ctx, filename, contentType, int64(len(fileBytes))); err != nil {
		return nil, err
	}

	extras := repository.CloneExtras(existing.Extras)
	versions := stringListFromAny(extras["versions"])

	pathPrefix := path.Join("versions", slug)
	if existingPrefix, hasPrefix := extras["path_prefix"].(string); hasPrefix && existingPrefix != "" {
		pathPrefix = path.Join(existingPrefix, "versions", slug)
	}

	versionExtras := types.JSON{
		"version_of":       existing.ID,
		"version_number":   len(versions) + 1,
		"source_path":      existing.Path,
		"source_file_name": existing.OriginalName,
	}

	size := len(fileBytes)
	createBody := &structs.CreateFileBody{
		File:         bufferedMultipartFile{bytes.NewReader(fileBytes)},
		Name:         strings.TrimSuffix(filename, ext),
		OriginalName: filename,
		Path:         filename,
		PathPrefix:   pathPrefix,
		Type:         contentType,
		Size:         &size,
		OwnerID:      existing.OwnerID,
		AccessLevel:  structs.AccessLevel(existing.AccessLevel),
		ExpiresAt:    existing.ExpiresAt,
		Tags:         existing.Tags,
		IsPublic:     existing.IsPublic,
		Extras:       &versionExtras,
	}

	version, err := s.Create(ctx, createBody)
	if err != nil {
		return nil, err
	}

	versions = appendUniqueString(versions, version.ID)
	extras["versions"] = versions
	if _, err := s.fileRepo.Update(ctx, slug, types.JSON{"extras": extras}); err != nil {
		logger.Errorf(ctx, "Error updating version metadata for file %s: %v", slug, err)
		_ = s.Delete(ctx, version.ID)
		return nil, handleEntError(ctx, "File", err)
	}

	return version, nil
}

// GetVersions gets file versions
func (s *fileService) GetVersions(ctx context.Context, slug string) ([]*structs.ReadFile, error) {
	current, err := s.Get(ctx, slug)
	if err != nil {
		return nil, err
	}

	extras := repository.CloneExtrasPtr(current.Extras)
	versions := stringListFromAny(extras["versions"])
	if len(versions) == 0 {
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
	if err := s.validateSharePolicy(ctx, repository.SerializeFile(row), accessLevel); err != nil {
		return nil, err
	}

	extras := repository.CloneExtras(row.Extras)
	isPublic := accessLevel == structs.AccessLevelPublic
	extras["access_level"] = string(accessLevel)
	extras["is_public"] = isPublic

	updated, err := s.fileRepo.Update(ctx, slug, types.JSON{
		"access_level": accessLevel,
		"is_public":    isPublic,
		"extras":       extras,
	})
	if err != nil {
		return nil, handleEntError(ctx, "File", err)
	}

	return repository.SerializeFile(updated), nil
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

	if s.configProvider != nil {
		options = s.configProvider.NormalizeProcessingOptions(ctx, options)
	}
	if options == nil || !options.CreateThumbnail {
		return nil, fmt.Errorf("thumbnail creation is disabled")
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
		options.CompressionQuality,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating thumbnail: %w", err)
	}

	thumbnailPath := s.generateThumbnailPath(row.Path)
	_, err = storageClient.Put(thumbnailPath, bytes.NewReader(thumbnailBytes))
	if err != nil {
		return nil, fmt.Errorf("error storing thumbnail: %w", err)
	}

	extras := repository.CloneExtras(row.Extras)
	extras["thumbnail_path"] = thumbnailPath

	updated, err := s.fileRepo.Update(ctx, slug, types.JSON{
		"extras": extras,
	})
	if err != nil {
		return nil, handleEntError(ctx, "File", err)
	}

	return repository.SerializeFile(updated), nil
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
