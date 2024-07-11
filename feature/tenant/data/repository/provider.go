package repository

import (
	"ncobase/feature/tenant/data"
)

// Repository represents the tenant repository.
type Repository struct {
	tenant     TenantRepositoryInterface
	userTenant UserTenantRepositoryInterface
}

// NewRepository creates a new repository.
func NewRepository(d *data.Data) *Repository {
	return &Repository{
		tenant:     NewTenantRepository(d),
		userTenant: NewUserTenantRepository(d),
	}
}
