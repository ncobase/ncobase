package service

import (
	"context"
	"ncobase/auth/data"
	"ncobase/auth/data/ent"
	"ncobase/auth/data/repository"
	"ncobase/auth/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
)

// SessionServiceInterface defines the session service interface
type SessionServiceInterface interface {
	Create(ctx context.Context, body *structs.SessionBody, tokenID string) (*structs.ReadSession, error)
	Update(ctx context.Context, id string, body *structs.UpdateSessionBody) (*structs.ReadSession, error)
	GetByID(ctx context.Context, id string) (*structs.ReadSession, error)
	GetByTokenID(ctx context.Context, tokenID string) (*structs.ReadSession, error)
	List(ctx context.Context, params *structs.ListSessionParams) (paging.Result[*structs.ReadSession], error)
	Delete(ctx context.Context, id string) error
	DeactivateByUserID(ctx context.Context, userID string) error
	DeactivateByTokenID(ctx context.Context, tokenID string) error
	UpdateLastAccess(ctx context.Context, tokenID string) error
	CleanupExpiredSessions(ctx context.Context) error
	GetActiveSessionsCount(ctx context.Context, userID string) int
	Serialize(session *ent.Session) *structs.ReadSession
	Serializes(sessions []*ent.Session) []*structs.ReadSession
}

// sessionService implements the SessionServiceInterface
type sessionService struct {
	r repository.SessionRepositoryInterface
}

// NewSessionService creates a new session service
func NewSessionService(d *data.Data) SessionServiceInterface {
	return &sessionService{
		r: repository.NewSessionRepository(d),
	}
}

// Create creates a new session
func (s *sessionService) Create(ctx context.Context, body *structs.SessionBody, tokenID string) (*structs.ReadSession, error) {
	row, err := s.r.Create(ctx, body, tokenID)
	if err != nil {
		return nil, err
	}
	return s.Serialize(row), nil
}

// Update updates a session
func (s *sessionService) Update(ctx context.Context, id string, body *structs.UpdateSessionBody) (*structs.ReadSession, error) {
	row, err := s.r.Update(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return s.Serialize(row), nil
}

// GetByID retrieves a session by ID
func (s *sessionService) GetByID(ctx context.Context, id string) (*structs.ReadSession, error) {
	row, err := s.r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.Serialize(row), nil
}

// GetByTokenID retrieves a session by token ID
func (s *sessionService) GetByTokenID(ctx context.Context, tokenID string) (*structs.ReadSession, error) {
	row, err := s.r.GetByTokenID(ctx, tokenID)
	if err != nil {
		return nil, err
	}
	return s.Serialize(row), nil
}

// List retrieves sessions with pagination
func (s *sessionService) List(ctx context.Context, params *structs.ListSessionParams) (paging.Result[*structs.ReadSession], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadSession, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, err := s.r.List(ctx, &lp)
		if err != nil {
			logger.Errorf(ctx, "Error listing sessions: %v", err)
			return nil, 0, err
		}

		total := s.r.CountX(ctx, params)
		return s.Serializes(rows), total, nil
	})
}

// Delete deletes a session
func (s *sessionService) Delete(ctx context.Context, id string) error {
	return s.r.Delete(ctx, id)
}

// DeactivateByUserID deactivates all sessions for a user
func (s *sessionService) DeactivateByUserID(ctx context.Context, userID string) error {
	return s.r.DeactivateByUserID(ctx, userID)
}

// DeactivateByTokenID deactivates a session by token ID
func (s *sessionService) DeactivateByTokenID(ctx context.Context, tokenID string) error {
	return s.r.DeactivateByTokenID(ctx, tokenID)
}

// UpdateLastAccess updates the last access time for a session
func (s *sessionService) UpdateLastAccess(ctx context.Context, tokenID string) error {
	return s.r.UpdateLastAccess(ctx, tokenID)
}

// CleanupExpiredSessions removes expired sessions
func (s *sessionService) CleanupExpiredSessions(ctx context.Context) error {
	return s.r.CleanupExpiredSessions(ctx)
}

// GetActiveSessionsCount gets the count of active sessions for a user
func (s *sessionService) GetActiveSessionsCount(ctx context.Context, userID string) int {
	isActive := true
	params := &structs.ListSessionParams{
		UserID:   userID,
		IsActive: &isActive,
	}
	return s.r.CountX(ctx, params)
}

// Serialize converts ent.Session to structs.ReadSession
func (s *sessionService) Serialize(row *ent.Session) *structs.ReadSession {
	return &structs.ReadSession{
		ID:           row.ID,
		UserID:       row.UserID,
		TokenID:      row.TokenID,
		DeviceInfo:   &row.DeviceInfo,
		IPAddress:    row.IPAddress,
		UserAgent:    row.UserAgent,
		Location:     row.Location,
		LoginMethod:  row.LoginMethod,
		IsActive:     row.IsActive,
		LastAccessAt: &row.LastAccessAt,
		ExpiresAt:    &row.ExpiresAt,
		CreatedAt:    &row.CreatedAt,
		UpdatedAt:    &row.UpdatedAt,
	}
}

// Serializes converts multiple ent.Session to structs.ReadSession
func (s *sessionService) Serializes(rows []*ent.Session) []*structs.ReadSession {
	var sessions []*structs.ReadSession
	for _, row := range rows {
		sessions = append(sessions, s.Serialize(row))
	}
	return sessions
}
