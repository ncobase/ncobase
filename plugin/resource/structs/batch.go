package structs

import "github.com/ncobase/ncore/types"

// BatchUploadParams for batch uploads
type BatchUploadParams struct {
	SpaceID           string             `json:"space_id" binding:"required"`
	OwnerID           string             `json:"owner_id" binding:"required"`
	FolderPath        string             `json:"folder_path,omitempty"`
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
