package service

import (
	"ncobase/core/organization/data"
	"ncobase/core/organization/wrapper"

	ext "github.com/ncobase/ncore/extension/types"
)

// Service represents the organization service
type Service struct {
	Organization     OrganizationServiceInterface
	OrganizationRole OrganizationRoleServiceInterface
	UserOrganization UserOrganizationServiceInterface
	usw              *wrapper.UserServiceWrapper
}

// New creates a new service
func New(d *data.Data, em ext.ManagerInterface) *Service {
	os := NewOrganizationService(d)
	ors := NewOrganizationRoleService(d)
	usw := wrapper.NewUserServiceWrapper(em)
	uos := NewUserOrganizationService(d, os, usw)

	return &Service{
		Organization:     os,
		OrganizationRole: ors,
		UserOrganization: uos,
		usw:              usw,
	}
}

// RefreshDependencies refreshes external service dependencies
func (s *Service) RefreshDependencies() {
	s.usw.RefreshServices()
}
