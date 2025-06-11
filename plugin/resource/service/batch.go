package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"ncobase/resource/event"
	"ncobase/resource/structs"
	"sync"

	"github.com/google/uuid"
	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/data/storage"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/validation/validator"
)

// BatchServiceInterface defines batch operations
type BatchServiceInterface interface {
	BatchUpload(ctx context.Context, files []*multipart.FileHeader, params *structs.BatchUploadParams) (*structs.BatchUploadResult, error)
	ProcessImages(ctx context.Context, files []*structs.ReadFile, options *structs.ProcessingOptions) ([]*structs.ReadFile, error)
}

type batchService struct {
	file           FileServiceInterface
	imageProcessor ImageProcessorInterface
	publisher      event.PublisherInterface
}

// NewBatchService creates new batch service
func NewBatchService(
	fileService FileServiceInterface,
	imageProcessor ImageProcessorInterface,
	publisher event.PublisherInterface,
) BatchServiceInterface {
	return &batchService{
		file:           fileService,
		imageProcessor: imageProcessor,
		publisher:      publisher,
	}
}

// BatchUpload handles uploading multiple files
func (s *batchService) BatchUpload(
	ctx context.Context,
	files []*multipart.FileHeader,
	params *structs.BatchUploadParams,
) (*structs.BatchUploadResult, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files to upload")
	}

	if params.SpaceID == "" {
		return nil, fmt.Errorf("space ID is required")
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

	// Publish batch upload started event
	if s.publisher != nil {
		eventData := &event.BatchOperationEventData{
			OperationID: operationID,
			ItemCount:   len(files),
			SpaceID:     params.SpaceID,
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
	for _, fileHeader := range files {
		wg.Add(1)
		sem <- struct{}{}

		go func(header *multipart.FileHeader) {
			defer wg.Done()
			defer func() { <-sem }()

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

			// Read file content
			fileContent, err := io.ReadAll(file)
			if err != nil {
				mu.Lock()
				batchResult.FailureCount++
				batchResult.FailedFiles = append(batchResult.FailedFiles, header.Filename)
				batchResult.Errors = append(batchResult.Errors, fmt.Sprintf("Failed to read file %s: %v", header.Filename, err))
				mu.Unlock()
				return
			}

			// Create file body
			body := &structs.CreateFileBody{}
			fileHeader := storage.GetFileHeader(header, params.FolderPath)

			body.Path = fileHeader.Path
			body.Type = fileHeader.Type
			body.Name = fileHeader.Name
			body.Size = &fileHeader.Size
			body.OwnerID = params.OwnerID
			body.SpaceID = params.SpaceID

			// Add extended fields
			body.FolderPath = params.FolderPath
			if params.AccessLevel != "" {
				body.AccessLevel = params.AccessLevel
			} else {
				body.AccessLevel = structs.AccessLevelPrivate
			}

			body.Tags = params.Tags
			body.Extras = params.Extras

			// Set metadata
			_ = structs.GetFileCategory(fileHeader.Ext)

			// Process image if needed
			if validator.IsImageFile(header.Filename) && s.imageProcessor != nil {
				if params.ProcessingOptions != nil && params.ProcessingOptions.CreateThumbnail {
					s.processImageOnUpload(ctx, fileContent, header.Filename, params.ProcessingOptions)
				}
			}

			// Create reader for storage
			body.File = &readCloser{bytes.NewReader(fileContent)}

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
			mu.Unlock()
		}(fileHeader)
	}

	// Wait for all goroutines
	wg.Wait()

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
			SpaceID:     params.SpaceID,
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

// processImageOnUpload processes image during upload
func (s *batchService) processImageOnUpload(
	ctx context.Context,
	fileContent []byte,
	filename string,
	options *structs.ProcessingOptions,
) {
	reader := bytes.NewReader(fileContent)

	_, processMetadata, err := s.imageProcessor.ProcessImage(ctx, reader, filename, options)
	if err != nil {
		logger.Warnf(ctx, "Failed to process image %s: %v", filename, err)
		return
	}

	logger.Infof(ctx, "Processed image %s: %v", filename, processMetadata)
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

		// TODO: Implement image processing for existing files
		processedFiles = append(processedFiles, file)
	}

	return processedFiles, nil
}
