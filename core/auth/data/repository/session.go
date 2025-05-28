package repository

import (
	"context"
	"fmt"
	"ncobase/auth/data"
	"ncobase/auth/data/ent"
	sessionEnt "ncobase/auth/data/ent/session"
	"ncobase/auth/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/redis/go-redis/v9"
)

// SessionRepositoryInterface represents the session repository interface
type SessionRepositoryInterface interface {
	Create(ctx context.Context, body *structs.SessionBody, tokenID string) (*ent.Session, error)
	Update(ctx context.Context, id string, body *structs.UpdateSessionBody) (*ent.Session, error)
	GetByID(ctx context.Context, id string) (*ent.Session, error)
	GetByTokenID(ctx context.Context, tokenID string) (*ent.Session, error)
	List(ctx context.Context, params *structs.ListSessionParams) ([]*ent.Session, error)
	CountX(ctx context.Context, params *structs.ListSessionParams) int
	Delete(ctx context.Context, id string) error
	DeactivateByUserID(ctx context.Context, userID string) error
	DeactivateByTokenID(ctx context.Context, tokenID string) error
	CleanupExpiredSessions(ctx context.Context) error
	UpdateLastAccess(ctx context.Context, tokenID string) error
}

// sessionRepository implements the SessionRepositoryInterface
type sessionRepository struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.Session]
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(d *data.Data) SessionRepositoryInterface {
	ec := d.GetMasterEntClient()
	rc := d.GetRedis()
	return &sessionRepository{ec, rc, cache.NewCache[ent.Session](rc, "ncse_session")}
}

// Create creates a new session
func (r *sessionRepository) Create(ctx context.Context, body *structs.SessionBody, tokenID string) (*ent.Session, error) {
	now := time.Now().UnixMilli()
	expiresAt := time.Now().Add(7 * 24 * time.Hour).UnixMilli() // 7 days default

	builder := r.ec.Session.Create()
	builder.SetUserID(body.UserID)
	builder.SetTokenID(tokenID)
	builder.SetIsActive(true)
	builder.SetLastAccessAt(now)
	builder.SetExpiresAt(expiresAt)

	if body.DeviceInfo != nil {
		builder.SetDeviceInfo(*body.DeviceInfo)
	}
	if body.IPAddress != "" {
		builder.SetIPAddress(body.IPAddress)
	}
	if body.UserAgent != "" {
		builder.SetUserAgent(body.UserAgent)
	}
	if body.Location != "" {
		builder.SetLocation(body.Location)
	}
	if body.LoginMethod != "" {
		builder.SetLoginMethod(body.LoginMethod)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "sessionRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the session
	cacheKey := fmt.Sprintf("token:%s", tokenID)
	if err := r.c.Set(ctx, cacheKey, row); err != nil {
		logger.Errorf(ctx, "sessionRepo.Create cache error: %v", err)
	}

	return row, nil
}

// Update updates a session
func (r *sessionRepository) Update(ctx context.Context, id string, body *structs.UpdateSessionBody) (*ent.Session, error) {
	builder := r.ec.Session.UpdateOneID(id)

	if body.LastAccessAt != nil {
		builder.SetLastAccessAt(*body.LastAccessAt)
	}
	if body.Location != "" {
		builder.SetLocation(body.Location)
	}
	if body.IsActive != nil {
		builder.SetIsActive(*body.IsActive)
	}
	if body.DeviceInfo != nil {
		builder.SetDeviceInfo(*body.DeviceInfo)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "sessionRepo.Update error: %v", err)
		return nil, err
	}

	// Remove from cache to force refresh
	cacheKey := fmt.Sprintf("token:%s", row.TokenID)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		logger.Errorf(ctx, "sessionRepo.Update cache error: %v", err)
	}

	return row, nil
}

// GetByID retrieves a session by ID
func (r *sessionRepository) GetByID(ctx context.Context, id string) (*ent.Session, error) {
	cacheKey := fmt.Sprintf("id:%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	row, err := r.ec.Session.Query().Where(sessionEnt.IDEQ(id)).Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "sessionRepo.GetByID error: %v", err)
		return nil, err
	}

	// Cache the result
	if err := r.c.Set(ctx, cacheKey, row); err != nil {
		logger.Errorf(ctx, "sessionRepo.GetByID cache error: %v", err)
	}

	return row, nil
}

// GetByTokenID retrieves a session by token ID
func (r *sessionRepository) GetByTokenID(ctx context.Context, tokenID string) (*ent.Session, error) {
	cacheKey := fmt.Sprintf("token:%s", tokenID)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	row, err := r.ec.Session.Query().Where(sessionEnt.TokenIDEQ(tokenID)).Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "sessionRepo.GetByTokenID error: %v", err)
		return nil, err
	}

	// Cache the result
	if err := r.c.Set(ctx, cacheKey, row); err != nil {
		logger.Errorf(ctx, "sessionRepo.GetByTokenID cache error: %v", err)
	}

	return row, nil
}

