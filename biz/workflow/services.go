package workflow

import (
	"fmt"
	accessService "ncobase/access/service"
	authService "ncobase/auth/service"
	orgService "ncobase/organization/service"
	spaceService "ncobase/space/service"
	userService "ncobase/user/service"
)

// GetUserService returns the user service
func (m *Module) getUserService() (*userService.Service, error) {
	f, err := m.em.GetServiceByName("user")
	if err != nil {
		return nil, fmt.Errorf("failed to get user service: %v", err)
	}

	us, ok := f.(*userService.Service)
	if !ok {
		return nil, fmt.Errorf("user service does not implement UserServiceInterface")
	}

	return us, nil
}

// GetSpaceService returns the space service
func (m *Module) getSpaceService() (*spaceService.Service, error) {
	f, err := m.em.GetServiceByName("space")
	if err != nil {
		return nil, fmt.Errorf("failed to get space service: %v", err)
	}
	ts, ok := f.(*spaceService.Service)
	if !ok {
		return nil, fmt.Errorf("space service does not implement")
	}
	return ts, nil
}

// GetOrganizationService returns the organization service
func (m *Module) getOrganizationService() (*orgService.Service, error) {
	f, err := m.em.GetServiceByName("organization")
	if err != nil {
		return nil, fmt.Errorf("failed to get organization service: %v", err)
	}
	ss, ok := f.(*orgService.Service)
	if !ok {
		return nil, fmt.Errorf("organization service does not implement")
	}
	return ss, nil
}

// GetAccessService returns the access service
func (m *Module) getAccessService() (*accessService.Service, error) {
	f, err := m.em.GetServiceByName("access")
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
	f, err := m.em.GetServiceByName("auth")
	if err != nil {
		return nil, fmt.Errorf("failed to get access service: %v", err)
	}
	as, ok := f.(*authService.Service)
	if !ok {
		return nil, fmt.Errorf("access service does not implement")
	}
	return as, nil
}
