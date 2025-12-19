package event

import (
	"context"
	"fmt"
	"time"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// NotifierInterface defines notification methods
type NotifierInterface interface {
	NotifyQuotaWarning(data *StorageQuotaEventData)
	NotifyQuotaExceeded(data *StorageQuotaEventData)
	NotifyLargeFileUploaded(data *FileEventData)
	NotifyBatchComplete(data *BatchOperationEventData)
	NotifyBatchFailed(data *BatchOperationEventData)
}

// notifier handles sending notifications for various events
type notifier struct {
	em ext.ManagerInterface
}

// NewNotifier creates new notifier
func NewNotifier(em ext.ManagerInterface) NotifierInterface {
	return &notifier{em: em}
}

// NotifyQuotaWarning sends quota warning notifications
func (n *notifier) NotifyQuotaWarning(data *StorageQuotaEventData) {
	payload := NotificationPayload{
		Severity:  "warning",
		Category:  "storage.quota.warning",
		Title:     "Storage quota warning",
		Message:   formatQuotaMessage(data),
		SpaceID:   data.SpaceID,
		Timestamp: data.Timestamp,
	}
	n.publish(context.Background(), payload)

	logger.Infof(context.Background(),
		"[NOTIFICATION] Quota warning for space %s: %.2f%% used (%d/%d bytes)",
		data.SpaceID, data.UsagePercent, data.CurrentUsage, data.Quota)
}

// NotifyQuotaExceeded sends quota exceeded notifications
func (n *notifier) NotifyQuotaExceeded(data *StorageQuotaEventData) {
	payload := NotificationPayload{
		Severity:  "error",
		Category:  "storage.quota.exceeded",
		Title:     "Storage quota exceeded",
		Message:   formatQuotaMessage(data),
		SpaceID:   data.SpaceID,
		Timestamp: data.Timestamp,
	}
	n.publish(context.Background(), payload)

	logger.Errorf(context.Background(),
		"[NOTIFICATION] Quota exceeded for space %s: %.2f%% used (%d/%d bytes)",
		data.SpaceID, data.UsagePercent, data.CurrentUsage, data.Quota)
}

// NotifyLargeFileUploaded sends notifications for large file uploads
func (n *notifier) NotifyLargeFileUploaded(data *FileEventData) {
	payload := NotificationPayload{
		Severity:  "info",
		Category:  "file.large_upload",
		Title:     "Large file uploaded",
		Message:   formatFileMessage(data),
		SpaceID:   data.SpaceID,
		UserID:    data.UserID,
		Timestamp: data.Timestamp,
	}
	n.publish(context.Background(), payload)

	logger.Infof(context.Background(),
		"[NOTIFICATION] Large file uploaded: %s (%d bytes) in space %s",
		data.Name, data.Size, data.SpaceID)
}

// NotifyBatchComplete sends batch operation completion notifications
func (n *notifier) NotifyBatchComplete(data *BatchOperationEventData) {
	payload := NotificationPayload{
		Severity:  "success",
		Category:  "batch.completed",
		Title:     "Batch operation completed",
		Message:   formatBatchMessage(data),
		SpaceID:   data.SpaceID,
		UserID:    data.UserID,
		Timestamp: data.Timestamp,
	}
	n.publish(context.Background(), payload)

	logger.Infof(context.Background(),
		"[NOTIFICATION] Batch operation completed: %s (%d items) in space %s",
		data.OperationID, data.ItemCount, data.SpaceID)
}

// NotifyBatchFailed sends batch operation failure notifications
func (n *notifier) NotifyBatchFailed(data *BatchOperationEventData) {
	payload := NotificationPayload{
		Severity:  "error",
		Category:  "batch.failed",
		Title:     "Batch operation failed",
		Message:   formatBatchMessage(data),
		SpaceID:   data.SpaceID,
		UserID:    data.UserID,
		Timestamp: data.Timestamp,
	}
	n.publish(context.Background(), payload)

	logger.Errorf(context.Background(),
		"[NOTIFICATION] Batch operation failed: %s (%d items) in space %s - %s",
		data.OperationID, data.ItemCount, data.SpaceID, data.Message)
}

// NotificationPayload represents a normalized notification message.
type NotificationPayload struct {
	Severity  string    `json:"severity"`
	Category  string    `json:"category"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	SpaceID   string    `json:"space_id,omitempty"`
	UserID    string    `json:"user_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

func (n *notifier) publish(ctx context.Context, payload NotificationPayload) {
	if n.em != nil {
		n.em.PublishEvent(ResourceNotification, payload)
	}
}

func formatQuotaMessage(data *StorageQuotaEventData) string {
	return fmt.Sprintf("Space %s is at %.2f%% usage (%d/%d bytes).", data.SpaceID, data.UsagePercent, data.CurrentUsage, data.Quota)
}

func formatFileMessage(data *FileEventData) string {
	return fmt.Sprintf("File %s (%d bytes) uploaded in space %s.", data.Name, data.Size, data.SpaceID)
}

func formatBatchMessage(data *BatchOperationEventData) string {
	if data.Message != "" {
		return fmt.Sprintf("Operation %s processed %d items. %s", data.OperationID, data.ItemCount, data.Message)
	}
	return fmt.Sprintf("Operation %s processed %d items.", data.OperationID, data.ItemCount)
}
