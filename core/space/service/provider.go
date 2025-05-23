package service

import "ncobase/space/data"

// Service represents the space service.
type Service struct {
	Group     GroupServiceInterface
	GroupRole GroupRoleServiceInterface
	UserGroup UserGroupServiceInterface
}

// New creates a new service.
func New(d *data.Data, userService any) *Service {
	gs := NewGroupService(d)
	grs := NewGroupRoleService(d)
	ugs := NewUserGroupService(d, gs)

	if ugs, ok := ugs.(*userGroupService); ok {
		if userService != nil {
			ugs.SetUserService(NewUserServiceWrapper(userService), NewUserProfileServiceWrapper(userService))
		}
	}

	return &Service{
		Group:     gs,
		GroupRole: grs,
		UserGroup: ugs,
	}
}
