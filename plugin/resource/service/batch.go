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
	"time"

	"github.com/google/uuid"
	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/data/storage"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/validation/validator"
)

// Create a ReadCloser that wraps a bytes.Reader
type readCloser struct {
	*bytes.Reader
}

// Close implements the io.Closer interface
func (r *readCloser) Close() error {
	return nil
}

// BatchServiceInterface defines the interface for batch operations
type BatchServiceInterface interface {
	BatchUpload(ctx context.Context, files []*multipart.FileHeader, params *structs.BatchUploadParams) (*structs.BatchUploadResult, error)
	ProcessImages(ctx context.Context, files []*structs.ReadFile, options *structs.ProcessingOptions) ([]*structs.ReadFile, error)
}

// batchService handles batch operations on resources
type batchService struct {
	file           FileServiceInterface
	imageProcessor ImageProcessorInterface
	publisher      event.PublisherInterface
}

// NewBatchService creates a new batch service
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

// BatchUpload handles uploading multiple files in a batch
func (s *batchService) BatchUpload(
	ctx context.Context,
	files []*multipart.FileHeader,
	params *structs.BatchUploadParams,
) (*structs.BatchUploadResult, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files to upload")
	}

	if params.TenantID == "" {
		return nil, fmt.Errorf("tenant ID is required")
	}

	if params.ObjectID == "" {
		return nil, fmt.Errorf("object ID is required")
	}

	// Generate operation ID
	operationID := uuid.New().String()

	// Create result structure
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
			TenantID:    params.TenantID,
			UserID:      ctxutil.GetUserID(ctx),
			Status:      "started",
			Message:     fmt.Sprintf("Started batch upload of %d files", len(files)),
		}
		s.publisher.PublishBatchUploadStarted(ctx, eventData)
	}

	// Use a wait group to handle concurrent uploads
	var wg sync.WaitGroup
	var mu sync.Mutex // Mutex for result updates

	// Control concurrency with a semaphore
	maxConcurrent := 5 // Maximum number of concurrent uploads
	sem := make(chan struct{}, maxConcurrent)

	// Process each file
	for _, fileHeader := range files {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		// Process files concurrently
		go func(header *multipart.FileHeader) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			// Open the file
			file, err := header.Open()
			if err != nil {
				mu.Lock()
				batchResult.FailureCount++
				batchResult.FailedFiles = append(batchResult.FailedFiles, header.Filename)
				batchResult.Errors = append(batchResult.Errors, fmt.Sprintf("Failed to open file %s: %v", header.Filename, err))
				mu.Unlock()
				return
			}
			defer file.Close() // Ensure file is closed

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
			body.ObjectID = params.ObjectID
			body.TenantID = params.TenantID

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
			category := structs.GetFileCategory(fileHeader.Ext)
			metadata := &structs.FileMetadata{
				Category:     category,
				CreationDate: &time.Time{},
				CustomFields: make(map[string]any),
			}

			// If it's an image, extract dimensions
			if validator.IsImageFile(header.Filename) && s.imageProcessor != nil {
				// Create a reader for the file for dimension extraction
				dimensionReader := bytes.NewReader(fileContent)
				width, height, _ := s.imageProcessor.GetImageDimensions(ctx, dimensionReader, header.Filename)
				metadata.Width = &width
				metadata.Height = &height

				// Process the image if requested
				if params.ProcessingOptions != nil && params.ProcessingOptions.CreateThumbnail {
					// Process using the file content directly
					s.processImageOnUpload(ctx, fileContent, header.Filename, params.ProcessingOptions, metadata)
				}
			}

			body.Metadata = metadata

			// Create a reader for storage
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

	// Wait for all goroutines to complete
	wg.Wait()

	// Publish batch upload completed/failed event
	if s.publisher != nil {
		status := "completed"
		message := fmt.Sprintf("Completed batch upload of %d files. Success: %d, Failure: %d",
			batchResult.TotalFiles, batchResult.SuccessCount, batchResult.FailureCount)

		if batchResult.FailureCount > 0 {
			status = "failed"
			message = fmt.Sprintf("Completed batch upload with %d failures out of %d files",
				batchResult.FailureCount, batchResult.TotalFiles)
		}

		eventData := &event.BatchOperationEventData{
			OperationID: operationID,
			ItemCount:   len(files),
			TenantID:    params.TenantID,
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

// Helper function to process image during upload
func (s *batchService) processImageOnUpload(
	ctx context.Context,
	fileContent []byte,
	filename string,
	options *structs.ProcessingOptions,
	metadata *structs.FileMetadata,
) {
	// Create a new reader from the file content
	reader := bytes.NewReader(fileContent)

	// Process the image according to options
	_, processMetadata, err := s.imageProcessor.ProcessImage(ctx, reader, filename, options)
	if err != nil {
		logger.Warnf(ctx, "Failed to process image %s: %v", filename, err)
		return
	}

	// Update metadata with processing results
	if metadata.CustomFields == nil {
		metadata.CustomFields = make(map[string]any)
	}

	metadata.CustomFields["processing"] = processMetadata

	// If dimensions changed, update them
	if width, ok := processMetadata["width"].(int); ok {
		metadata.Width = &width
	}

	if height, ok := processMetadata["height"].(int); ok {
		metadata.Height = &height
	}
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
		// Check if it's an image
		if !validator.IsImageFile(file.Path) {
			// Skip non-images
			processedFiles = append(processedFiles, file)
			continue
		}

		// TODO: Implement image processing for existing files
		// This would involve:
		// 1. Fetching the file from storage
		// 2. Processing the image
		// 3. Storing the processed image
		// 4. Updating the file metadata

		// For now, just add the original file
		processedFiles = append(processedFiles, file)
	}

	return processedFiles, nil
}
