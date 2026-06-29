package structs

// DeleteImpactMaxFiles limits one delete-impact request to a bounded, predictable workload.
const DeleteImpactMaxFiles = 100

// DeleteImpactRequest represents a batch delete-impact request.
type DeleteImpactRequest struct {
	IDs []string `json:"ids"`
}

// DeleteImpactResponse represents delete-impact details for one or more files.
type DeleteImpactResponse struct {
	Impacts []*DeleteImpact      `json:"impacts"`
	Summary *DeleteImpactSummary `json:"summary"`
}

// DeleteImpactSummary summarizes whether the requested files can be deleted.
type DeleteImpactSummary struct {
	FileCount           int  `json:"file_count"`
	ReferencedFileCount int  `json:"referenced_file_count"`
	MediaReferenceCount int  `json:"media_reference_count"`
	TopicReferenceCount int  `json:"topic_reference_count"`
	ErrorCount          int  `json:"error_count"`
	CanDelete           bool `json:"can_delete"`
}

// DeleteImpact represents references that would be affected by deleting a file.
type DeleteImpact struct {
	File                    *ReadFile         `json:"file,omitempty"`
	MediaReferences         []*MediaReference `json:"media_references"`
	TopicReferences         []*TopicReference `json:"topic_references"`
	MediaReferenceTotal     int               `json:"media_reference_total"`
	TopicReferenceTotal     int               `json:"topic_reference_total"`
	MediaReferencesComplete bool              `json:"media_references_complete"`
	TopicReferencesComplete bool              `json:"topic_references_complete"`
	Errors                  []string          `json:"errors"`
	CanDelete               bool              `json:"can_delete"`
}

// MediaReference is a content-media reference to a resource file.
type MediaReference struct {
	ID          string         `json:"id"`
	Title       string         `json:"title,omitempty"`
	Type        string         `json:"type,omitempty"`
	ResourceID  string         `json:"resource_id,omitempty"`
	URL         string         `json:"url,omitempty"`
	Path        string         `json:"path,omitempty"`
	MimeType    string         `json:"mime_type,omitempty"`
	Size        *int           `json:"size,omitempty"`
	Description string         `json:"description,omitempty"`
	Alt         string         `json:"alt,omitempty"`
	SpaceID     string         `json:"space_id,omitempty"`
	OwnerID     string         `json:"owner_id,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedBy   *string        `json:"created_by,omitempty"`
	CreatedAt   *int64         `json:"created_at,omitempty"`
	UpdatedBy   *string        `json:"updated_by,omitempty"`
	UpdatedAt   *int64         `json:"updated_at,omitempty"`
}

// TopicMediaReference is a topic-media relation that points at a content media item.
type TopicMediaReference struct {
	ID        string  `json:"id"`
	TopicID   string  `json:"topic_id,omitempty"`
	MediaID   string  `json:"media_id,omitempty"`
	Type      string  `json:"type,omitempty"`
	Order     int     `json:"order,omitempty"`
	CreatedBy *string `json:"created_by,omitempty"`
	CreatedAt *int64  `json:"created_at,omitempty"`
	UpdatedBy *string `json:"updated_by,omitempty"`
	UpdatedAt *int64  `json:"updated_at,omitempty"`
}

// TopicReference is a topic usage of a content media item.
type TopicReference struct {
	Media    *MediaReference      `json:"media,omitempty"`
	Relation *TopicMediaReference `json:"relation,omitempty"`
	Topic    *TopicSummary        `json:"topic,omitempty"`
}

// TopicSummary is a compact topic projection used by resource delete impact responses.
type TopicSummary struct {
	ID            string   `json:"id,omitempty"`
	Name          string   `json:"name,omitempty"`
	Title         string   `json:"title,omitempty"`
	Slug          string   `json:"slug,omitempty"`
	ContentType   string   `json:"content_type,omitempty"`
	Status        int      `json:"status,omitempty"`
	FeaturedMedia string   `json:"featured_media,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	SpaceID       string   `json:"space_id,omitempty"`
	CreatedBy     *string  `json:"created_by,omitempty"`
	CreatedAt     *int64   `json:"created_at,omitempty"`
	UpdatedBy     *string  `json:"updated_by,omitempty"`
	UpdatedAt     *int64   `json:"updated_at,omitempty"`
}
