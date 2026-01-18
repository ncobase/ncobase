package service

import (
	"context"
	"errors"
	"fmt"
	"ncobase/core/user/data/ent"
	"ncobase/core/user/data/repository"
	"ncobase/core/user/event"
	"ncobase/core/user/structs"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/messaging/email"
	"github.com/ncobase/ncore/security/crypto"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/nanoid"
)

// UserServiceInterface is the interface for the service.
type UserServiceInterface interface {
	Get(ctx context.Context, username string) (*structs.ReadUser, error)
	UpdatePassword(ctx context.Context, body *structs.UserPassword) error
	CreateUser(ctx context.Context, body *structs.UserBody) (*structs.ReadUser, error)
	UpdateUser(ctx context.Context, user string, updates types.JSON) (*structs.ReadUser, error)
	GetByID(ctx context.Context, u string) (*structs.ReadUser, error)
	Delete(ctx context.Context, u string) error
	List(ctx context.Context, params *structs.ListUserParams) (paging.Result[*structs.ReadUser], error)
	FindByID(ctx context.Context, id string) (*structs.ReadUser, error)
	FindUser(ctx context.Context, m *structs.FindUser) (*structs.ReadUser, error)
	VerifyPassword(ctx context.Context, userID string, password string) any
	Serializes(rows []*ent.User) []*structs.ReadUser
	Serialize(user *ent.User) *structs.ReadUser
	CountX(ctx context.Context, params *structs.ListUserParams) int
	GetFiltered(ctx context.Context, searchQuery, roleFilter, statusFilter, sortBy string) ([]*structs.ReadUser, error)
	GetActiveUsers(ctx context.Context) ([]*structs.ReadUser, error)
	GetUserByEmail(ctx context.Context, email string) (*structs.ReadUser, error)
	GetUserByUsername(ctx context.Context, username string) (*structs.ReadUser, error)
	UpdateStatus(ctx context.Context, userID string, status int) (*structs.ReadUser, error)
	SendPasswordResetEmail(ctx context.Context, userID string) error
}

// userService is the struct for the service.
type userService struct {
	user repository.UserRepositoryInterface
	ep   event.PublisherInterface
}

// NewUserService creates a new service.
func NewUserService(repo *repository.Repository, ep event.PublisherInterface) UserServiceInterface {
	return &userService{
		user: repo.User,
		ep:   ep,
	}
}

// Get get user service
func (s *userService) Get(ctx context.Context, username string) (*structs.ReadUser, error) {
	if username == "" {
		return nil, errors.New(ecode.FieldIsInvalid("username"))
	}
	user, err := s.FindUser(ctx, &structs.FindUser{Username: username})
	if err := handleEntError(ctx, "User", err); err != nil {
		return nil, err
	}
	return user, nil
}

// UpdatePassword update user password service
func (s *userService) UpdatePassword(ctx context.Context, body *structs.UserPassword) error {
	if body.NewPassword == "" {
		return errors.New(ecode.FieldIsEmpty("new password"))
	}
	if body.Confirm != body.NewPassword {
		return errors.New(ecode.FieldIsInvalid("confirm password"))
	}
	verifyResult := s.VerifyPassword(ctx, body.User, body.OldPassword)
	switch v := verifyResult.(type) {
	case VerifyPasswordResult:
		if v.Valid == false {
			return errors.New(v.Error)
		} else if v.Valid && v.NeedsPasswordSet == true { // print a log for user's first password setting
			logger.Warnf(ctx, "User %s is setting password for the first time", body.User)
		}
	case error:
		return v
	}

	err := s.updatePassword(ctx, body)

	if err := handleEntError(ctx, "User", err); err != nil {
		return err
	}

	return nil
}

