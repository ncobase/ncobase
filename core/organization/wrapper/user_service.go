package wrapper

import (
	"context"
	"fmt"
	userStructs "ncobase/core/user/structs"

	ext "github.com/ncobase/ncore/extension/types"
)

// UserServiceInterface defines user service interface for organization module
type UserServiceInterface interface {
	GetByID(ctx context.Context, id string) (*userStructs.ReadUser, error)
	FindUser(ctx context.Context, m *userStructs.FindUser) (*userStructs.ReadUser, error)
}

// UserProfileServiceInterface defines user profile service interface
type UserProfileServiceInterface interface {
	Get(ctx context.Context, id string) (*userStructs.ReadUserProfile, error)
}

// UserServiceWrapper wraps user service access with fallback behavior
type UserServiceWrapper struct {
	em             ext.ManagerInterface
	userService    UserServiceInterface
	profileService UserProfileServiceInterface
}

// NewUserServiceWrapper creates a new user service wrapper
func NewUserServiceWrapper(em ext.ManagerInterface) *UserServiceWrapper {
	wrapper := &UserServiceWrapper{em: em}
	wrapper.loadServices()
	return wrapper
}

// loadServices loads user services using existing extension manager methods
func (w *UserServiceWrapper) loadServices() {
	// Try to get user service using existing GetCrossService
	if userSvc, err := w.em.GetCrossService("user", "User"); err == nil {
		if service, ok := userSvc.(UserServiceInterface); ok {
			w.userService = service
		}
	}

	// Try to get user profile service
	if profileSvc, err := w.em.GetCrossService("user", "UserProfile"); err == nil {
		if service, ok := profileSvc.(UserProfileServiceInterface); ok {
			w.profileService = service
		}
	}
}

// RefreshServices refreshes service references
func (w *UserServiceWrapper) RefreshServices() {
	w.loadServices()
}

// GetUserByID gets user by ID
func (w *UserServiceWrapper) GetUserByID(ctx context.Context, id string) (*userStructs.ReadUser, error) {
	if w.userService != nil {
		return w.userService.GetByID(ctx, id)
	}
	return nil, fmt.Errorf("user service not available")
}

// FindUser finds user
func (w *UserServiceWrapper) FindUser(ctx context.Context, m *userStructs.FindUser) (*userStructs.ReadUser, error) {
	if w.userService != nil {
		return w.userService.FindUser(ctx, m)
	}
	return nil, fmt.Errorf("user service not available")
}

// GetUserProfile gets user profile
func (w *UserServiceWrapper) GetUserProfile(ctx context.Context, id string) (*userStructs.ReadUserProfile, error) {
	if w.profileService != nil {
		return w.profileService.Get(ctx, id)
	}
	return nil, fmt.Errorf("user profile service not available")
}

// HasUserService checks if user service is available
func (w *UserServiceWrapper) HasUserService() bool {
	return w.userService != nil
}

// HasProfileService checks if profile service is available
func (w *UserServiceWrapper) HasProfileService() bool {
	return w.profileService != nil
}
