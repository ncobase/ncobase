package structs

import "github.com/ncobase/ncore/types"

// BatchUploadParams represents parameters for batch upload operations
type BatchUploadParams struct {
	TenantID          string             `json:"tenant_id" binding:"required"`
	ObjectID          string             `json:"object_id" binding:"required"`
	FolderPath        string             `json:"folder_path,omitempty"`
	AccessLevel       AccessLevel        `json:"access_level,omitempty"`
	Tags              []string           `json:"tags,omitempty"`
	ProcessingOptions *ProcessingOptions `json:"processing_options,omitempty"`
	Extras            *types.JSON        `json:"extras,omitempty"`
}

// BatchUploadResult represents the result of a batch upload operation
type BatchUploadResult struct {
	OperationID  string      `json:"operation_id"`
	TotalFiles   int         `json:"total_files"`
	SuccessCount int         `json:"success_count"`
	FailureCount int         `json:"failure_count"`
	Files        []*ReadFile `json:"files"`
	FailedFiles  []string    `json:"failed_files,omitempty"`
	Errors       []string    `json:"errors,omitempty"`
}
