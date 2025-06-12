package structs

// SpaceMenu represents the space menu relationship.
type SpaceMenu struct {
	SpaceID string `json:"space_id,omitempty"`
	MenuID  string `json:"menu_id,omitempty"`
}

// AddMenuToSpaceRequest represents the request to add a menu to a space
type AddMenuToSpaceRequest struct {
	MenuID string `json:"menu_id" binding:"required"`
}
