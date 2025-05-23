package service

import (
	"context"
	"fmt"
	userStructs "ncobase/user/structs"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
)

// wrapper defines the wrapper interface
type wrapper interface {
	GetService() any
}

// serviceWrapper wraps the actual service implementation
type serviceWrapper struct {
	service any
}

// GetService returns the wrapped service
func (w *serviceWrapper) GetService() any {
	return w.service
}

// UserServiceInterface defines the user service interface needed by visualization module
type UserServiceInterface interface {
	GetByID(ctx context.Context, id string) (*userStructs.ReadUser, error)
	FindUser(ctx context.Context, m *userStructs.FindUser) (*userStructs.ReadUser, error)
	UpdateUser(ctx context.Context, user string, updates types.JSON) (*userStructs.ReadUser, error)
}

// UserServiceWrapper wraps the actual user service implementation
type UserServiceWrapper struct {
	serviceWrapper
}

var _ UserServiceInterface = (*UserServiceWrapper)(nil)
var _ wrapper = (*UserServiceWrapper)(nil)

func NewUserServiceWrapper(service any) *UserServiceWrapper {
	return &UserServiceWrapper{
		serviceWrapper: serviceWrapper{service: service},
	}
}

func (w *UserServiceWrapper) GetByID(ctx context.Context, id string) (*userStructs.ReadUser, error) {
	logger.Infof(ctx, "UserServiceWrapper.GetByID called with service type: %T", w.service)

	// Try to access through field
	if userField, ok := ext.GetServiceInterface(w.service, "User"); ok {
		logger.Infof(ctx, "Found User field of type: %T", userField)
		if user, ok := userField.(interface {
			GetByID(ctx context.Context, id string) (*userStructs.ReadUser, error)
		}); ok {
			return user.GetByID(ctx, id)
		}
	}

	// Try to access method directly
	if s, ok := w.service.(interface {
		GetByID(ctx context.Context, id string) (*userStructs.ReadUser, error)
	}); ok {
		return s.GetByID(ctx, id)
	}

	logger.Errorf(ctx, "UserServiceWrapper: service type %T does not implement required method", w.service)
	return nil, fmt.Errorf("user service does not implement GetByID method")
}

func (w *UserServiceWrapper) FindUser(ctx context.Context, m *userStructs.FindUser) (*userStructs.ReadUser, error) {
	logger.Infof(ctx, "UserServiceWrapper.FindUser called with service type: %T", w.service)

	// Try to access through field
	if userField, ok := ext.GetServiceInterface(w.service, "User"); ok {
		logger.Infof(ctx, "Found User field of type: %T", userField)
		if user, ok := userField.(interface {
			FindUser(ctx context.Context, m *userStructs.FindUser) (*userStructs.ReadUser, error)
		}); ok {
			return user.FindUser(ctx, m)
		}
	}

	// Try to access method directly
	if s, ok := w.service.(interface {
		FindUser(ctx context.Context, m *userStructs.FindUser) (*userStructs.ReadUser, error)
	}); ok {
		return s.FindUser(ctx, m)
	}

	logger.Errorf(ctx, "UserServiceWrapper: service type %T does not implement required method", w.service)
	return nil, fmt.Errorf("user service does not implement FindUser method")
}

func (w *UserServiceWrapper) UpdateUser(ctx context.Context, user string, updates types.JSON) (*userStructs.ReadUser, error) {
	logger.Infof(ctx, "UserServiceWrapper.UpdateUser called with service type: %T", w.service)

	// Try to access through field
	if userField, ok := ext.GetServiceInterface(w.service, "User"); ok {
		logger.Infof(ctx, "Found User field of type: %T", userField)
		if u, ok := userField.(interface {
			UpdateUser(ctx context.Context, user string, updates types.JSON) (*userStructs.ReadUser, error)
		}); ok {
			return u.UpdateUser(ctx, user, updates)
		}
	}

	// Try to access method directly
	if s, ok := w.service.(interface {
		UpdateUser(ctx context.Context, user string, updates types.JSON) (*userStructs.ReadUser, error)
	}); ok {
		return s.UpdateUser(ctx, user, updates)
	}

	logger.Errorf(ctx, "UserServiceWrapper: service type %T does not implement required method", w.service)
	return nil, fmt.Errorf("user service does not implement UpdateUser method")
}

// UserProfileServiceInterface defines the user profile service interface needed by visualization module
type UserProfileServiceInterface interface {
	Get(ctx context.Context, id string) (*userStructs.ReadUserProfile, error)
}

// UserProfileServiceWrapper wraps the actual user profile service implementation
type UserProfileServiceWrapper struct {
	serviceWrapper
}

var _ UserProfileServiceInterface = (*UserProfileServiceWrapper)(nil)
var _ wrapper = (*UserProfileServiceWrapper)(nil)

func NewUserProfileServiceWrapper(service any) *UserProfileServiceWrapper {
	return &UserProfileServiceWrapper{
		serviceWrapper: serviceWrapper{service: service},
	}
}

func (w *UserProfileServiceWrapper) Get(ctx context.Context, id string) (*userStructs.ReadUserProfile, error) {
	logger.Infof(ctx, "UserProfileServiceWrapper.Get called with service type: %T", w.service)

	// Try to access through field
	if profileField, ok := ext.GetServiceInterface(w.service, "UserProfile"); ok {
		logger.Infof(ctx, "Found UserProfile field of type: %T", profileField)
		if profile, ok := profileField.(interface {
			Get(ctx context.Context, id string) (*userStructs.ReadUserProfile, error)
		}); ok {
			return profile.Get(ctx, id)
		}
	}

	// Try to access method directly
	if s, ok := w.service.(interface {
		Get(ctx context.Context, id string) (*userStructs.ReadUserProfile, error)
	}); ok {
		return s.Get(ctx, id)
	}

	logger.Errorf(ctx, "UserProfileServiceWrapper: service type %T does not implement required method", w.service)
	return nil, fmt.Errorf("user profile service does not implement required method")
}
