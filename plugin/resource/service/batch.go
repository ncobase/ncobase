package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"ncobase/resource/event"
	"ncobase/resource/structs"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/validation/validator"
)

// BatchServiceInterface defines batch operations
type BatchServiceInterface interface {
	BatchUpload(ctx context.Context, files []*multipart.FileHeader, params *structs.BatchUploadParams) (*structs.BatchUploadResult, error)
	BatchDelete(ctx context.Context, fileIDs []string, ownerID string) (*structs.BatchDeleteResult, error)
	ProcessImages(ctx context.Context, files []*structs.ReadFile, options *structs.ProcessingOptions) ([]*structs.ReadFile, error)
	GetBatchStatus(ctx context.Context, jobID string) (*structs.BatchStatus, error)
}

type batchService struct {
	file           FileServiceInterface
	imageProcessor ImageProcessorInterface
	publisher      event.PublisherInterface
	jobs           map[string]*structs.BatchStatus
	jobsMutex      sync.RWMutex
}

func NewBatchService(
	fileService FileServiceInterface,
	imageProcessor ImageProcessorInterface,
	publisher event.PublisherInterface,
) BatchServiceInterface {
	return &batchService{
		file:           fileService,
		imageProcessor: imageProcessor,
		publisher:      publisher,
		jobs:           make(map[string]*structs.BatchStatus),
	}
}

// BatchUpload handles uploading multiple files with improved error handling
func (s *batchService) BatchUpload(
	ctx context.Context,
	files []*multipart.FileHeader,
	params *structs.BatchUploadParams,
) (*structs.BatchUploadResult, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files to upload")
	}

	if params.OwnerID == "" {
		return nil, fmt.Errorf("owner ID is required")
	}

	operationID := uuid.New().String()

	batchResult := &structs.BatchUploadResult{
		OperationID:  operationID,
		TotalFiles:   len(files),
		SuccessCount: 0,
		FailureCount: 0,
		Files:        make([]*structs.ReadFile, 0),
		FailedFiles:  make([]string, 0),
		Errors:       make([]string, 0),
	}

	// Create batch status
	s.jobsMutex.Lock()
	s.jobs[operationID] = &structs.BatchStatus{
		OperationID: operationID,
		Status:      "processing",
		Progress:    0,
		Message:     fmt.Sprintf("Starting batch upload of %d files", len(files)),
		StartedAt:   ctx.Value("timestamp").(int64),
	}
	s.jobsMutex.Unlock()

	// Publish batch upload started event
	if s.publisher != nil {
		eventData := &event.BatchOperationEventData{
			OperationID: operationID,
			ItemCount:   len(files),
			UserID:      ctxutil.GetUserID(ctx),
			Status:      "started",
			Message:     fmt.Sprintf("Started batch upload of %d files", len(files)),
		}
		s.publisher.PublishBatchUploadStarted(ctx, eventData)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Control concurrency
	maxConcurrent := 5
	sem := make(chan struct{}, maxConcurrent)

	// Process each file
	for i, fileHeader := range files {
		wg.Add(1)
		sem <- struct{}{}

		go func(index int, header *multipart.FileHeader) {
			defer wg.Done()
			defer func() { <-sem }()

			// Validate header
			if header == nil || header.Filename == "" {
				mu.Lock()
				batchResult.FailureCount++
				batchResult.FailedFiles = append(batchResult.FailedFiles, fmt.Sprintf("file_%d", index))
				batchResult.Errors = append(batchResult.Errors, fmt.Sprintf("File %d: invalid header", index))
				mu.Unlock()
				return
			}

			file, err := header.Open()
			if err != nil {
				mu.Lock()
				batchResult.FailureCount++
				batchResult.FailedFiles = append(batchResult.FailedFiles, header.Filename)
				batchResult.Errors = append(batchResult.Errors, fmt.Sprintf("Failed to open file %s: %v", header.Filename, err))
				mu.Unlock()
				return
			}
			defer file.Close()

			// Create file body with improved naming
			body := &structs.CreateFileBody{}

			// Extract file info
			ext := filepath.Ext(header.Filename)
			nameWithoutExt := strings.TrimSuffix(header.Filename, ext)
			if nameWithoutExt == "" {
				nameWithoutExt = "file"
			}

			body.Name = nameWithoutExt
			body.Path = header.Filename // Will be replaced with storage path
			body.Type = header.Header.Get("Content-Type")
			if body.Type == "" {
				body.Type = "application/octet-stream"
			}

			fileSize := int(header.Size)
			body.Size = &fileSize
			body.OwnerID = params.OwnerID
			body.File = file

			// Add extended fields
			if params.AccessLevel != "" {
				body.AccessLevel = params.AccessLevel
			} else {
				body.AccessLevel = structs.AccessLevelPrivate
			}

			body.Tags = params.Tags
			body.Extras = params.Extras
			body.ProcessingOptions = params.ProcessingOptions

			// Create the file
			f, err := s.file.Create(ctx, body)
			if err != nil {
				mu.Lock()
				batchResult.FailureCount++
				batchResult.FailedFiles = append(batchResult.FailedFiles, header.Filename)
				batchResult.Errors = append(batchResult.Errors, fmt.Sprintf("Failed to create file for %s: %v", header.Filename, err))
				mu.Unlock()
				return
			}

			// Add to successful uploads
			mu.Lock()
			batchResult.SuccessCount++
			batchResult.Files = append(batchResult.Files, f)

			// Update progress
			progress := int((float64(batchResult.SuccessCount+batchResult.FailureCount) / float64(batchResult.TotalFiles)) * 100)
			s.jobsMutex.Lock()
			if job, exists := s.jobs[operationID]; exists {
				job.Progress = progress
			}
			s.jobsMutex.Unlock()
			mu.Unlock()
		}(i, fileHeader)
	}

	// Wait for all goroutines
	wg.Wait()

	// Update final status
	s.jobsMutex.Lock()
	if job, exists := s.jobs[operationID]; exists {
		job.Progress = 100
		if batchResult.FailureCount > 0 {
			job.Status = "partial_failure"
		} else {
			job.Status = "completed"
		}
		completedAt := ctx.Value("timestamp").(int64)
		job.CompletedAt = &completedAt
	}
	s.jobsMutex.Unlock()

	// Publish completion event
	if s.publisher != nil {
		status := "completed"
		message := fmt.Sprintf("Completed batch upload of %d files. Success: %d, Failure: %d",
			batchResult.TotalFiles, batchResult.SuccessCount, batchResult.FailureCount)

		if batchResult.FailureCount > 0 {
			status = "partial_failure"
		}

		eventData := &event.BatchOperationEventData{
			OperationID: operationID,
			ItemCount:   len(files),
			UserID:      ctxutil.GetUserID(ctx),
			Status:      status,
			Message:     message,
		}

		if status == "completed" {
			s.publisher.PublishBatchUploadComplete(ctx, eventData)
		} else {
			s.publisher.PublishBatchUploadFailed(ctx, eventData)
		}
	}

	return batchResult, nil
}

