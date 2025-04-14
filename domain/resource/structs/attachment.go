package structs

import (
	"fmt"
	"mime/multipart"

	"github.com/ncobase/ncore/types"
)

// FindAttachment represents the parameters for finding an attachment.
type FindAttachment struct {
	Attachment string `json:"attachment,omitempty"`
	Tenant     string `json:"tenant,omitempty"`
	User       string `json:"user,omitempty"`
}

// AttachmentBody represents the common fields for creating and updating an attachment.
type AttachmentBody struct {
	File      multipart.File `json:"-"` // For internal use only, not to be serialized
	Name      string         `json:"name,omitempty"`
	Path      string         `json:"path,omitempty"`
	Type      string         `json:"type,omitempty"`
	Size      *int           `json:"size,omitempty"`
	Storage   string         `json:"storage,omitempty"`
	Bucket    string         `json:"bucket,omitempty"`
	Endpoint  string         `json:"endpoint,omitempty"`
	ObjectID  string         `json:"object_id,omitempty"`
	TenantID  string         `json:"tenant_id,omitempty"`
	Extras    *types.JSON    `json:"extras,omitempty"`
	CreatedBy *string        `json:"created_by,omitempty"`
	UpdatedBy *string        `json:"updated_by,omitempty"`
}

// CreateAttachmentBody represents the body for creating an attachment.
type CreateAttachmentBody struct {
	AttachmentBody
}

// UpdateAttachmentBody represents the body for updating an attachment.
type UpdateAttachmentBody struct {
	ID string `json:"id"`
	AttachmentBody
}

// ReadAttachment represents the output schema for retrieving an attachment.
type ReadAttachment struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Path      string      `json:"path"`
	Type      string      `json:"type"`
	Size      *int        `json:"size"`
	Storage   string      `json:"storage"`
	Bucket    string      `json:"bucket"`
	Endpoint  string      `json:"endpoint"`
	ObjectID  string      `json:"object_id"`
	TenantID  string      `json:"tenant_id"`
	Extras    *types.JSON `json:"extras,omitempty"`
	CreatedBy *string     `json:"created_by,omitempty"`
	CreatedAt *int64      `json:"created_at,omitempty"`
	UpdatedBy *string     `json:"updated_by,omitempty"`
	UpdatedAt *int64      `json:"updated_at,omitempty"`
}

// GetCursorValue returns the cursor value.
func (r *ReadAttachment) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, types.ToValue(r.CreatedAt))
}

// ListAttachmentParams represents the parameters for listing attachments.
type ListAttachmentParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	Tenant    string `form:"tenant,omitempty" json:"tenant,omitempty" validate:"required"`
	Object    string `form:"object,omitempty" json:"object,omitempty" validate:"required"`
	User      string `form:"user,omitempty" json:"user,omitempty"`
	Type      string `form:"type,omitempty" json:"type,omitempty"`
	Storage   string `form:"storage,omitempty" json:"storage,omitempty"`
}
