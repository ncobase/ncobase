package event

import (
	"context"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// QuotaUpdaterInterface abstracts quota operations for event handler
type QuotaUpdaterInterface interface {
	UpdateUsage(ctx context.Context, spaceID string, quotaType string, delta int64) error
}

// HandlerInterface defines event handler methods
type HandlerInterface interface {
	HandleFileCreated(data any)
	HandleFileDeleted(data any)
	HandleFileUpdated(data any)
	HandleFileAccessed(data any)
	HandleQuotaWarning(data any)
	HandleQuotaExceeded(data any)
	HandleBatchUploadStarted(data any)
	HandleBatchUploadComplete(data any)
	HandleBatchUploadFailed(data any)
	SetQuotaUpdater(updater QuotaUpdaterInterface)
}

// handler handles various resource events
type handler struct {
	quotaUpdater QuotaUpdaterInterface
	notifier     NotifierInterface
	em           ext.ManagerInterface
}

// NewHandler creates new event handler
func NewHandler(em ext.ManagerInterface, notifier NotifierInterface) HandlerInterface {
	if notifier == nil {
		notifier = NewNotifier(em)
	}
	return &handler{
		em:       em,
		notifier: notifier,
	}
}

// SetQuotaUpdater sets the quota updater dependency
func (h *handler) SetQuotaUpdater(updater QuotaUpdaterInterface) {
	h.quotaUpdater = updater
}

// HandleFileCreated handles file creation events
func (h *handler) HandleFileCreated(data any) {
	eventData, ok := data.(*FileEventData)
	if !ok {
		logger.Warnf(context.Background(), "Invalid file created event data")
		return
	}

	logger.Infof(context.Background(), "File created: %s, size: %d bytes, space: %s, owner: %s",
		eventData.Name, eventData.Size, eventData.SpaceID, eventData.OwnerID)

	// Send notification for large files
	if eventData.Size > 100*1024*1024 { // > 100MB
		h.notifier.NotifyLargeFileUploaded(eventData)
	}
}

// HandleFileDeleted handles file deletion events
func (h *handler) HandleFileDeleted(data any) {
	eventData, ok := data.(*FileEventData)
	if !ok {
		logger.Warnf(context.Background(), "Invalid file deleted event data")
		return
	}

	logger.Infof(context.Background(), "File deleted: %s, space: %s, owner: %s",
		eventData.Name, eventData.SpaceID, eventData.OwnerID)

	// Update usage in quota service when file is deleted
	if h.quotaUpdater != nil && eventData.Size > 0 {
		ctx := context.Background()
		err := h.quotaUpdater.UpdateUsage(ctx, eventData.SpaceID, "storage", -int64(eventData.Size))
		if err != nil {
			logger.Warnf(ctx, "Failed to update quota usage after file deletion: %v", err)
		} else {
			logger.Debugf(ctx, "Updated quota usage after file deletion: -%d bytes", eventData.Size)
		}
	}
}

// HandleFileUpdated handles file update events
func (h *handler) HandleFileUpdated(data any) {
	eventData, ok := data.(*FileEventData)
	if !ok {
		logger.Warnf(context.Background(), "Invalid file updated event data")
		return
	}

	logger.Infof(context.Background(), "File updated: %s, space: %s, owner: %s",
		eventData.Name, eventData.SpaceID, eventData.OwnerID)

	// Invalidate caches, update search index, etc.
	h.handleFileIndexUpdate(eventData)
}

// HandleFileAccessed handles file access events
func (h *handler) HandleFileAccessed(data any) {
	eventData, ok := data.(*FileEventData)
	if !ok {
		logger.Warnf(context.Background(), "Invalid file accessed event data")
		return
	}

	logger.Debugf(context.Background(), "File accessed: %s, space: %s, user: %s",
		eventData.Name, eventData.SpaceID, eventData.UserID)

	// Update access analytics
	h.updateAccessAnalytics(eventData)
}

// HandleQuotaWarning handles storage quota warning events
func (h *handler) HandleQuotaWarning(data any) {
	eventData, ok := data.(*StorageQuotaEventData)
	if !ok {
		logger.Warnf(context.Background(), "Invalid quota warning event data")
		return
	}

	logger.Warnf(context.Background(), "Storage quota warning: space %s at %.2f%% (%d/%d bytes)",
		eventData.SpaceID, eventData.UsagePercent, eventData.CurrentUsage, eventData.Quota)

	// Send quota warning notification
	h.notifier.NotifyQuotaWarning(eventData)
}

// HandleQuotaExceeded handles storage quota exceeded events
func (h *handler) HandleQuotaExceeded(data any) {
	eventData, ok := data.(*StorageQuotaEventData)
	if !ok {
		logger.Warnf(context.Background(), "Invalid quota exceeded event data")
		return
	}

	logger.Errorf(context.Background(), "Storage quota exceeded: space %s at %.2f%% (%d/%d bytes)",
		eventData.SpaceID, eventData.UsagePercent, eventData.CurrentUsage, eventData.Quota)

	// Send urgent quota exceeded notification
	h.notifier.NotifyQuotaExceeded(eventData)
}

// HandleBatchUploadStarted handles batch upload started events
func (h *handler) HandleBatchUploadStarted(data any) {
	eventData, ok := data.(*BatchOperationEventData)
	if !ok {
		logger.Warnf(context.Background(), "Invalid batch upload started event data")
		return
	}

	logger.Infof(context.Background(), "Batch upload started: operation %s, %d files, space: %s",
		eventData.OperationID, eventData.ItemCount, eventData.SpaceID)

	// Track batch operations
	h.trackBatchOperation(eventData, "started")
}

// HandleBatchUploadComplete handles batch upload completion events
func (h *handler) HandleBatchUploadComplete(data any) {
	eventData, ok := data.(*BatchOperationEventData)
	if !ok {
		logger.Warnf(context.Background(), "Invalid batch upload complete event data")
		return
	}

	logger.Infof(context.Background(), "Batch upload completed: operation %s, %d files, space: %s",
		eventData.OperationID, eventData.ItemCount, eventData.SpaceID)

	// Send completion notification
	h.notifier.NotifyBatchComplete(eventData)
	h.trackBatchOperation(eventData, "completed")
}

// HandleBatchUploadFailed handles batch upload failure events
func (h *handler) HandleBatchUploadFailed(data any) {
	eventData, ok := data.(*BatchOperationEventData)
	if !ok {
		logger.Warnf(context.Background(), "Invalid batch upload failed event data")
		return
	}

	logger.Errorf(context.Background(), "Batch upload failed: operation %s, %d files, space: %s, message: %s",
		eventData.OperationID, eventData.ItemCount, eventData.SpaceID, eventData.Message)

	// Send failure notification
	h.notifier.NotifyBatchFailed(eventData)
	h.trackBatchOperation(eventData, "failed")
}

// Additional processing

func (h *handler) handleFileIndexUpdate(eventData *FileEventData) {
	logger.Debugf(context.Background(), "Updating file index for: %s", eventData.ID)
	if h.em != nil {
		h.em.PublishEvent(FileIndexUpdateRequested, eventData)
	}
}

func (h *handler) updateAccessAnalytics(eventData *FileEventData) {
	logger.Debugf(context.Background(), "Updating access analytics for: %s", eventData.ID)
	if h.em != nil {
		h.em.PublishEvent(FileAccessAnalyticsUpdated, eventData)
	}
}

func (h *handler) trackBatchOperation(eventData *BatchOperationEventData, status string) {
	eventData.Status = status
	logger.Debugf(context.Background(), "Tracking batch operation %s: %s", eventData.OperationID, status)
	if h.em != nil {
		h.em.PublishEvent(BatchOperationTracked, eventData)
	}
}
