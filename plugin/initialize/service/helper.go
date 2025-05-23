package service

import (
	"context"
	"fmt"
	userStructs "ncobase/user/structs"

	"github.com/ncobase/ncore/logging/logger"
)

// getAdminUser retrieves the admin user based on predefined priorities
func (s *Service) getAdminUser(ctx context.Context, operation string) (*userStructs.ReadUser, error) {
	// Try to get admin user by priority
	adminCandidates := []string{
		"chief.executive", // CEO - Highest priority
		"super",           // System administrator
		"hr.manager",      // HR manager
		"finance.manager", // Finance manager
		"tech.lead",       // Technical lead
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
