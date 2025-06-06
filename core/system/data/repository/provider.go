package repository

import "ncobase/system/data"

// Repository represents the system repository.
type Repository struct {
	Menu       MenuRepositoryInterface
	Dictionary DictionaryRepositoryInterface
	Options    OptionRepositoryInterface
}

// New creates a new repository.
func New(d *data.Data) *Repository {
	return &Repository{
		Menu:       NewMenuRepository(d),
		Dictionary: NewDictionaryRepository(d),
		Options:    NewOptionRepository(d),
	}
}
