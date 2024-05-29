package structs

// Sample represents a sample message body.
type Sample struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

type FindSample struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListSamples - get list of samples
type ListSamples struct {
	Cursor string `json:"cursor,omitempty"`
	Limit  int32  `json:"limit,omitempty"`
}
