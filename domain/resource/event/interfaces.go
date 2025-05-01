package event

import "context"

const (
	// File events

	FileCreated  = "resource.file.created"
	FileUpdated  = "resource.file.updated"
	FileDeleted  = "resource.file.deleted"
	FileAccessed = "resource.file.accessed"

	// Folder events

	FolderCreated = "resource.folder.created"
	FolderUpdated = "resource.folder.updated"
	FolderDeleted = "resource.folder.deleted"

	// Batch operation events

	BatchUploadStarted  = "resource.batch.upload.started"
	BatchUploadComplete = "resource.batch.upload.completed"
	BatchUploadFailed   = "resource.batch.upload.failed"

	// Storage events

	StorageQuotaWarning  = "resource.storage.quota.warning"
	StorageQuotaExceeded = "resource.storage.quota.exceeded"
)

// PublisherInterface defines methods to publish resource-related events
type PublisherInterface interface {
	// File events

	PublishFileCreated(ctx context.Context, data *FileEventData)
	PublishFileUpdated(ctx context.Context, data *FileEventData)
	PublishFileDeleted(ctx context.Context, data *FileEventData)
	PublishFileAccessed(ctx context.Context, data *FileEventData)

	// Batch operation events

	PublishBatchUploadStarted(ctx context.Context, data *BatchOperationEventData)
	PublishBatchUploadComplete(ctx context.Context, data *BatchOperationEventData)
	PublishBatchUploadFailed(ctx context.Context, data *BatchOperationEventData)

	// Storage quota events

	PublishStorageQuotaWarning(ctx context.Context, data *StorageQuotaEventData)
	PublishStorageQuotaExceeded(ctx context.Context, data *StorageQuotaEventData)
}
