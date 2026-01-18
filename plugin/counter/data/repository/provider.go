package repository

import "ncobase/plugin/counter/data"

// Repository represents the counter repository.
type Repository struct {
	Counter CounterRepositoryInterface
}

// New creates a new repository.
func New(d *data.Data) *Repository {
	return &Repository{
		Counter: NewCounterRepository(d),
	}
}
