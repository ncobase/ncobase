package repository

import "ncobase/feature/resource/data"

// Repository represents the resource repository.
type Repository struct {
	Asset AssetRepositoryInterface
}

// NewRepository creates a new repository.
func NewRepository(d *data.Data) *Repository {
	return &Repository{
		Asset: NewAssetRepository(d),
	}
}
