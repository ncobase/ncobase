package structs

// SpaceOption represents the space option relationship.
type SpaceOption struct {
	SpaceID  string `json:"space_id,omitempty"`
	OptionID string `json:"option_id,omitempty"`
}

// AddOptionsToSpaceRequest represents the request to add options to a space
type AddOptionsToSpaceRequest struct {
	OptionID string `json:"option_id" binding:"required"`
}
