package relation

import (
	"fmt"
	accessService "ncobase/core/access/service"
	authService "ncobase/core/auth/service"
	groupService "ncobase/core/group/service"
	tenantService "ncobase/core/tenant/service"
	userService "ncobase/core/user/service"
)

// GetUserService returns the user service
func (m *Module) getUserService() (*userService.Service, error) {
	f, err := m.fm.GetService("user")
	if err != nil {
		return nil, fmt.Errorf("failed to get user service: %v", err)
	}

	us, ok := f.(*userService.Service)
	if !ok {
		return nil, fmt.Errorf("user service does not implement UserServiceInterface")
	}

	return us, nil
}

// GetTenantService returns the tenant service
func (m *Module) getTenantService() (*tenantService.Service, error) {
	f, err := m.fm.GetService("tenant")
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant service: %v", err)
	}
	ts, ok := f.(*tenantService.Service)
	if !ok {
		return nil, fmt.Errorf("tenant service does not implement")
	}
	return ts, nil
}

// GetGroupService returns the group service
func (m *Module) getGroupService() (*groupService.Service, error) {
	f, err := m.fm.GetService("group")
	if err != nil {
		return nil, fmt.Errorf("failed to get group service: %v", err)
	}
	gs, ok := f.(*groupService.Service)
	if !ok {
		return nil, fmt.Errorf("group service does not implement")
	}
	return gs, nil
}

// GetAccessService returns the access service
func (m *Module) getAccessService() (*accessService.Service, error) {
	f, err := m.fm.GetService("access")
	if err != nil {
		return nil, fmt.Errorf("failed to get access service: %v", err)
	}
	as, ok := f.(*accessService.Service)
	if !ok {
		return nil, fmt.Errorf("access service does not implement")
	}
	return as, nil
}

// GetAuthService returns the auth service
func (m *Module) getAuthService() (*authService.Service, error) {
	f, err := m.fm.GetService("auth")
	if err != nil {
		return nil, fmt.Errorf("failed to get access service: %v", err)
	}
	as, ok := f.(*authService.Service)
	if !ok {
		return nil, fmt.Errorf("access service does not implement")
	}
	return as, nil
}
