package structs

import "encoding/json"

type MessageInput struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ResponseFormatInput struct {
	Type        string         `json:"type,omitempty"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Schema      map[string]any `json:"schema,omitempty"`
	Strict      bool           `json:"strict,omitempty"`
}

type ReasoningInput struct {
	Effort          string `json:"effort,omitempty"`
	BudgetTokens    int    `json:"budget_tokens,omitempty"`
	IncludeThoughts bool   `json:"include_thoughts,omitempty"`
}

type CompleteRequest struct {
	Prompt          string               `json:"prompt,omitempty"`
	Messages        []MessageInput       `json:"messages,omitempty"`
	Model           string               `json:"model,omitempty"`
	MaxOutputTokens int                  `json:"max_output_tokens,omitempty"`
	Temperature     *float64             `json:"temperature,omitempty"`
	TopP            *float64             `json:"top_p,omitempty"`
	Stop            []string             `json:"stop,omitempty"`
	ResponseFormat  *ResponseFormatInput `json:"response_format,omitempty"`
	Reasoning       *ReasoningInput      `json:"reasoning,omitempty"`
	Metadata        map[string]string    `json:"metadata,omitempty"`
	Action          string               `json:"action,omitempty"`
}

type CompleteResponse struct {
	RunID           string          `json:"run_id"`
	OperationID     string          `json:"operation_id,omitempty"`
	Content         string          `json:"content"`
	Reasoning       string          `json:"reasoning,omitempty"`
	Provider        string          `json:"provider"`
	Model           string          `json:"model"`
	FinishReason    string          `json:"finish_reason,omitempty"`
	InputTokens     int             `json:"input_tokens"`
	OutputTokens    int             `json:"output_tokens"`
	TotalTokens     int             `json:"total_tokens"`
	ReasoningTokens int             `json:"reasoning_tokens"`
	CacheUsage      CacheUsage      `json:"cache_usage"`
	ToolCalls       []ToolCall      `json:"tool_calls,omitempty"`
	Raw             json.RawMessage `json:"raw,omitempty"`
}

type CacheUsage struct {
	CreatedTokens int `json:"created_tokens"`
	ReadTokens    int `json:"read_tokens"`
}

type ToolCall struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Name      string          `json:"name,omitempty"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

type StreamStartResponse struct {
	RunID       string `json:"run_id"`
	OperationID string `json:"operation_id,omitempty"`
}

type StreamChunkResponse struct {
	RunID           string     `json:"run_id"`
	Content         string     `json:"content,omitempty"`
	Reasoning       string     `json:"reasoning,omitempty"`
	Done            bool       `json:"done"`
	FinishReason    string     `json:"finish_reason,omitempty"`
	InputTokens     int        `json:"input_tokens,omitempty"`
	OutputTokens    int        `json:"output_tokens,omitempty"`
	TotalTokens     int        `json:"total_tokens,omitempty"`
	ReasoningTokens int        `json:"reasoning_tokens,omitempty"`
	CacheUsage      CacheUsage `json:"cache_usage,omitempty"`
	Error           string     `json:"error,omitempty"`
}

type EmbedRequest struct {
	Input     []string          `json:"input"`
	Model     string            `json:"model,omitempty"`
	InputType string            `json:"input_type,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type EmbedResponse struct {
	RunID      string      `json:"run_id"`
	Provider   string      `json:"provider,omitempty"`
	Model      string      `json:"model"`
	Embeddings [][]float64 `json:"embeddings"`
	TokensUsed int         `json:"tokens_used"`
}

type ActionRequest struct {
	Content         string         `json:"content,omitempty"`
	Instruction     string         `json:"instruction,omitempty"`
	Context         map[string]any `json:"context,omitempty"`
	Language        string         `json:"language,omitempty"`
	Tone            string         `json:"tone,omitempty"`
	OutputFormat    string         `json:"output_format,omitempty"`
	Model           string         `json:"model,omitempty"`
	MaxOutputTokens int            `json:"max_output_tokens,omitempty"`
	Temperature     *float64       `json:"temperature,omitempty"`
}

type ActionDescriptor struct {
	Key         string   `json:"key"`
	Domain      string   `json:"domain"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	OutputType  string   `json:"output_type"`
	Permissions []string `json:"permissions,omitempty"`
}

type ActionResponse struct {
	RunID       string         `json:"run_id"`
	Action      string         `json:"action"`
	Content     string         `json:"content"`
	JSON        map[string]any `json:"json,omitempty"`
	Provider    string         `json:"provider"`
	Model       string         `json:"model"`
	TotalTokens int            `json:"total_tokens"`
}
