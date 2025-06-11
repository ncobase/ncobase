package event

import (
	"time"

	"github.com/ncobase/ncore/types"
)

// FileEventData represents file event data
type FileEventData struct {
	Timestamp time.Time   `json:"timestamp"`
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Path      string      `json:"path"`
	Type      string      `json:"type"`
	Size      int         `json:"size"`
	Storage   string      `json:"storage"`
	Bucket    string      `json:"bucket"`
	OwnerID   string      `json:"owner_id"`
	SpaceID   string      `json:"space_id"`
	UserID    string      `json:"user_id,omitempty"`
	Extras    *types.JSON `json:"extras,omitempty"`
}

// NewFileEventData creates new file event data
func NewFileEventData(
	id, name, path, fileType string,
	size int,
	storage, bucket, ownerID, spaceID, userID string,
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
		OwnerID:   ownerID,
		SpaceID:   spaceID,
		UserID:    userID,
		Extras:    extras,
	}
}

// BatchOperationEventData represents batch operation event data
type BatchOperationEventData struct {
	Timestamp   time.Time   `json:"timestamp"`
	OperationID string      `json:"operation_id"`
	ItemCount   int         `json:"item_count"`
	SpaceID     string      `json:"space_id"`
	UserID      string      `json:"user_id,omitempty"`
	Status      string      `json:"status"`
	Message     string      `json:"message,omitempty"`
	Extras      *types.JSON `json:"extras,omitempty"`
}

// NewBatchOperationEventData creates new batch operation event data
func NewBatchOperationEventData(
	operationID string,
	itemCount int,
	spaceID, userID, status, message string,
	extras *types.JSON,
) *BatchOperationEventData {
	return &BatchOperationEventData{
		Timestamp:   time.Now(),
		OperationID: operationID,
		ItemCount:   itemCount,
		SpaceID:     spaceID,
		UserID:      userID,
		Status:      status,
		Message:     message,
		Extras:      extras,
	}
}

// StorageQuotaEventData represents storage quota event data
type StorageQuotaEventData struct {
	Timestamp    time.Time `json:"timestamp"`
	SpaceID      string    `json:"space_id"`
	CurrentUsage int64     `json:"current_usage"` // in bytes
	Quota        int64     `json:"quota"`         // in bytes
	UsagePercent float64   `json:"usage_percent"`
	StorageType  string    `json:"storage_type"`
}

// NewStorageQuotaEventData creates new storage quota event data
func NewStorageQuotaEventData(
	spaceID string,
	currentUsage, quota int64,
	usagePercent float64,
	storageType string,
) *StorageQuotaEventData {
	return &StorageQuotaEventData{
		Timestamp:    time.Now(),
		SpaceID:      spaceID,
		CurrentUsage: currentUsage,
		Quota:        quota,
		UsagePercent: usagePercent,
		StorageType:  storageType,
	}
}
