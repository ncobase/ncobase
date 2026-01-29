package service

import (
	"context"
	"fmt"
	"ncobase/core/space/data"
	"ncobase/core/space/data/repository"
	"ncobase/core/space/structs"
	"ncobase/core/space/wrapper"
	userStructs "ncobase/core/user/structs"
	"sort"
	"strings"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
)

// UserSpaceRoleServiceInterface is the interface for the service.
type UserSpaceRoleServiceInterface interface {
	AddRoleToUserInSpace(ctx context.Context, u, t, r string) (*structs.UserSpaceRole, error)
	GetUserRolesInSpace(ctx context.Context, u, t string) ([]string, error)
	RemoveRoleFromUserInSpace(ctx context.Context, u, t, r string) error
	IsUserInRoleInSpace(ctx context.Context, u, t, r string) (bool, error)
	GetSpaceUsersByRole(ctx context.Context, t, r string) ([]string, error)
	ListSpaceUsers(ctx context.Context, spaceID string, params *structs.ListSpaceUsersParams) (*structs.SpaceUsersListResponse, error)
	ListSpaceRoleIDs(ctx context.Context, spaceID string) ([]string, error)
	UpdateUserSpaceRole(ctx context.Context, userID, spaceID string, req *structs.UpdateUserSpaceRoleRequest) (*structs.UserSpaceRoleResponse, error)
	BulkUpdateUserSpaceRoles(ctx context.Context, spaceID string, req *structs.BulkUpdateUserSpaceRolesRequest) (*structs.BulkUpdateResponse, error)
}

// userSpaceRoleService is the struct for the service.
type userSpaceRoleService struct {
	userSpaceRole repository.UserSpaceRoleRepositoryInterface
	userSpace     repository.UserSpaceRepositoryInterface
	usw           *wrapper.UserServiceWrapper
}

// NewUserSpaceRoleService creates a new service.
func NewUserSpaceRoleService(d *data.Data, usw *wrapper.UserServiceWrapper) UserSpaceRoleServiceInterface {
	return &userSpaceRoleService{
		userSpaceRole: repository.NewUserSpaceRoleRepository(d),
		userSpace:     repository.NewUserSpaceRepository(d),
		usw:           usw,
	}
}

// AddRoleToUserInSpace adds a role to a user in a space.
func (s *userSpaceRoleService) AddRoleToUserInSpace(ctx context.Context, u, t, r string) (*structs.UserSpaceRole, error) {
	if exists, err := s.userSpace.IsSpaceInUser(ctx, t, u); err == nil && !exists {
		if _, createErr := s.userSpace.Create(ctx, &structs.UserSpace{UserID: u, SpaceID: t}); createErr != nil {
			logger.Warnf(ctx, "Failed to create user space relation for user %s in space %s: %v", u, t, createErr)
		}
	}

	row, err := s.userSpaceRole.Create(ctx, &structs.UserSpaceRole{UserID: u, SpaceID: t, RoleID: r})
	if err := handleEntError(ctx, "UserSpaceRole", err); err != nil {
		return nil, err
	}
	return repository.SerializeUserSpaceRole(row), nil
}

// GetUserRolesInSpace retrieves all roles associated with a user in a space.
func (s *userSpaceRoleService) GetUserRolesInSpace(ctx context.Context, u string, t string) ([]string, error) {
	roleIDs, err := s.userSpaceRole.GetRolesByUserAndSpace(ctx, u, t)
	if err := handleEntError(ctx, "UserSpaceRole", err); err != nil {
		return nil, err
	}
	return roleIDs, nil
}

// RemoveRoleFromUserInSpace removes a role from a user in a space.
func (s *userSpaceRoleService) RemoveRoleFromUserInSpace(ctx context.Context, u, t, r string) error {
	err := s.userSpaceRole.DeleteByUserIDAndSpaceIDAndRoleID(ctx, u, t, r)
	if err := handleEntError(ctx, "UserSpaceRole", err); err != nil {
		return err
	}
	return nil
}

// IsUserInRoleInSpace checks if a user has a specific role in a space.
func (s *userSpaceRoleService) IsUserInRoleInSpace(ctx context.Context, u, t, r string) (bool, error) {
	hasRole, err := s.userSpaceRole.IsUserInRoleInSpace(ctx, u, t, r)
	if err := handleEntError(ctx, "UserSpaceRole", err); err != nil {
		return false, err
	}
	return hasRole, nil
}

// GetSpaceUsersByRole retrieves all users with a specific role
func (s *userSpaceRoleService) GetSpaceUsersByRole(ctx context.Context, t, r string) ([]string, error) {
	userSpaceRoles, err := s.userSpaceRole.GetBySpaceID(ctx, t)
	if err := handleEntError(ctx, "UserSpaceRole", err); err != nil {
		return nil, err
	}

	var userIDs []string
	for _, utr := range userSpaceRoles {
		if utr.RoleID == r {
			userIDs = append(userIDs, utr.UserID)
		}
	}

	return userIDs, nil
}

