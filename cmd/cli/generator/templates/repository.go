package templates

import "fmt"

func RepositoryTemplate(name, extType string) string {
	return fmt.Sprintf(`package repository

import "ncobase/%s/%s/data"

// Repository represents the %s repository.
type Repository struct {
	// Add your repository fields here
}

// New creates a new repository.
func New(d *data.Data) *Repository {
	return &Repository{
		// Initialize your repository fields here
	}
}

// Add your repository methods here
`, extType, name, name)
}
