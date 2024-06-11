package structs

import "stocms/pkg/types"

// UserProfile represents the user profile schema.
type UserProfile struct {
	ID          string        `json:"id,omitempty"`
	DisplayName string        `json:"display_name,omitempty"`
	ShortBio    string        `json:"short_bio,omitempty"`
	About       *string       `json:"about,omitempty"`
	Thumbnail   *string       `json:"thumbnail,omitempty"`
	Links       *[]types.JSON `json:"links,omitempty"`
}
