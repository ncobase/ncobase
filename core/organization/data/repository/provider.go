package repository

import (
	"ncobase/organization/data"
)

// Repository represents the organization repository.
type Repository struct {
	Organization     OrganizationRepositoryInterface
	OrganizationRole OrganizationRoleRepositoryInterface
	UserOrganization UserOrganizationRepositoryInterface
}

// New creates a new organization repository.
func New(d *data.Data) *Repository {
	return &Repository{
		Organization:     NewOrganizationRepository(d),
		OrganizationRole: NewOrganizationRoleRepository(d),
		UserOrganization: NewUserOrganizationRepository(d),
	}
}
