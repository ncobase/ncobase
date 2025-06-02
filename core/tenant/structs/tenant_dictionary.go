package structs

// TenantDictionary represents the tenant dictionary relationship.
type TenantDictionary struct {
	TenantID     string `json:"tenant_id,omitempty"`
	DictionaryID string `json:"dictionary_id,omitempty"`
}

// AddDictionaryToTenantRequest represents the request to add a dictionary to a tenant
type AddDictionaryToTenantRequest struct {
	DictionaryID string `json:"dictionary_id" binding:"required"`
}
