package structs

// UserSpace represents the user space.
type UserSpace struct {
	UserID  string `json:"user_id,omitempty"`
	SpaceID string `json:"space_id,omitempty"`
}
