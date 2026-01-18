package proxy

import (
	"fmt"
	accessService "ncobase/core/access/service"
	orgService "ncobase/core/organization/service"
	spaceService "ncobase/core/space/service"
	userService "ncobase/core/user/service"
)

// GetUserService returns the user service
func (p *Plugin) getUserService() (*userService.Service, error) {
	f, err := p.em.GetServiceByName("user")
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
func (p *Plugin) getSpaceService() (*spaceService.Service, error) {
	f, err := p.em.GetServiceByName("space")
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
func (p *Plugin) getOrganizationService() (*orgService.Service, error) {
	f, err := p.em.GetServiceByName("organization")
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
func (p *Plugin) getAccessService() (*accessService.Service, error) {
	f, err := p.em.GetServiceByName("access")
	if err != nil {
		return nil, fmt.Errorf("failed to get access service: %v", err)
	}
	as, ok := f.(*accessService.Service)
	if !ok {
		return nil, fmt.Errorf("access service does not implement")
	}
	return as, nil
}
