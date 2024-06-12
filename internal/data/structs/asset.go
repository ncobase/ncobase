package structs

import (
	"mime/multipart"
	"stocms/pkg/types"
)

// FindAsset represents the parameters for finding an asset.
type FindAsset struct {
	ID       string `json:"id,omitempty"`
	DomainID string `json:"domain_id,omitempty"`
	UserID   string `json:"user_id,omitempty"`
}

// AssetBody represents the common fields for creating and updating an asset.
type AssetBody struct {
	File     multipart.File `json:"-"` // For internal use only, not to be serialized
	Name     string         `json:"name,omitempty"`
	Path     string         `json:"path,omitempty"`
	Type     string         `json:"type,omitempty"`
	Size     *int64         `json:"size,omitempty"`
	Storage  string         `json:"storage,omitempty"`
	URL      string         `json:"url,omitempty"`
	ObjectID string         `json:"object_id,omitempty"`
	DomainID string         `json:"domain_id,omitempty"`
	Extras   *types.JSON    `json:"extras,omitempty"`
	BaseEntity
}

// CreateAssetBody represents the body for creating an asset.
type CreateAssetBody struct {
	AssetBody
}

// UpdateAssetBody represents the body for updating an asset.
type UpdateAssetBody struct {
	ID string `json:"id"`
	AssetBody
}

// ReadAsset represents the output schema for retrieving an asset.
type ReadAsset struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	Type     string      `json:"type"`
	Size     *int64      `json:"size"`
	Storage  string      `json:"storage"`
	URL      string      `json:"url"`
	ObjectID string      `json:"object_id"`
	DomainID string      `json:"domain_id"`
	Extras   *types.JSON `json:"extras,omitempty"`
	BaseEntity
}

// ListAssetParams represents the parameters for listing assets.
type ListAssetParams struct {
	Cursor   string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit    int64  `form:"limit,omitempty" json:"limit,omitempty"`
	DomainID string `form:"domain_id,omitempty" json:"domain_id,omitempty"`
	ObjectID string `form:"object_id,omitempty" json:"object_id,omitempty"`
	UserID   string `form:"user_id,omitempty" json:"user_id,omitempty"`
	Type     string `form:"type,omitempty" json:"type,omitempty"`
	Storage  string `form:"storage,omitempty" json:"storage,omitempty"`
}

// Validate validates ListAssetParams
func (p *ListAssetParams) Validate() error {
	return validate.Struct(p)
}
