package structs

import "ncobase/common/types"

// UserProfileBody represents the user profile schema.
type UserProfileBody struct {
	ID          string        `json:"id,omitempty"`
	DisplayName string        `json:"display_name,omitempty"`
	ShortBio    string        `json:"short_bio,omitempty"`
	About       *string       `json:"about,omitempty"`
	Thumbnail   *string       `json:"thumbnail,omitempty"`
	Links       *[]types.JSON `json:"links,omitempty"`
	Extras      *types.JSON   `json:"extras,omitempty"`
}
