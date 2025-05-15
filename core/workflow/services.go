package workflow

import (
	"fmt"
	accessService "ncobase/access/service"
	authService "ncobase/auth/service"
	spaceService "ncobase/space/service"
	tenantService "ncobase/tenant/service"
	userService "ncobase/user/service"
)

// GetUserService returns the user service
func (m *Module) getUserService() (*userService.Service, error) {
	f, err := m.em.GetService("user")
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
	f, err := m.em.GetService("tenant")
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant service: %v", err)
	}
	ts, ok := f.(*tenantService.Service)
	if !ok {
		return nil, fmt.Errorf("tenant service does not implement")
	}
	return ts, nil
}

// GetSpaceService returns the space service
func (m *Module) getSpaceService() (*spaceService.Service, error) {
	f, err := m.em.GetService("space")
	if err != nil {
		return nil, fmt.Errorf("failed to get space service: %v", err)
	}
	ss, ok := f.(*spaceService.Service)
	if !ok {
		return nil, fmt.Errorf("space service does not implement")
	}
	return ss, nil
}

// GetAccessService returns the access service
func (m *Module) getAccessService() (*accessService.Service, error) {
	f, err := m.em.GetService("access")
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
	f, err := m.em.GetService("auth")
	if err != nil {
		return nil, fmt.Errorf("failed to get access service: %v", err)
	}
	as, ok := f.(*authService.Service)
	if !ok {
		return nil, fmt.Errorf("access service does not implement")
	}
	return as, nil
}