// List retrieves sessions with pagination
func (r *sessionRepository) List(ctx context.Context, params *structs.ListSessionParams) ([]*ent.Session, error) {
	builder := r.ec.Session.Query()

	// Apply filters
	if params.UserID != "" {
		builder.Where(sessionEnt.UserIDEQ(params.UserID))
	}
	if params.IsActive != nil {
		builder.Where(sessionEnt.IsActiveEQ(*params.IsActive))
	}

	// Apply cursor-based pagination
	if params.Cursor != "" {
		id, timestamp, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %v", err)
		}

		if !nanoid.IsPrimaryKey(id) {
			return nil, fmt.Errorf("invalid id in cursor: %s", id)
		}

		if params.Direction == "backward" {
			builder.Where(
				sessionEnt.Or(
					sessionEnt.CreatedAtGT(timestamp),
					sessionEnt.And(
						sessionEnt.CreatedAtEQ(timestamp),
						sessionEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				sessionEnt.Or(
					sessionEnt.CreatedAtLT(timestamp),
					sessionEnt.And(
						sessionEnt.CreatedAtEQ(timestamp),
						sessionEnt.IDLT(id),
					),
				),
			)
		}
	}

	// Apply ordering
	if params.Direction == "backward" {
		builder.Order(ent.Asc(sessionEnt.FieldCreatedAt), ent.Asc(sessionEnt.FieldID))
	} else {
		builder.Order(ent.Desc(sessionEnt.FieldCreatedAt), ent.Desc(sessionEnt.FieldID))
	}

	// Apply limit
	builder.Limit(params.Limit)

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "sessionRepo.List error: %v", err)
		return nil, err
	}

	return rows, nil
}

// CountX gets a count of sessions
func (r *sessionRepository) CountX(ctx context.Context, params *structs.ListSessionParams) int {
	builder := r.ec.Session.Query()

	if params.UserID != "" {
		builder.Where(sessionEnt.UserIDEQ(params.UserID))
	}
	if params.IsActive != nil {
		builder.Where(sessionEnt.IsActiveEQ(*params.IsActive))
	}

	return builder.CountX(ctx)
}

// Delete deletes a session
func (r *sessionRepository) Delete(ctx context.Context, id string) error {
	session, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := r.ec.Session.DeleteOneID(id).Exec(ctx); err != nil {
		logger.Errorf(ctx, "sessionRepo.Delete error: %v", err)
		return err
	}

	// Remove from cache
	cacheKeys := []string{
		fmt.Sprintf("id:%s", id),
		fmt.Sprintf("token:%s", session.TokenID),
	}
	for _, key := range cacheKeys {
		if err := r.c.Delete(ctx, key); err != nil {
			logger.Errorf(ctx, "sessionRepo.Delete cache error: %v", err)
		}
	}

	return nil
}

// DeactivateByUserID deactivates all sessions for a user
func (r *sessionRepository) DeactivateByUserID(ctx context.Context, userID string) error {
	_, err := r.ec.Session.Update().
		Where(sessionEnt.UserIDEQ(userID)).
		SetIsActive(false).
		Save(ctx)

	if err != nil {
		logger.Errorf(ctx, "sessionRepo.DeactivateByUserID error: %v", err)
		return err
	}

	// Clear cache for user sessions (simple approach - clear all session cache)
	// In production, you might want to implement more efficient cache invalidation
	return nil
}

// DeactivateByTokenID deactivates a session by token ID
func (r *sessionRepository) DeactivateByTokenID(ctx context.Context, tokenID string) error {
	session, err := r.GetByTokenID(ctx, tokenID)
	if err != nil {
		return err
	}

	_, err = r.ec.Session.UpdateOneID(session.ID).
		SetIsActive(false).
		Save(ctx)

	if err != nil {
		logger.Errorf(ctx, "sessionRepo.DeactivateByTokenID error: %v", err)
		return err
	}

	// Remove from cache
	cacheKey := fmt.Sprintf("token:%s", tokenID)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		logger.Errorf(ctx, "sessionRepo.DeactivateByTokenID cache error: %v", err)
	}

	return nil
}

// CleanupExpiredSessions removes expired sessions
func (r *sessionRepository) CleanupExpiredSessions(ctx context.Context) error {
	now := time.Now().UnixMilli()

	count, err := r.ec.Session.Delete().
		Where(sessionEnt.ExpiresAtLT(now)).
		Exec(ctx)

	if err != nil {
		logger.Errorf(ctx, "sessionRepo.CleanupExpiredSessions error: %v", err)
		return err
	}

	logger.Infof(ctx, "Cleaned up %d expired sessions", count)
	return nil
}

// UpdateLastAccess updates the last access time for a session
func (r *sessionRepository) UpdateLastAccess(ctx context.Context, tokenID string) error {
	now := time.Now().UnixMilli()

	session, err := r.GetByTokenID(ctx, tokenID)
	if err != nil {
		return err
	}

	_, err = r.ec.Session.UpdateOneID(session.ID).
		SetLastAccessAt(now).
		Save(ctx)

	if err != nil {
		logger.Errorf(ctx, "sessionRepo.UpdateLastAccess error: %v", err)
		return err
	}

	// Remove from cache to force refresh
	cacheKey := fmt.Sprintf("token:%s", tokenID)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		logger.Errorf(ctx, "sessionRepo.UpdateLastAccess cache error: %v", err)
	}

	return nil
}
