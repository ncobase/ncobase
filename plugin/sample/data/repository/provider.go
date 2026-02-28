package repository

// Repository aggregates all repositories for the sample plugin
type Repository struct {
	Sample SampleRepositoryInterface
}

// NewRepository creates a new repository instance
func NewRepository() *Repository {
	return &Repository{
		Sample: NewSampleRepository(),
	}
}
