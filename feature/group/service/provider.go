package service

import "ncobase/feature/group/data"

// Service represents the group service.
type Service struct {
	Group     GroupServiceInterface
	GroupRole GroupRoleServiceInterface
	UserGroup UserGroupServiceInterface
}

// New creates a new service.
func New(d *data.Data) *Service {
	gs := NewGroupService(d)
	return &Service{
		Group:     gs,
		GroupRole: NewGroupRoleService(d),
		UserGroup: NewUserGroupService(d, gs),
	}
}
