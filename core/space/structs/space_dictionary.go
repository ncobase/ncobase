package structs

// SpaceDictionary represents the space dictionary relationship.
type SpaceDictionary struct {
	SpaceID      string `json:"space_id,omitempty"`
	DictionaryID string `json:"dictionary_id,omitempty"`
}

// AddDictionaryToSpaceRequest represents the request to add a dictionary to a space
type AddDictionaryToSpaceRequest struct {
	DictionaryID string `json:"dictionary_id" binding:"required"`
}