// CreateUser creates a new user.
func (s *userService) CreateUser(ctx context.Context, body *structs.UserBody) (*structs.ReadUser, error) {
	if body != nil && body.Username == "" {
		return nil, errors.New(ecode.FieldIsInvalid("username"))
	}

	row, err := s.user.Create(ctx, body)
	if err := handleEntError(ctx, "User", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// UpdateUser updates an existing user.
func (s *userService) UpdateUser(ctx context.Context, u string, updates types.JSON) (*structs.ReadUser, error) {
	user, err := s.user.Update(ctx, u, updates)
	if err != nil {
		return nil, err
	}
	return s.Serialize(user), nil
}

// GetByID retrieves a user by their ID.
func (s *userService) GetByID(ctx context.Context, u string) (*structs.ReadUser, error) {
	row, err := s.user.GetByID(ctx, u)
	if err := handleEntError(ctx, "User", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Delete deletes a user by their ID.
func (s *userService) Delete(ctx context.Context, u string) error {
	err := s.user.Delete(ctx, u)
	if err := handleEntError(ctx, "User", err); err != nil {
		return err
	}
	return nil
}

// FindByID find user by ID
func (s *userService) FindByID(ctx context.Context, id string) (*structs.ReadUser, error) {
	row, err := s.user.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.Serialize(row), nil
}

// FindUser find user by username, email, or phone
func (s *userService) FindUser(ctx context.Context, m *structs.FindUser) (*structs.ReadUser, error) {
	row, err := s.user.Find(ctx, &structs.FindUser{
		Username: m.Username,
		Email:    m.Email,
		Phone:    m.Phone,
	})
	if err != nil {
		return nil, err
	}
	return s.Serialize(row), nil
}

// VerifyPasswordResult Verify password result
type VerifyPasswordResult struct {
	Valid            bool
	NeedsPasswordSet bool
	Error            string
}

// VerifyPassword verify user password
func (s *userService) VerifyPassword(ctx context.Context, u string, password string) any {
	user, err := s.user.FindUser(ctx, &structs.FindUser{Username: u})
	if ent.IsNotFound(err) {
		return VerifyPasswordResult{Valid: false, NeedsPasswordSet: false, Error: "user not found"}
	} else if err != nil {
		return VerifyPasswordResult{Valid: false, NeedsPasswordSet: false, Error: fmt.Sprintf("error getting user by ID: %v", err)}
	}
	if user.Password == "" {
		return VerifyPasswordResult{Valid: true, NeedsPasswordSet: true, Error: "user password not set"}
	}

	if crypto.ComparePassword(user.Password, password) {
		return VerifyPasswordResult{Valid: true, NeedsPasswordSet: false, Error: ""}
	}

	return VerifyPasswordResult{Valid: false, NeedsPasswordSet: false, Error: "wrong password"}
}

// updatePassword update user password
func (s *userService) updatePassword(ctx context.Context, body *structs.UserPassword) error {
	err := s.user.UpdatePassword(ctx, body)
	if err != nil {
		logger.Errorf(ctx, "Error updating password for user %s: %v", body.User, err)
	}

	return err
}

// List lists all users.
func (s *userService) List(ctx context.Context, params *structs.ListUserParams) (paging.Result[*structs.ReadUser], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadUser, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, err := s.user.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			logger.Errorf(ctx, "Error listing users: %v", err)
			return nil, 0, err
		}

		total := s.CountX(ctx, params)

		return s.Serializes(rows), total, nil
	})
}

// CountX gets a count of users.
func (s *userService) CountX(ctx context.Context, params *structs.ListUserParams) int {
	return s.user.CountX(ctx, params)
}

// GetFiltered gets filtered users.
func (s *userService) GetFiltered(ctx context.Context, searchQuery, roleFilter, statusFilter, sortBy string) ([]*structs.ReadUser, error) {
	params := &structs.ListUserParams{
		SearchQuery:  searchQuery,
		RoleFilter:   roleFilter,
		StatusFilter: statusFilter,
		SortBy:       sortBy,
		Limit:        100, // Default limit
	}

	// Clean up filters
	if roleFilter == "all" {
		params.RoleFilter = ""
	}
	if statusFilter == "all" {
		params.StatusFilter = ""
	}

	rows, err := s.user.List(ctx, params)
	if err != nil {
		return nil, err
	}

	return s.Serializes(rows), nil
}

// GetActiveUsers gets active users
func (s *userService) GetActiveUsers(ctx context.Context) ([]*structs.ReadUser, error) {
	// Active users are those with status = 0
	params := &structs.ListUserParams{
		Status: 0,                 // 0: activated, 1: unactivated, 2: disabled
		SortBy: "last_login_desc", // Sort by most recent login
		Limit:  100,               // Reasonable limit for active users
	}

	result, err := s.List(ctx, params)
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}

// GetUserByEmail gets a user by email.
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*structs.ReadUser, error) {
	if email == "" {
		return nil, errors.New(ecode.FieldIsInvalid("email"))
	}
	user, err := s.FindUser(ctx, &structs.FindUser{Email: email})
	if err = handleEntError(ctx, "User", err); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByUsername gets a user by username.
func (s *userService) GetUserByUsername(ctx context.Context, username string) (*structs.ReadUser, error) {
	if username == "" {
		return nil, errors.New(ecode.FieldIsInvalid("username"))
	}
	user, err := s.FindUser(ctx, &structs.FindUser{Username: username})
	if err = handleEntError(ctx, "User", err); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateStatus updates a user's status.
func (s *userService) UpdateStatus(ctx context.Context, userID string, status int) (*structs.ReadUser, error) {
	updates := types.JSON{
		"status": status,
	}
	user, err := s.UpdateUser(ctx, userID, updates)
	if err != nil {
		return nil, err
	}

	// Publish status updated event
	if s.ep != nil {
		statusText := "unknown"
		switch status {
		case 0:
			statusText = "active"
		case 1:
			statusText = "inactive"
		case 2:
			statusText = "pending"
		}

		s.ep.PublishStatusUpdated(ctx, userID, &types.JSON{
			"status":      status,
			"status_text": statusText,
		})
	}

	return user, nil
}

func (s *userService) SendPasswordResetEmail(ctx context.Context, userID string) error {
	user, err := s.user.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.Email == "" {
		return errors.New("user does not have an email address")
	}
	// Generate a temporary password
	tempPassword := nanoid.String(12)
	// Hash the temporary password
	hashedPassword, err := crypto.HashPassword(ctx, tempPassword)
	if err != nil {
		return err
	}

	// Update the user's password
	err = s.user.UpdatePasswordByID(ctx, userID, hashedPassword)
	if err != nil {
		return err
	}

	// Send the password reset email
	conf := ctxutil.GetConfig(ctx)
	template := email.Template{
		Subject:  "Password Reset",
		Template: "password-reset",
		Keyword:  "Password Reset",
		URL:      conf.Frontend.SignInURL,
		Data: map[string]any{
			"username":     user.Username,
			"tempPassword": tempPassword,
		},
	}

	_, err = ctxutil.SendEmailWithTemplate(ctx, user.Email, template)

	// Publish password reset event
	if s.ep != nil {
		s.ep.PublishPasswordReset(ctx, userID, &types.JSON{
			"username": user.Username,
			"email":    user.Email,
		})
	}

	return err
}

// Serializes serializes users
func (s *userService) Serializes(rows []*ent.User) []*structs.ReadUser {
	rs := make([]*structs.ReadUser, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize serialize a user
func (s *userService) Serialize(user *ent.User) *structs.ReadUser {
	return &structs.ReadUser{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Phone:       user.Phone,
		IsCertified: user.IsCertified,
		IsAdmin:     user.IsAdmin,
		Status:      user.Status,
		Extras:      &user.Extras,
		CreatedAt:   &user.CreatedAt,
		UpdatedAt:   &user.UpdatedAt,
	}
}

// // serializeUserRoles serialize user roles
// func (svc *Service) serializeUserRoles(rows []*ent.UserRole) []*structs.UserRole {
// 	var userRoles []*structs.UserRole
// 	for _, row := range rows {
// 		userRoles = append(userRoles, svc.serializeUserRole(row))
// 	}
// 	return userRoles
// }
//
// // serializeUserRole serialize user role
// func (svc *Service) serializeUserRole(row *ent.UserRole) *structs.UserRole {
// 	return &structs.UserRole{
// 		UserID: row.UserID,
// 		RoleID: row.RoleID,
// 	}
// }
