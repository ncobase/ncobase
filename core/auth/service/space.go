package service

import (
	"context"
	"errors"
	accessStructs "ncobase/access/structs"
	"ncobase/auth/data"
	"ncobase/auth/wrapper"
	spaceStructs "ncobase/space/structs"
	userStructs "ncobase/user/structs"

	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils/slug"
)

// AuthSpaceServiceInterface is the interface for the service.
type AuthSpaceServiceInterface interface {
	CreateInitialSpace(ctx context.Context, body *spaceStructs.CreateSpaceBody) (*spaceStructs.ReadSpace, error)
	IsCreateSpace(ctx context.Context, body *spaceStructs.CreateSpaceBody) (*spaceStructs.ReadSpace, error)
}

// authSpaceService is the struct for the service.
type authSpaceService struct {
	d *data.Data

	usw *wrapper.UserServiceWrapper
	tsw *wrapper.SpaceServiceWrapper
	asw *wrapper.AccessServiceWrapper
}

// NewAuthSpaceService creates a new service.
func NewAuthSpaceService(d *data.Data, usw *wrapper.UserServiceWrapper, tsw *wrapper.SpaceServiceWrapper, asw *wrapper.AccessServiceWrapper) AuthSpaceServiceInterface {
	return &authSpaceService{
		d:   d,
		usw: usw,
		tsw: tsw,
		asw: asw,
	}
}

// CreateInitialSpace creates the initial space and sets up roles and user relationships
func (s *authSpaceService) CreateInitialSpace(ctx context.Context, body *spaceStructs.CreateSpaceBody) (*spaceStructs.ReadSpace, error) {
	// Create the space
	space, err := s.tsw.CreateSpace(ctx, body)
	if err != nil {
		logger.Errorf(ctx, "Failed to create space: %v", err)
		return nil, err
	}

	if body.CreatedBy == nil {
		logger.Infof(ctx, "No user specified for space creation, skipping role assignment")
		return space, nil
	}

	// Get or create the super admin role
	superAdminRole, err := s.getSuperAdminRole(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to get super admin role: %v", err)
		return nil, err
	}

	// Add user to space
	if _, err := s.tsw.AddUserToSpace(ctx, *body.CreatedBy, space.ID); err != nil {
		logger.Errorf(ctx, "Failed to add user to space: %v", err)
		return nil, err
	}

	// Assign global role to user
	if err := s.asw.AddRoleToUser(ctx, *body.CreatedBy, superAdminRole.ID); err != nil {
		logger.Warnf(ctx, "Failed to assign global role to user: %v", err)
	}

	// Assign space-specific role
	if _, err := s.tsw.AddRoleToUserInSpace(ctx, *body.CreatedBy, space.ID, superAdminRole.ID); err != nil {
		logger.Warnf(ctx, "Failed to assign space role to user: %v", err)
	}

	logger.Infof(ctx, "Successfully created initial space '%s' with user assignments", space.Name)
	return space, nil
}

// IsCreateSpace checks conditions and creates space if needed
func (s *authSpaceService) IsCreateSpace(ctx context.Context, body *spaceStructs.CreateSpaceBody) (*spaceStructs.ReadSpace, error) {
	if body.CreatedBy == nil {
		return nil, errors.New("user ID is required for space creation")
	}

	// Generate slug if not provided
	if body.Slug == "" && body.Name != "" {
		body.Slug = slug.Unicode(body.Name)
	}

	// Check if user already has a space
	existingSpace, err := s.tsw.GetSpaceByUser(ctx, *body.CreatedBy)
	if err == nil && existingSpace != nil {
		logger.Infof(ctx, "User already has space '%s'", existingSpace.Name)
		return existingSpace, nil
	}

	// Check if this is system initialization (first user or explicitly requested)
	userCount := s.usw.CountX(ctx, &userStructs.ListUserParams{})
	shouldCreateInitial := userCount <= 1 || body.Name != ""

	if shouldCreateInitial {
		logger.Infof(ctx, "Creating initial space for user (user count: %d)", userCount)
		return s.CreateInitialSpace(ctx, body)
	}

	logger.Infof(ctx, "Space creation conditions not met")
	return nil, nil
}

// getSuperAdminRole gets or creates super admin role
func (s *authSpaceService) getSuperAdminRole(ctx context.Context) (*accessStructs.ReadRole, error) {
	// Try to find existing super admin role
	role, err := s.asw.FindRole(ctx, &accessStructs.FindRole{
		Slug: "super-admin",
	})
	if err == nil && role != nil {
		return role, nil
	}

	// Try system admin as fallback
	role, err = s.asw.FindRole(ctx, &accessStructs.FindRole{
		Slug: "system-admin",
	})
	if err == nil && role != nil {
		return role, nil
	}

	// Create new super admin role if none exists
	logger.Infof(ctx, "Creating new super admin role")
	return s.asw.CreateSuperAdminRole(ctx)
}
