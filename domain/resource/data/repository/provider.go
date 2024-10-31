package repository

import "ncobase/domain/resource/data"

// Repository represents the resource repository.
type Repository struct {
	Attachment AttachmentRepositoryInterface
}

// New creates a new repository.
func New(d *data.Data) *Repository {
	return &Repository{
		Attachment: NewAttachmentRepository(d),
	}
}
