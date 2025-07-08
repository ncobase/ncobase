package structs

import "github.com/ncobase/ncore/types"

// BatchUploadParams for batch uploads
type BatchUploadParams struct {
	OwnerID           string             `json:"owner_id" binding:"required"`
	PathPrefix        string             `json:"path_prefix,omitempty"`
	AccessLevel       AccessLevel        `json:"access_level,omitempty"`
	Tags              []string           `json:"tags,omitempty"`
	ProcessingOptions *ProcessingOptions `json:"processing_options,omitempty"`
	Extras            *types.JSON        `json:"extras,omitempty"`
}

// BatchUploadResult for batch upload results
type BatchUploadResult struct {
	OperationID  string      `json:"operation_id"`
	TotalFiles   int         `json:"total_files"`
	SuccessCount int         `json:"success_count"`
	FailureCount int         `json:"failure_count"`
	Files        []*ReadFile `json:"files"`
	FailedFiles  []string    `json:"failed_files,omitempty"`
	Errors       []string    `json:"errors,omitempty"`
}

// BatchDeleteResult for batch delete results
type BatchDeleteResult struct {
	OperationID  string   `json:"operation_id"`
	TotalFiles   int      `json:"total_files"`
	SuccessCount int      `json:"success_count"`
	FailureCount int      `json:"failure_count"`
	DeletedIDs   []string `json:"deleted_ids"`
	FailedIDs    []string `json:"failed_ids,omitempty"`
	Errors       []string `json:"errors,omitempty"`
}

// BatchStatus for batch operation status
type BatchStatus struct {
	OperationID string `json:"operation_id"`
	Status      string `json:"status"`   // pending, processing, completed, failed
	Progress    int    `json:"progress"` // 0-100
	Message     string `json:"message,omitempty"`
	StartedAt   int64  `json:"started_at"`
	CompletedAt *int64 `json:"completed_at,omitempty"`
}
