package repository

import "ncobase/domain/resource/data"

// Repository represents the resource repository.
type Repository struct {
	File FileRepositoryInterface
}

// New creates a new repository.
func New(d *data.Data) *Repository {
	return &Repository{
		File: NewFileRepository(d),
	}
}
