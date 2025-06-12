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
	"time"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/data/storage"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation/validator"
)

// FileServiceInterface defines file service methods
type FileServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateFileBody) (*structs.ReadFile, error)
	Update(ctx context.Context, slug string, updates map[string]any) (*structs.ReadFile, error)
	Get(ctx context.Context, slug string) (*structs.ReadFile, error)
	Delete(ctx context.Context, slug string) error
	List(ctx context.Context, params *structs.ListFileParams) (paging.Result[*structs.ReadFile], error)
	GetFileStream(ctx context.Context, slug string) (io.ReadCloser, *structs.ReadFile, error)
	SearchByTags(ctx context.Context, spaceID string, tags []string, limit int) ([]*structs.ReadFile, error)
	GeneratePublicURL(ctx context.Context, slug string, expirationHours int) (string, error)
	CreateVersion(ctx context.Context, slug string, file io.Reader, filename string) (*structs.ReadFile, error)
	GetVersions(ctx context.Context, slug string) ([]*structs.ReadFile, error)
	SetAccessLevel(ctx context.Context, slug string, accessLevel structs.AccessLevel) (*structs.ReadFile, error)
	CreateThumbnail(ctx context.Context, slug string, options *structs.ProcessingOptions) (*structs.ReadFile, error)
}

type fileService struct {
	fileRepo       repository.FileRepositoryInterface
	imageProcessor ImageProcessorInterface
	quotaService   QuotaServiceInterface
	publisher      event.PublisherInterface
}

// NewFileService creates new file service
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
	if validator.IsEmpty(body.OwnerID) {
		return nil, errors.New(ecode.FieldIsRequired("owner_id"))
	}

	if validator.IsEmpty(body.SpaceID) {
		body.SpaceID = ctxutil.GetSpaceID(ctx)
		if body.SpaceID == "" {
			return nil, errors.New(ecode.FieldIsRequired("space_id"))
		}
	}

	// Check quota
	if s.quotaService != nil && body.Size != nil {
		canProceed, err := s.quotaService.CheckAndUpdateQuota(ctx, body.SpaceID, *body.Size)
		if err != nil {
			logger.Warnf(ctx, "Error checking quota: %v", err)
		} else if !canProceed {
			return nil, errors.New("storage quota exceeded")
		}
	}

	// Get storage
	storage, storageConfig := ctxutil.GetStorage(ctx)

	// Read file content
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
		body.File = &readCloser{bytes.NewReader(fileBytes)}
	} else {
		return nil, errors.New("file content is required")
	}

	// Store file
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

	// Set defaults
	if body.AccessLevel == "" {
		body.AccessLevel = structs.AccessLevelPrivate
	}

	// Set storage info
	body.Storage = storageConfig.Provider
	body.Bucket = storageConfig.Bucket
	body.Endpoint = storageConfig.Endpoint

	// Set created by
	userID := ctxutil.GetUserID(ctx)
	body.CreatedBy = &userID

	// Process image if needed
	thumbnailPath := ""
	category := structs.GetFileCategory(filepath.Ext(body.Path))

	if category == structs.FileCategoryImage && s.imageProcessor != nil {
		if body.ProcessingOptions == nil {
			body.ProcessingOptions = &structs.ProcessingOptions{
				CreateThumbnail: true,
				MaxWidth:        300,
				MaxHeight:       300,
			}
		}

		// Create thumbnail
		thumbnailBytes, err := s.imageProcessor.CreateThumbnail(
			ctx,
			bytes.NewReader(fileBytes),
			body.Name,
			body.ProcessingOptions.MaxWidth,
			body.ProcessingOptions.MaxHeight,
		)

		if err == nil {
			thumbnailPath = fmt.Sprintf("thumbnails/%s", body.Path)
			_, err = storage.Put(thumbnailPath, bytes.NewReader(thumbnailBytes))
			if err != nil {
				logger.Warnf(ctx, "Error storing thumbnail: %v", err)
				thumbnailPath = ""
			}
		}
	}

	// Prepare extras
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
	if body.IsPublic {
		extendedData["is_public"] = body.IsPublic
	}
	if thumbnailPath != "" {
		extendedData["thumbnail_path"] = thumbnailPath
	}
	if body.ExpiresAt != nil {
		extendedData["expires_at"] = *body.ExpiresAt
	}
	extendedData["category"] = string(category)

	// Merge with existing extras
	if body.Extras != nil {
		for k, v := range *body.Extras {
			extendedData[k] = v
		}
	}
	body.Extras = &extendedData

	// Create file record
	row, err := s.fileRepo.Create(ctx, body)
	if err != nil {
		logger.Errorf(ctx, "Error creating file: %v", err)
		if thumbnailPath != "" {
			_ = storage.Delete(thumbnailPath)
		}
		return nil, errors.New("failed to create file")
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
			SpaceID: row.SpaceID,
			UserID:  userID,
			Extras:  &row.Extras,
		}
		s.publisher.PublishFileCreated(ctx, eventData)
	}

	return s.serialize(row), nil
}

