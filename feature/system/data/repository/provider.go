package repository

import "ncobase/feature/system/data"

// Repository represents the system repository.
type Repository struct {
	Menu       MenuRepositoryInterface
	Dictionary DictionaryRepositoryInterface
}

// NewRepository creates a new repository.
func NewRepository(d *data.Data) *Repository {
	return &Repository{
		Menu:       NewMenuRepository(d),
		Dictionary: NewDictionaryRepository(d),
	}
}
