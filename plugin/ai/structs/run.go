package structs

import "fmt"

type RunMode string

const (
	RunModeComplete RunMode = "complete"
	RunModeStream   RunMode = "stream"
	RunModeEmbed    RunMode = "embed"
	RunModeAction   RunMode = "action"
)

type RunStatus string

const (
	RunStatusRunning   RunStatus = "running"
	RunStatusSucceeded RunStatus = "succeeded"
	RunStatusFailed    RunStatus = "failed"
	RunStatusCanceled  RunStatus = "canceled"
)

type Run struct {
	ID                 string         `json:"id"`
	OperationID        string         `json:"operation_id,omitempty"`
	Action             string         `json:"action,omitempty"`
	Mode               RunMode        `json:"mode"`
	Status             RunStatus      `json:"status"`
	Provider           string         `json:"provider,omitempty"`
	Model              string         `json:"model,omitempty"`
	FallbackModel      string         `json:"fallback_model,omitempty"`
	InputTokens        int            `json:"input_tokens"`
	OutputTokens       int            `json:"output_tokens"`
	TotalTokens        int            `json:"total_tokens"`
	ReasoningTokens    int            `json:"reasoning_tokens"`
	CacheCreatedTokens int            `json:"cache_created_tokens"`
	CacheReadTokens    int            `json:"cache_read_tokens"`
	DurationMS         int64          `json:"duration_ms"`
	ErrorCode          string         `json:"error_code,omitempty"`
	ErrorMessage       string         `json:"error_message,omitempty"`
	RequestHash        string         `json:"request_hash,omitempty"`
	ResponseHash       string         `json:"response_hash,omitempty"`
	EstimatedCost      float64        `json:"estimated_cost,omitempty"`
	Currency           string         `json:"currency,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
	SpaceID            string         `json:"space_id,omitempty"`
	UserID             string         `json:"user_id,omitempty"`
	CreatedBy          string         `json:"created_by,omitempty"`
	UpdatedBy          string         `json:"updated_by,omitempty"`
	CreatedAt          int64          `json:"created_at,omitempty"`
	UpdatedAt          int64          `json:"updated_at,omitempty"`
}

func (r *Run) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, r.CreatedAt)
}

type CreateRunInput struct {
	OperationID string         `json:"operation_id,omitempty"`
	Action      string         `json:"action,omitempty"`
	Mode        RunMode        `json:"mode"`
	Status      RunStatus      `json:"status"`
	Provider    string         `json:"provider,omitempty"`
	Model       string         `json:"model,omitempty"`
	RequestHash string         `json:"request_hash,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	SpaceID     string         `json:"space_id,omitempty"`
	UserID      string         `json:"user_id,omitempty"`
	CreatedBy   string         `json:"created_by,omitempty"`
	UpdatedBy   string         `json:"updated_by,omitempty"`
}

type UpdateRunInput struct {
	ID                 string         `json:"id"`
	OperationID        string         `json:"operation_id,omitempty"`
	Action             string         `json:"action,omitempty"`
	Mode               RunMode        `json:"mode,omitempty"`
	Status             RunStatus      `json:"status,omitempty"`
	Provider           string         `json:"provider,omitempty"`
	Model              string         `json:"model,omitempty"`
	FallbackModel      string         `json:"fallback_model,omitempty"`
	InputTokens        int            `json:"input_tokens,omitempty"`
	OutputTokens       int            `json:"output_tokens,omitempty"`
	TotalTokens        int            `json:"total_tokens,omitempty"`
	ReasoningTokens    int            `json:"reasoning_tokens,omitempty"`
	CacheCreatedTokens int            `json:"cache_created_tokens,omitempty"`
	CacheReadTokens    int            `json:"cache_read_tokens,omitempty"`
	DurationMS         int64          `json:"duration_ms,omitempty"`
	ErrorCode          string         `json:"error_code,omitempty"`
	ErrorMessage       string         `json:"error_message,omitempty"`
	RequestHash        string         `json:"request_hash,omitempty"`
	ResponseHash       string         `json:"response_hash,omitempty"`
	EstimatedCost      float64        `json:"estimated_cost,omitempty"`
	Currency           string         `json:"currency,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
	UpdatedBy          string         `json:"updated_by,omitempty"`
}

type RunQuery struct {
	ID        string    `form:"id" json:"id,omitempty"`
	Action    string    `form:"action" json:"action,omitempty"`
	Mode      RunMode   `form:"mode" json:"mode,omitempty"`
	Status    RunStatus `form:"status" json:"status,omitempty"`
	Provider  string    `form:"provider" json:"provider,omitempty"`
	Model     string    `form:"model" json:"model,omitempty"`
	UserID    string    `form:"user_id" json:"user_id,omitempty"`
	SpaceID   string    `form:"space_id" json:"space_id,omitempty"`
	StartDate int64     `form:"start_date" json:"start_date,omitempty"`
	EndDate   int64     `form:"end_date" json:"end_date,omitempty"`
	Cursor    string    `form:"cursor" json:"cursor,omitempty"`
	PageSize  int       `form:"page_size,default=20" json:"page_size,omitempty"`
	Direction string    `form:"direction,default=forward" json:"direction,omitempty"`
}

type UsageQuery struct {
	Action    string `form:"action" json:"action,omitempty"`
	Mode      string `form:"mode" json:"mode,omitempty"`
	Status    string `form:"status" json:"status,omitempty"`
	Provider  string `form:"provider" json:"provider,omitempty"`
	Model     string `form:"model" json:"model,omitempty"`
	UserID    string `form:"user_id" json:"user_id,omitempty"`
	SpaceID   string `form:"space_id" json:"space_id,omitempty"`
	StartDate int64  `form:"start_date" json:"start_date,omitempty"`
	EndDate   int64  `form:"end_date" json:"end_date,omitempty"`
}

type UsageSummary struct {
	TotalRuns          int64                   `json:"total_runs"`
	SucceededRuns      int64                   `json:"succeeded_runs"`
	FailedRuns         int64                   `json:"failed_runs"`
	InputTokens        int64                   `json:"input_tokens"`
	OutputTokens       int64                   `json:"output_tokens"`
	TotalTokens        int64                   `json:"total_tokens"`
	ReasoningTokens    int64                   `json:"reasoning_tokens"`
	CacheCreatedTokens int64                   `json:"cache_created_tokens"`
	CacheReadTokens    int64                   `json:"cache_read_tokens"`
	DurationMS         int64                   `json:"duration_ms"`
	EstimatedCost      float64                 `json:"estimated_cost,omitempty"`
	Currency           string                  `json:"currency,omitempty"`
	ByProvider         map[string]*UsageBucket `json:"by_provider"`
	ByAction           map[string]*UsageBucket `json:"by_action"`
	ByMode             map[string]*UsageBucket `json:"by_mode"`
}

type UsageBucket struct {
	Runs          int64   `json:"runs"`
	SucceededRuns int64   `json:"succeeded_runs"`
	FailedRuns    int64   `json:"failed_runs"`
	TotalTokens   int64   `json:"total_tokens"`
	EstimatedCost float64 `json:"estimated_cost,omitempty"`
}
