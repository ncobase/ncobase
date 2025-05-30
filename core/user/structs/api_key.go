package structs

// ApiKey represents an API key
type ApiKey struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Key       string `json:"key"`
	UserID    string `json:"user_id"`
	CreatedAt int64  `json:"created_at"`
	LastUsed  *int64 `json:"last_used,omitempty"`
}

// CreateApiKeyRequest represents a request to create an API key
type CreateApiKeyRequest struct {
	Name string `json:"name" validate:"required"`
}
