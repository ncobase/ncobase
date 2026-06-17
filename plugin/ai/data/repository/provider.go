package repository

import "ncobase/plugin/ai/data"

type Repository struct {
	Run RunRepositoryInterface
}

func New(d *data.Data) *Repository {
	return &Repository{
		Run: NewRunRepository(d),
	}
}
