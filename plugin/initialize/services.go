package initialize

import (
	"fmt"
	accessService "ncobase/access/service"
	authService "ncobase/auth/service"
	orgService "ncobase/organization/service"
	spaceService "ncobase/space/service"
	systemService "ncobase/system/service"
	userService "ncobase/user/service"
)

// GetUserService returns the user service
func (p *Plugin) getUserService() (*userService.Service, error) {
	f, err := p.em.GetServiceByName("user")
	if err != nil {
		return nil, fmt.Errorf("failed to get user service: %v", err)
	}

	svc, ok := f.(*userService.Service)
	if !ok {
		return nil, fmt.Errorf("user service does not implement UserServiceInterface")
	}

	return svc, nil
}

// GetSpaceService returns the space service
func (p *Plugin) getSpaceService() (*spaceService.Service, error) {
	f, err := p.em.GetServiceByName("space")
	if err != nil {
		return nil, fmt.Errorf("failed to get space service: %v", err)
	}
	svc, ok := f.(*spaceService.Service)
	if !ok {
		return nil, fmt.Errorf("space service does not implement")
	}
	return svc, nil
}

// GetOrganizationService returns the organization service
func (p *Plugin) getOrganizationService() (*orgService.Service, error) {
	f, err := p.em.GetServiceByName("organization")
	if err != nil {
		return nil, fmt.Errorf("failed to get organization service: %v", err)
	}
	svc, ok := f.(*orgService.Service)
	if !ok {
		return nil, fmt.Errorf("organization service does not implement")
	}
	return svc, nil
}

// GetAccessService returns the access service
func (p *Plugin) getAccessService() (*accessService.Service, error) {
	f, err := p.em.GetServiceByName("access")
	if err != nil {
		return nil, fmt.Errorf("failed to get access service: %v", err)
	}
	svc, ok := f.(*accessService.Service)
	if !ok {
		return nil, fmt.Errorf("access service does not implement")
	}
	return svc, nil
}

// GetAuthService returns the auth service
func (p *Plugin) getAuthService() (*authService.Service, error) {
	f, err := p.em.GetServiceByName("auth")
	if err != nil {
		return nil, fmt.Errorf("failed to get access service: %v", err)
	}
	svc, ok := f.(*authService.Service)
	if !ok {
		return nil, fmt.Errorf("access service does not implement")
	}
	return svc, nil
}

// getSystemService returns the system service
func (p *Plugin) getSystemService() (*systemService.Service, error) {
	f, err := p.em.GetServiceByName("system")
	if err != nil {
		return nil, fmt.Errorf("failed to get system service: %v", err)
	}
	svc, ok := f.(*systemService.Service)
	if !ok {
		return nil, fmt.Errorf("system service does not implement")
	}
	return svc, nil
}
