package service

import (
	"ncobase/space/data"
	"ncobase/space/wrapper"

	ext "github.com/ncobase/ncore/extension/types"
)

// Service represents the space service
type Service struct {
	Group     GroupServiceInterface
	GroupRole GroupRoleServiceInterface
	UserGroup UserGroupServiceInterface
	usw       *wrapper.UserServiceWrapper
}

// New creates a new service
func New(d *data.Data, em ext.ManagerInterface) *Service {
	gs := NewGroupService(d)
	grs := NewGroupRoleService(d)
	usw := wrapper.NewUserServiceWrapper(em)
	ugs := NewUserGroupService(d, gs, usw)

	return &Service{
		Group:     gs,
		GroupRole: grs,
		UserGroup: ugs,
		usw:       usw,
	}
}

// RefreshDependencies refreshes external service dependencies
func (s *Service) RefreshDependencies() {
	s.usw.RefreshServices()
}
