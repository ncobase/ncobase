package repository

import (
	"ncobase/feature/resource/data"
	"ncobase/feature/resource/data/repository/asset"
)

type Repository struct {
	Asset asset.RepositoryInterface
}

func New(d *data.Data) *Repository {
	return &Repository{
		Asset: asset.NewAsset(d),
	}
}
