package structs

import (
	"mime/multipart"
	"ncobase/pkg/types"
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
	Size     *int           `json:"size,omitempty"`
	Storage  string         `json:"storage,omitempty"`
	Bucket   string         `json:"bucket,omitempty"`
	Endpoint string         `json:"endpoint,omitempty"`
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
	Size     *int        `json:"size"`
	Storage  string      `json:"storage"`
	Bucket   string      `json:"bucket"`
	Endpoint string      `json:"endpoint"`
	ObjectID string      `json:"object_id"`
	DomainID string      `json:"domain_id"`
	Extras   *types.JSON `json:"extras,omitempty"`
	BaseEntity
}

// ListAssetParams represents the parameters for listing assets.
type ListAssetParams struct {
	Cursor   string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit    int    `form:"limit,omitempty" json:"limit,omitempty"` // validate:"gte=1,lte=100"
	DomainID string `form:"domain_id,omitempty" json:"domain_id,omitempty" validate:"required"`
	ObjectID string `form:"object_id,omitempty" json:"object_id,omitempty" validate:"required"`
	UserID   string `form:"user_id,omitempty" json:"user_id,omitempty"`
	Type     string `form:"type,omitempty" json:"type,omitempty"`
	Storage  string `form:"storage,omitempty" json:"storage,omitempty"`
}