// Update updates an existing file
func (s *fileService) Update(ctx context.Context, slug string, updates map[string]any) (*structs.ReadFile, error) {
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

	// Handle file update
	if fileReader, ok := updates["file"].(io.Reader); ok {
		storage, storageConfig := ctxutil.GetStorage(ctx)

		fileBytes, err := io.ReadAll(fileReader)
		if err != nil {
			return nil, errors.New("error reading uploaded file")
		}

		if closer, ok := fileReader.(io.Closer); ok {
			closer.Close()
		}

		// Store new file
		if path, ok := updates["path"].(string); ok {
			if _, err := storage.Put(path, bytes.NewReader(fileBytes)); err != nil {
				return nil, errors.New("error updating file")
			}

			updates["storage"] = storageConfig.Provider
			updates["bucket"] = storageConfig.Bucket
			updates["endpoint"] = storageConfig.Endpoint
			updates["size"] = len(fileBytes)
		}

		delete(updates, "file")
	}

	// Update extras
	extras := getExtrasFromFile(existing)
	for field, value := range updates {
		switch field {
		case "folder_path":
			extras["folder_path"] = value
			delete(updates, field)
		case "access_level":
			extras["access_level"] = value
			delete(updates, field)
		case "tags":
			extras["tags"] = value
			delete(updates, field)
		case "is_public":
			extras["is_public"] = value
			delete(updates, field)
		case "expires_at":
			extras["expires_at"] = value
			delete(updates, field)
		}
	}
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
			ID:      row.ID,
			Name:    row.Name,
			Path:    row.Path,
			Type:    row.Type,
			Size:    row.Size,
			Storage: row.Storage,
			Bucket:  row.Bucket,
			OwnerID: row.OwnerID,
			SpaceID: row.SpaceID,
			UserID:  userID,
			Extras:  &row.Extras,
		}
		s.publisher.PublishFileUpdated(ctx, eventData)
	}

	return s.serialize(row), nil
}

// Get retrieves a file by ID
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

// Delete deletes a file by ID
func (s *fileService) Delete(ctx context.Context, slug string) error {
	if validator.IsEmpty(slug) {
		return errors.New(ecode.FieldIsRequired("slug"))
	}

	storage, _ := ctxutil.GetStorage(ctx)

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

	// Delete from database
	err = s.fileRepo.Delete(ctx, slug)
	if err != nil {
		return errors.New("error deleting file")
	}

	// Delete from storage
	if err := storage.Delete(row.Path); err != nil {
		logger.Errorf(ctx, "Error deleting file: %v", err)
	}

	// Delete thumbnail
	if thumbnailPath != "" {
		if err := storage.Delete(thumbnailPath); err != nil {
			logger.Warnf(ctx, "Error deleting thumbnail: %v", err)
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
			SpaceID: row.SpaceID,
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

// GetFileStream retrieves file stream
func (s *fileService) GetFileStream(ctx context.Context, slug string) (io.ReadCloser, *structs.ReadFile, error) {
	if validator.IsEmpty(slug) {
		return nil, nil, errors.New(ecode.FieldIsRequired("slug"))
	}

	storage, _ := ctxutil.GetStorage(ctx)

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
		if time.Now().Unix() > exp {
			return nil, nil, errors.New("file access has expired")
		}
	}

	fileStream, err := storage.GetStream(row.Path)
	if err != nil {
		return nil, nil, errors.New("error retrieving file stream")
	}

	return fileStream, s.serialize(row), nil
}

// SearchByTags searches files by tags
func (s *fileService) SearchByTags(ctx context.Context, spaceID string, tags []string, limit int) ([]*structs.ReadFile, error) {
	if len(tags) == 0 {
		return nil, errors.New("at least one tag is required")
	}

	files, err := s.fileRepo.SearchByTags(ctx, spaceID, tags, limit)
	if err != nil {
		return nil, err
	}

	results := make([]*structs.ReadFile, 0, len(files))
	for _, row := range files {
		results = append(results, s.serialize(row))
	}

	return results, nil
}

// GeneratePublicURL generates a temporary public URL
func (s *fileService) GeneratePublicURL(ctx context.Context, slug string, expirationHours int) (string, error) {
	row, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		return "", handleEntError(ctx, "File", err)
	}

	if expirationHours <= 0 {
		expirationHours = 24
	}
	expiresAt := time.Now().Add(time.Duration(expirationHours) * time.Hour).Unix()

	extras := getExtrasFromFile(row)
	extras["is_public"] = true
	extras["expires_at"] = expiresAt

	_, err = s.fileRepo.Update(ctx, slug, map[string]any{
		"extras": extras,
	})
	if err != nil {
		return "", handleEntError(ctx, "File", err)
	}

	downloadURL := fmt.Sprintf("/res/files/%s?type=download&token=%d", row.ID, expiresAt)
	return downloadURL, nil
}

// CreateVersion creates a new version of existing file
func (s *fileService) CreateVersion(ctx context.Context, slug string, file io.Reader, filename string) (*structs.ReadFile, error) {
	existing, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		return nil, handleEntError(ctx, "File", err)
	}

	so, storageConfig := ctxutil.GetStorage(ctx)

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if closer, ok := file.(io.Closer); ok {
		closer.Close()
	}

	fileHeader := &storage.FileHeader{
		Name: filename,
		Size: len(fileBytes),
		Path: fmt.Sprintf("versions/%s/%s", slug, filename),
		Type: "", // Will be determined by content
	}

	_, err = so.Put(fileHeader.Path, bytes.NewReader(fileBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to store file version: %w", err)
	}

	// Extract current extras
	extras := getExtrasFromFile(existing)
	var versions []string
	if v, ok := extras["versions"].([]string); ok {
		versions = v
	}
	versions = append(versions, existing.ID)

	createBody := &structs.CreateFileBody{
		Name:     fileHeader.Name,
		Path:     fileHeader.Path,
		Type:     fileHeader.Type,
		Size:     &fileHeader.Size,
		Storage:  storageConfig.Provider,
		Bucket:   storageConfig.Bucket,
		Endpoint: storageConfig.Endpoint,
		OwnerID:  existing.OwnerID,
		SpaceID:  existing.SpaceID,
	}

	// Copy extended properties
	if folderPath, ok := extras["folder_path"].(string); ok {
		createBody.FolderPath = folderPath
	}
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

// GetVersions retrieves all versions of a file
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

	updated, err := s.fileRepo.Update(ctx, slug, map[string]any{
		"extras": extras,
	})
	if err != nil {
		return nil, handleEntError(ctx, "File", err)
	}

	return s.serialize(updated), nil
}

// CreateThumbnail creates thumbnail for image file
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

	storage, _ := ctxutil.GetStorage(ctx)

	file, err := storage.Get(row.Path)
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

	thumbnailPath := fmt.Sprintf("thumbnails/%s", row.Path)
	_, err = storage.Put(thumbnailPath, bytes.NewReader(thumbnailBytes))
	if err != nil {
		return nil, fmt.Errorf("error storing thumbnail: %w", err)
	}

	extras := getExtrasFromFile(row)
	extras["thumbnail_path"] = thumbnailPath

	updated, err := s.fileRepo.Update(ctx, slug, map[string]any{
		"extras": extras,
	})
	if err != nil {
		return nil, handleEntError(ctx, "File", err)
	}

	return s.serialize(updated), nil
}

