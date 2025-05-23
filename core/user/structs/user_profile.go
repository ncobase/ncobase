package structs

import "github.com/ncobase/ncore/types"

// UserProfileBody represents the user profile schema.
type UserProfileBody struct {
	UserID      string        `json:"userid,omitempty"`
	DisplayName string        `json:"display_name,omitempty"`
	FirstName   string        `json:"first_name,omitempty"`
	LastName    string        `json:"last_name,omitempty"`
	Title       string        `json:"title,omitempty"`
	ShortBio    string        `json:"short_bio,omitempty"`
	About       *string       `json:"about,omitempty"`
	Thumbnail   *string       `json:"thumbnail,omitempty"`
	Links       *[]types.JSON `json:"links,omitempty"`
	Extras      *types.JSON   `json:"extras,omitempty"`
}

// ReadUserProfile represents the user profile schema.
type ReadUserProfile struct {
	UserID      string        `json:"userid,omitempty"`
	DisplayName string        `json:"display_name,omitempty"`
	FirstName   string        `json:"first_name,omitempty"`
	LastName    string        `json:"last_name,omitempty"`
	Title       string        `json:"title,omitempty"`
	ShortBio    string        `json:"short_bio,omitempty"`
	About       *string       `json:"about,omitempty"`
	Thumbnail   *string       `json:"thumbnail,omitempty"`
	Links       *[]types.JSON `json:"links,omitempty"`
	Extras      *types.JSON   `json:"extras,omitempty"`
}
