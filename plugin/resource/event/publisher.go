package event

import (
	"context"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// Publisher provides methods to publish resource-related events
type publisher struct {
	em ext.ManagerInterface
}

// NewPublisher creates a new event publisher
func NewPublisher(em ext.ManagerInterface) PublisherInterface {
	return &publisher{
		em: em,
	}
}

// Generic publish method
func (p *publisher) publish(ctx context.Context, eventType string, data any) {
	// Log the event based on its type
	switch d := data.(type) {
	case *FileEventData:
		logger.Infof(ctx, "Publishing file event: %s, id: %s, name: %s",
			eventType, d.ID, d.Name)
	case *BatchOperationEventData:
		logger.Infof(ctx, "Publishing batch operation event: %s, id: %s, status: %s",
			eventType, d.OperationID, d.Status)
	case *StorageQuotaEventData:
		logger.Infof(ctx, "Publishing storage quota event: %s, tenant: %s, usage: %.2f%%",
			eventType, d.SpaceID, d.UsagePercent)
	}

	// Publish the event through the extension manager
	if p.em != nil {
		p.em.PublishEvent(eventType, data)
	}
}

// File event publishing methods

func (p *publisher) PublishFileCreated(ctx context.Context, data *FileEventData) {
	p.publish(ctx, FileCreated, data)
}

func (p *publisher) PublishFileUpdated(ctx context.Context, data *FileEventData) {
	p.publish(ctx, FileUpdated, data)
}

func (p *publisher) PublishFileDeleted(ctx context.Context, data *FileEventData) {
	p.publish(ctx, FileDeleted, data)
}

func (p *publisher) PublishFileAccessed(ctx context.Context, data *FileEventData) {
	p.publish(ctx, FileAccessed, data)
}

// Batch operation event publishing methods

func (p *publisher) PublishBatchUploadStarted(ctx context.Context, data *BatchOperationEventData) {
	p.publish(ctx, BatchUploadStarted, data)
}

func (p *publisher) PublishBatchUploadComplete(ctx context.Context, data *BatchOperationEventData) {
	p.publish(ctx, BatchUploadComplete, data)
}

func (p *publisher) PublishBatchUploadFailed(ctx context.Context, data *BatchOperationEventData) {
	p.publish(ctx, BatchUploadFailed, data)
}

// Storage quota event publishing methods

func (p *publisher) PublishStorageQuotaWarning(ctx context.Context, data *StorageQuotaEventData) {
	p.publish(ctx, StorageQuotaWarning, data)
}

func (p *publisher) PublishStorageQuotaExceeded(ctx context.Context, data *StorageQuotaEventData) {
	p.publish(ctx, StorageQuotaExceeded, data)
}