// serialize converts ent.File to structs.ReadFile
func (s *fileService) serialize(row *ent.File) *structs.ReadFile {
	extras := getExtrasFromFile(row)

	file := &structs.ReadFile{
		ID:        row.ID,
		Name:      row.Name,
		Path:      row.Path,
		Type:      row.Type,
		Size:      &row.Size,
		Storage:   row.Storage,
		Bucket:    row.Bucket,
		Endpoint:  row.Endpoint,
		OwnerID:   row.OwnerID,
		SpaceID:   row.SpaceID,
		Extras:    &row.Extras,
		CreatedBy: &row.CreatedBy,
		CreatedAt: &row.CreatedAt,
		UpdatedBy: &row.UpdatedBy,
		UpdatedAt: &row.UpdatedAt,
	}

	// Extract from extras
	if folderPath, ok := extras["folder_path"].(string); ok {
		file.FolderPath = folderPath
	}
	if accessLevel, ok := extras["access_level"].(string); ok {
		file.AccessLevel = structs.AccessLevel(accessLevel)
	}
	if tags, ok := extras["tags"].([]string); ok {
		file.Tags = tags
	}
	if isPublic, ok := extras["is_public"].(bool); ok {
		file.IsPublic = isPublic
	}
	if category, ok := extras["category"].(string); ok {
		file.Category = structs.FileCategory(category)
	}
	if exp, ok := extras["expires_at"].(int64); ok {
		file.ExpiresAt = &exp
		file.IsExpired = time.Now().Unix() > exp
	}

	// Set URLs if public
	if file.IsPublic {
		file.DownloadURL = fmt.Sprintf("/res/files/%s?type=download", row.ID)
		if thumbnailPath, ok := extras["thumbnail_path"].(string); ok && thumbnailPath != "" {
			file.ThumbnailURL = fmt.Sprintf("/res/files/%s?type=thumbnail", row.ID)
		}
	}

	return file
}

// Helper functions
func getExtrasFromFile(file *ent.File) map[string]any {
	if file == nil || file.Extras == nil {
		return make(map[string]any)
	}

	extras := make(map[string]any)
	for k, v := range file.Extras {
		extras[k] = v
	}
	return extras
}
