package repository

import (
	"ncobase/feature/asset/data"
	"ncobase/feature/asset/data/repository/asset"
)

type Repository struct {
	Asset asset.RepositoryInterface
}

func New(d *data.Data) *Repository {
	return &Repository{
		Asset: asset.NewAsset(d),
	}
}
