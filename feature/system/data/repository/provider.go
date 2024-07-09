package repository

import (
	"ncobase/feature/system/data"
	"ncobase/feature/system/data/repository/menu"
)

type Repository struct {
	Menu menu.RepositoryInterface
}

func New(d *data.Data) *Repository {
	return &Repository{
		Menu: menu.NewMenu(d),
	}
}
