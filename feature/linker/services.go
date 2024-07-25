package linker

import (
	"fmt"
	accessService "ncobase/feature/access/service"
	authService "ncobase/feature/auth/service"
	groupService "ncobase/feature/group/service"
	tenantService "ncobase/feature/tenant/service"
	userService "ncobase/feature/user/service"
)

// GetUserService returns the user service
func (l *Linker) getUserService() (*userService.Service, error) {
	f, err := l.fm.GetService("user")
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
func (l *Linker) getTenantService() (*tenantService.Service, error) {
	f, err := l.fm.GetService("tenant")
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
func (l *Linker) getGroupService() (*groupService.Service, error) {
	f, err := l.fm.GetService("group")
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
func (l *Linker) getAccessService() (*accessService.Service, error) {
	f, err := l.fm.GetService("access")
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
func (l *Linker) getAuthService() (*authService.Service, error) {
	f, err := l.fm.GetService("auth")
	if err != nil {
		return nil, fmt.Errorf("failed to get access service: %v", err)
	}
	as, ok := f.(*authService.Service)
	if !ok {
		return nil, fmt.Errorf("access service does not implement")
	}
	return as, nil
}
