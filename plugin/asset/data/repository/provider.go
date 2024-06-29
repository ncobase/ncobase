package repository

import (
	"ncobase/plugin/asset/data"
	"ncobase/plugin/asset/data/repository/asset"
)

type Repository struct {
	Asset asset.RepositoryInterface
}

func New(d *data.Data) *Repository {
	return &Repository{
		Asset: asset.NewAsset(d),
	}
}