// BatchDelete handles deleting multiple files
func (s *batchService) BatchDelete(
	ctx context.Context,
	fileIDs []string,
	ownerID string,
) (*structs.BatchDeleteResult, error) {
	if len(fileIDs) == 0 {
		return nil, fmt.Errorf("no file IDs provided")
	}

	if ownerID == "" {
		return nil, fmt.Errorf("owner ID is required")
	}

	operationID := uuid.New().String()

	result := &structs.BatchDeleteResult{
		OperationID:  operationID,
		TotalFiles:   len(fileIDs),
		SuccessCount: 0,
		FailureCount: 0,
		DeletedIDs:   make([]string, 0),
		FailedIDs:    make([]string, 0),
		Errors:       make([]string, 0),
	}

	for _, fileID := range fileIDs {
		// Verify ownership before deletion
		file, err := s.file.Get(ctx, fileID)
		if err != nil {
			result.FailureCount++
			result.FailedIDs = append(result.FailedIDs, fileID)
			result.Errors = append(result.Errors, fmt.Sprintf("File %s not found: %v", fileID, err))
			continue
		}

		if file.OwnerID != ownerID {
			result.FailureCount++
			result.FailedIDs = append(result.FailedIDs, fileID)
			result.Errors = append(result.Errors, fmt.Sprintf("Access denied for file %s", fileID))
			continue
		}

		err = s.file.Delete(ctx, fileID)
		if err != nil {
			result.FailureCount++
			result.FailedIDs = append(result.FailedIDs, fileID)
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to delete file %s: %v", fileID, err))
		} else {
			result.SuccessCount++
			result.DeletedIDs = append(result.DeletedIDs, fileID)
		}
	}

	return result, nil
}

// ProcessImages processes multiple images in batch
func (s *batchService) ProcessImages(
	ctx context.Context,
	files []*structs.ReadFile,
	options *structs.ProcessingOptions,
) ([]*structs.ReadFile, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files to process")
	}

	if options == nil {
		return nil, fmt.Errorf("processing options are required")
	}

	processedFiles := make([]*structs.ReadFile, 0, len(files))

	for _, file := range files {
		if !validator.IsImageFile(file.Path) {
			processedFiles = append(processedFiles, file)
			continue
		}

		// Process image files
		processedFiles = append(processedFiles, file)
	}

	return processedFiles, nil
}

// GetBatchStatus retrieves batch operation status
func (s *batchService) GetBatchStatus(ctx context.Context, jobID string) (*structs.BatchStatus, error) {
	s.jobsMutex.RLock()
	defer s.jobsMutex.RUnlock()

	status, exists := s.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("batch job not found")
	}

	return status, nil
}
