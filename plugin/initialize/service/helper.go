package service

import (
	"context"
	"fmt"
	spaceStructs "ncobase/core/space/structs"
	userStructs "ncobase/core/user/structs"

	"github.com/ncobase/ncore/logging/logger"
)

// getDefaultSpaceSlug returns default space slug based on mode
func (s *Service) getDefaultSpaceSlug() string {
	switch s.state.DataMode {
	case "website":
		return "website-platform"
	case "company":
		return "digital-company"
	case "enterprise":
		return "digital-enterprise"
	default:
		return "website-platform"
	}
}

// getDefaultSpace retrieves default space based on mode
func (s *Service) getDefaultSpace(ctx context.Context) (*spaceStructs.ReadSpace, error) {
	spaceSlug := s.getDefaultSpaceSlug()

	space, err := s.ts.Space.GetBySlug(ctx, spaceSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to get default space '%s': %v", spaceSlug, err)
	}

	return space, nil
}

// getAdminUser retrieves admin user based on mode
func (s *Service) getAdminUser(ctx context.Context, operation string) (*userStructs.ReadUser, error) {
	var adminCandidates []string

	switch s.state.DataMode {
	case "website":
		adminCandidates = []string{"admin", "super", "manager"}
	case "company":
		adminCandidates = []string{"company.admin", "super", "admin", "manager"}
	case "enterprise":
		adminCandidates = []string{"enterprise.admin", "super", "admin", "dept.manager"}
	default:
		adminCandidates = []string{"admin", "super"}
	}

	for _, username := range adminCandidates {
		user, err := s.us.User.Get(ctx, username)
		if err == nil && user != nil {
			logger.Debugf(ctx, "Using user '%s' for %s", username, operation)
			return user, nil
		}
	}

	return nil, fmt.Errorf("no suitable admin user found for %s", operation)
}