// ListSpaceUsers retrieves all users in a space with their roles.
func (s *userSpaceRoleService) ListSpaceUsers(ctx context.Context, spaceID string, params *structs.ListSpaceUsersParams) (*structs.SpaceUsersListResponse, error) {
	if params == nil {
		params = &structs.ListSpaceUsersParams{}
	}
	if params.Limit <= 0 {
		params.Limit = 50
	}

	// Get all user space roles for this space
	userSpaceRoles, err := s.userSpaceRole.GetBySpaceID(ctx, spaceID)
	if err := handleEntError(ctx, "UserSpaceRole", err); err != nil {
		return nil, err
	}

	// Group roles by user and track earliest role assignment time
	userRolesMap := make(map[string][]string)
	roleJoinedAt := make(map[string]int64)
	for _, utr := range userSpaceRoles {
		// Filter by role if specified
		if params.RoleID != "" && utr.RoleID != params.RoleID {
			continue
		}
		userRolesMap[utr.UserID] = append(userRolesMap[utr.UserID], utr.RoleID)
		if existing, ok := roleJoinedAt[utr.UserID]; !ok || (utr.CreatedAt > 0 && utr.CreatedAt < existing) {
			roleJoinedAt[utr.UserID] = utr.CreatedAt
		}
	}

	// Load user-space relations for join time
	userSpaceJoinedAt := map[string]int64{}
	if len(userRolesMap) > 0 {
		userSpaces, err := s.userSpace.GetBySpaceIDs(ctx, []string{spaceID})
		if err == nil {
			for _, us := range userSpaces {
				if us.UserID == "" {
					continue
				}
				if existing, ok := userSpaceJoinedAt[us.UserID]; !ok || (us.CreatedAt > 0 && us.CreatedAt < existing) {
					userSpaceJoinedAt[us.UserID] = us.CreatedAt
				}
			}
		}
	}

	// Convert to response format with user info
	var users []structs.SpaceUserInfo
	for userID, roleIDs := range userRolesMap {
		sort.Strings(roleIDs)
		userInfo := structs.SpaceUserInfo{
			UserID:      userID,
			RoleIDs:     roleIDs,
			AccessLevel: deriveAccessLevel(nil, roleIDs),
			IsActive:    true,
		}

		if joinedAt, ok := userSpaceJoinedAt[userID]; ok {
			userInfo.JoinedAt = joinedAt
		} else if joinedAt, ok := roleJoinedAt[userID]; ok {
			userInfo.JoinedAt = joinedAt
		}

		// Try to enrich user details
		if s.usw != nil && s.usw.HasUserService() {
			if user, err := s.usw.GetUserByID(ctx, userID); err == nil && user != nil {
				userInfo.Username = user.Username
				userInfo.Email = user.Email
				userInfo.IsActive = user.Status == 0
				userInfo.AccessLevel = deriveAccessLevel(user, roleIDs)
			}
		}

		users = append(users, userInfo)
	}

	// Filter by search, access level and active status
	if params.Search != "" {
		search := strings.ToLower(strings.TrimSpace(params.Search))
		filtered := users[:0]
		for _, user := range users {
			if strings.Contains(strings.ToLower(user.Username), search) || strings.Contains(strings.ToLower(user.Email), search) {
				filtered = append(filtered, user)
			}
		}
		users = filtered
	}

	if params.AccessLevel != "" {
		accessLevel := strings.ToLower(strings.TrimSpace(params.AccessLevel))
		filtered := users[:0]
		for _, user := range users {
			if strings.ToLower(user.AccessLevel) == accessLevel {
				filtered = append(filtered, user)
			}
		}
		users = filtered
	}

	if params.IsActive != "" {
		activeFilter := strings.ToLower(strings.TrimSpace(params.IsActive))
		filtered := users[:0]
		for _, user := range users {
			if (activeFilter == "true" && user.IsActive) || (activeFilter == "false" && !user.IsActive) {
				filtered = append(filtered, user)
			}
		}
		users = filtered
	}

	// Sort results
	sortBy := strings.ToLower(strings.TrimSpace(params.SortBy))
	switch sortBy {
	case "username":
		sort.SliceStable(users, func(i, j int) bool { return users[i].Username < users[j].Username })
	case "email":
		sort.SliceStable(users, func(i, j int) bool { return users[i].Email < users[j].Email })
	case "user_id":
		sort.SliceStable(users, func(i, j int) bool { return users[i].UserID < users[j].UserID })
	default:
		sort.SliceStable(users, func(i, j int) bool { return users[i].JoinedAt > users[j].JoinedAt })
	}

	if params.Direction == "backward" {
		for i, j := 0, len(users)-1; i < j; i, j = i+1, j-1 {
			users[i], users[j] = users[j], users[i]
		}
	}

	total := len(users)

	// Apply cursor-based pagination
	start := 0
	if params.Cursor != "" {
		id, ts, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		cursorValue := fmt.Sprintf("%s:%d", id, ts)
		found := false
		for i, user := range users {
			if user.GetCursorValue() == cursorValue {
				found = true
				if params.Direction == "backward" {
					start = maxInt(0, i-params.Limit)
					users = users[start:i]
				} else {
					start = i + 1
					if start < len(users) {
						users = users[start:]
					} else {
						users = []structs.SpaceUserInfo{}
					}
				}
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("invalid cursor: not found")
		}
	}

	if len(users) > params.Limit {
		users = users[:params.Limit]
	}

	var nextCursor string
	if len(users) > 0 {
		nextCursor = paging.EncodeCursor(users[len(users)-1].GetCursorValue())
	}

	response := &structs.SpaceUsersListResponse{
		Users:  users,
		Total:  total,
		Cursor: nextCursor,
	}

	return response, nil
}

// ListSpaceRoleIDs retrieves unique role IDs in a space.
func (s *userSpaceRoleService) ListSpaceRoleIDs(ctx context.Context, spaceID string) ([]string, error) {
	userSpaceRoles, err := s.userSpaceRole.GetBySpaceID(ctx, spaceID)
	if err := handleEntError(ctx, "UserSpaceRole", err); err != nil {
		return nil, err
	}

	roleSet := make(map[string]struct{})
	for _, utr := range userSpaceRoles {
		if utr.RoleID != "" {
			roleSet[utr.RoleID] = struct{}{}
		}
	}

	roleIDs := make([]string, 0, len(roleSet))
	for roleID := range roleSet {
		roleIDs = append(roleIDs, roleID)
	}
	sort.Strings(roleIDs)
	return roleIDs, nil
}

// UpdateUserSpaceRole updates a user's role in a space.
func (s *userSpaceRoleService) UpdateUserSpaceRole(ctx context.Context, userID, spaceID string, req *structs.UpdateUserSpaceRoleRequest) (*structs.UserSpaceRoleResponse, error) {
	// Remove old role
	if err := s.userSpaceRole.DeleteByUserIDAndSpaceIDAndRoleID(ctx, userID, spaceID, req.OldRoleID); err != nil {
		logger.Warnf(ctx, "Failed to remove old role %s for user %s in space %s: %v", req.OldRoleID, userID, spaceID, err)
	}

	// Add new role
	_, err := s.userSpaceRole.Create(ctx, &structs.UserSpaceRole{
		UserID:  userID,
		SpaceID: spaceID,
		RoleID:  req.NewRoleID,
	})
	if err := handleEntError(ctx, "UserSpaceRole", err); err != nil {
		return nil, err
	}

	response := &structs.UserSpaceRoleResponse{
		UserID:  userID,
		SpaceID: spaceID,
		RoleID:  req.NewRoleID,
		Status:  "updated",
	}

	return response, nil
}

// BulkUpdateUserSpaceRoles performs bulk updates on user space roles.
func (s *userSpaceRoleService) BulkUpdateUserSpaceRoles(ctx context.Context, spaceID string, req *structs.BulkUpdateUserSpaceRolesRequest) (*structs.BulkUpdateResponse, error) {
	response := &structs.BulkUpdateResponse{
		Total: len(req.Updates),
	}

	for _, update := range req.Updates {
		var err error
		var result *structs.UserSpaceRoleResponse

		switch update.Operation {
		case "add":
			var userRole *structs.UserSpaceRole
			userRole, err = s.AddRoleToUserInSpace(ctx, update.UserID, spaceID, update.RoleID)
			if err == nil {
				result = &structs.UserSpaceRoleResponse{
					UserID:  userRole.UserID,
					SpaceID: userRole.SpaceID,
					RoleID:  userRole.RoleID,
					Status:  "added",
				}
			}

		case "remove":
			err = s.RemoveRoleFromUserInSpace(ctx, update.UserID, spaceID, update.RoleID)
			if err == nil {
				result = &structs.UserSpaceRoleResponse{
					UserID:  update.UserID,
					SpaceID: spaceID,
					RoleID:  update.RoleID,
					Status:  "removed",
				}
			}

		case "update":
			if update.OldRoleID == "" {
				err = fmt.Errorf("old_role_id is required for update operation")
			} else {
				updateReq := &structs.UpdateUserSpaceRoleRequest{
					OldRoleID: update.OldRoleID,
					NewRoleID: update.RoleID,
				}
				result, err = s.UpdateUserSpaceRole(ctx, update.UserID, spaceID, updateReq)
			}

		default:
			err = fmt.Errorf("invalid operation: %s", update.Operation)
		}

		if err != nil {
			response.Failed++
			response.Errors = append(response.Errors, structs.BulkUpdateError{
				UserID: update.UserID,
				RoleID: update.RoleID,
				Error:  err.Error(),
			})
		} else {
			response.Success++
			if result != nil {
				response.Results = append(response.Results, *result)
			}
		}
	}

	return response, nil
}

func deriveAccessLevel(user *userStructs.ReadUser, roleIDs []string) string {
	if user != nil && user.IsAdmin {
		return "admin"
	}
	if len(roleIDs) == 0 {
		return "limited"
	}
	if len(roleIDs) > 1 {
		return "elevated"
	}
	return "standard"
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
