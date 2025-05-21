package initialize

import (
	"fmt"
	accessService "ncobase/access/service"
	authService "ncobase/auth/service"
	spaceService "ncobase/space/service"
	systemService "ncobase/system/service"
	tenantService "ncobase/tenant/service"
	userService "ncobase/user/service"
)

// GetUserService returns the user service
func (p *Plugin) getUserService() (*userService.Service, error) {
	f, err := p.em.GetService("user")
	if err != nil {
		return nil, fmt.Errorf("failed to get user service: %v", err)
	}

	svc, ok := f.(*userService.Service)
	if !ok {
		return nil, fmt.Errorf("user service does not implement UserServiceInterface")
	}

	return svc, nil
}

// GetTenantService returns the tenant service
func (p *Plugin) getTenantService() (*tenantService.Service, error) {
	f, err := p.em.GetService("tenant")
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant service: %v", err)
	}
	svc, ok := f.(*tenantService.Service)
	if !ok {
		return nil, fmt.Errorf("tenant service does not implement")
	}
	return svc, nil
}

// GetSpaceService returns the space service
func (p *Plugin) getSpaceService() (*spaceService.Service, error) {
	f, err := p.em.GetService("space")
	if err != nil {
		return nil, fmt.Errorf("failed to get space service: %v", err)
	}
	svc, ok := f.(*spaceService.Service)
	if !ok {
		return nil, fmt.Errorf("space service does not implement")
	}
	return svc, nil
}

// GetAccessService returns the access service
func (p *Plugin) getAccessService() (*accessService.Service, error) {
	f, err := p.em.GetService("access")
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
	f, err := p.em.GetService("auth")
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
	f, err := p.em.GetService("system")
	if err != nil {
		return nil, fmt.Errorf("failed to get system service: %v", err)
	}
	svc, ok := f.(*systemService.Service)
	if !ok {
		return nil, fmt.Errorf("system service does not implement")
	}
	return svc, nil
}
