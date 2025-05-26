package service

import (
	"context"
	"fmt"
	userStructs "ncobase/user/structs"

	"github.com/ncobase/ncore/logging/logger"
)

// getAdminUser retrieves admin user based on mode
func (s *Service) getAdminUser(ctx context.Context, operation string) (*userStructs.ReadUser, error) {
	var adminCandidates []string

	switch s.state.DataMode {
	case "website":
		adminCandidates = []string{
			"admin",   // Site administrator
			"super",   // System administrator
			"manager", // Content manager
		}
	case "company":
		adminCandidates = []string{
			"company.admin", // Updated from "enterprise.admin"
			"super",         // System administrator
			"admin",         // System admin
			"manager",       // Department manager
		}
	case "enterprise":
		adminCandidates = []string{
			"enterprise.admin", // Enterprise admin
			"super",            // System administrator
			"dept.manager",     // Updated from individual role names
			"team.leader",      // Updated from individual role names
			"admin",            // System admin
		}
	default:
		adminCandidates = []string{
			"admin",
			"super",
		}
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
