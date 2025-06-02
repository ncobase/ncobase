package repository

import (
	"ncobase/space/data"
)

// Repository represents the space repository.
type Repository struct {
	Group     GroupRepositoryInterface
	GroupRole GroupRoleRepositoryInterface
	UserGroup UserGroupRepositoryInterface
}

// New creates a new space repository.
func New(d *data.Data) *Repository {
	return &Repository{
		Group:     NewGroupRepository(d),
		GroupRole: NewGroupRoleRepository(d),
		UserGroup: NewUserGroupRepository(d),
	}
}
