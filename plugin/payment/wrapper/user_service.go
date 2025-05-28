package wrapper

import (
	"context"
	"fmt"
	userStructs "ncobase/user/structs"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/types"
)

// UserServiceInterface defines user service interface for proxy plugin
type UserServiceInterface interface {
	CreateUser(ctx context.Context, body *userStructs.UserBody) (*userStructs.ReadUser, error)
	GetByID(ctx context.Context, id string) (*userStructs.ReadUser, error)
	FindUser(ctx context.Context, m *userStructs.FindUser) (*userStructs.ReadUser, error)
	UpdateUser(ctx context.Context, user string, updates types.JSON) (*userStructs.ReadUser, error)
	UpdatePassword(ctx context.Context, body *userStructs.UserPassword) error
	VerifyPassword(ctx context.Context, userID string, password string) any
	CountX(ctx context.Context, params *userStructs.ListUserParams) int
}

// UserProfileServiceInterface defines user profile service interface
type UserProfileServiceInterface interface {
	Create(ctx context.Context, body *userStructs.UserProfileBody) (*userStructs.ReadUserProfile, error)
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

// loadServices loads user services using extension manager
func (w *UserServiceWrapper) loadServices() {
	if userSvc, err := w.em.GetCrossService("user", "User"); err == nil {
		if service, ok := userSvc.(UserServiceInterface); ok {
			w.userService = service
		}
	}

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

// CreateUser creates user with fallback
func (w *UserServiceWrapper) CreateUser(ctx context.Context, body *userStructs.UserBody) (*userStructs.ReadUser, error) {
	if w.userService != nil {
		return w.userService.CreateUser(ctx, body)
	}
	return nil, fmt.Errorf("user service not available")
}

// GetUserByID gets user by ID with fallback
func (w *UserServiceWrapper) GetUserByID(ctx context.Context, id string) (*userStructs.ReadUser, error) {
	if w.userService != nil {
		return w.userService.GetByID(ctx, id)
	}
	return nil, fmt.Errorf("user service not available")
}

// FindUser finds user with fallback
func (w *UserServiceWrapper) FindUser(ctx context.Context, m *userStructs.FindUser) (*userStructs.ReadUser, error) {
	if w.userService != nil {
		return w.userService.FindUser(ctx, m)
	}
	return nil, fmt.Errorf("user service not available")
}

// UpdateUser updates user with fallback
func (w *UserServiceWrapper) UpdateUser(ctx context.Context, user string, updates types.JSON) (*userStructs.ReadUser, error) {
	if w.userService != nil {
		return w.userService.UpdateUser(ctx, user, updates)
	}
	return nil, fmt.Errorf("user service not available")
}

// UpdatePassword updates user password with fallback
func (w *UserServiceWrapper) UpdatePassword(ctx context.Context, body *userStructs.UserPassword) error {
	if w.userService != nil {
		return w.userService.UpdatePassword(ctx, body)
	}
	return fmt.Errorf("user service not available")
}

// VerifyPassword verifies user password with fallback
func (w *UserServiceWrapper) VerifyPassword(ctx context.Context, userID string, password string) any {
	if w.userService != nil {
		return w.userService.VerifyPassword(ctx, userID, password)
	}
	return nil
}

// CountX counts users with fallback
func (w *UserServiceWrapper) CountX(ctx context.Context, params *userStructs.ListUserParams) int {
	if w.userService != nil {
		return w.userService.CountX(ctx, params)
	}
	return 0
}

// CreateUserProfile creates user profile with fallback
func (w *UserServiceWrapper) CreateUserProfile(ctx context.Context, body *userStructs.UserProfileBody) (*userStructs.ReadUserProfile, error) {
	if w.profileService != nil {
		return w.profileService.Create(ctx, body)
	}
	return nil, fmt.Errorf("user profile service not available")
}

// GetUserProfile gets user profile with graceful fallback
func (w *UserServiceWrapper) GetUserProfile(ctx context.Context, id string) (*userStructs.ReadUserProfile, error) {
	if w.profileService != nil {
		return w.profileService.Get(ctx, id)
	}
	// Return minimal profile as fallback
	return &userStructs.ReadUserProfile{UserID: id}, nil
}

// HasUserService checks if user service is available
func (w *UserServiceWrapper) HasUserService() bool {
	return w.userService != nil
}

// HasProfileService checks if profile service is available
func (w *UserServiceWrapper) HasProfileService() bool {
	return w.profileService != nil
}
