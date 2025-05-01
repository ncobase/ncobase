package event

import (
	"time"

	"github.com/ncobase/ncore/types"
)

// FileEventData represents data included in file events
type FileEventData struct {
	Timestamp time.Time   `json:"timestamp"`
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Path      string      `json:"path"`
	Type      string      `json:"type"`
	Size      int         `json:"size"`
	Storage   string      `json:"storage"`
	Bucket    string      `json:"bucket"`
	ObjectID  string      `json:"object_id"`
	TenantID  string      `json:"tenant_id"`
	UserID    string      `json:"user_id,omitempty"`
	Extras    *types.JSON `json:"extras,omitempty"`
}

// NewFileEventData creates a new file event data instance
func NewFileEventData(
	id, name, path, fileType string,
	size int,
	storage, bucket, objectID, tenantID, userID string,
	extras *types.JSON,
) *FileEventData {
	return &FileEventData{
		Timestamp: time.Now(),
		ID:        id,
		Name:      name,
		Path:      path,
		Type:      fileType,
		Size:      size,
		Storage:   storage,
		Bucket:    bucket,
		ObjectID:  objectID,
		TenantID:  tenantID,
		UserID:    userID,
		Extras:    extras,
	}
}

// BatchOperationEventData represents data for batch operations
type BatchOperationEventData struct {
	Timestamp   time.Time   `json:"timestamp"`
	OperationID string      `json:"operation_id"`
	ItemCount   int         `json:"item_count"`
	TenantID    string      `json:"tenant_id"`
	UserID      string      `json:"user_id,omitempty"`
	Status      string      `json:"status"`
	Message     string      `json:"message,omitempty"`
	Extras      *types.JSON `json:"extras,omitempty"`
}

// NewBatchOperationEventData creates a new batch operation event data instance
func NewBatchOperationEventData(
	operationID string,
	itemCount int,
	tenantID, userID, status, message string,
	extras *types.JSON,
) *BatchOperationEventData {
	return &BatchOperationEventData{
		Timestamp:   time.Now(),
		OperationID: operationID,
		ItemCount:   itemCount,
		TenantID:    tenantID,
		UserID:      userID,
		Status:      status,
		Message:     message,
		Extras:      extras,
	}
}

// StorageQuotaEventData represents data for storage quota events
type StorageQuotaEventData struct {
	Timestamp    time.Time `json:"timestamp"`
	TenantID     string    `json:"tenant_id"`
	CurrentUsage int64     `json:"current_usage"` // in bytes
	Quota        int64     `json:"quota"`         // in bytes
	UsagePercent float64   `json:"usage_percent"`
	StorageType  string    `json:"storage_type"`
}

// NewStorageQuotaEventData creates a new storage quota event data instance
func NewStorageQuotaEventData(
	tenantID string,
	currentUsage, quota int64,
	usagePercent float64,
	storageType string,
) *StorageQuotaEventData {
	return &StorageQuotaEventData{
		Timestamp:    time.Now(),
		TenantID:     tenantID,
		CurrentUsage: currentUsage,
		Quota:        quota,
		UsagePercent: usagePercent,
		StorageType:  storageType,
	}
}
