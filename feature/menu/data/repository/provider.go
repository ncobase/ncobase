package repository

import (
	"ncobase/feature/menu/data"
	"ncobase/feature/menu/data/repository/menu"
)

type Repository struct {
	Menu menu.RepositoryInterface
}

func New(d *data.Data) *Repository {
	return &Repository{
		Menu: menu.NewMenu(d),
	}
}
