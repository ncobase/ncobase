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
)

// SessionRepositoryInterface defines session repository operations
type SessionRepositoryInterface interface {
	Create(ctx context.Context, body *structs.SessionBody, tokenID string) (*ent.Session, error)
	Update(ctx context.Context, id string, body *structs.UpdateSessionBody) (*ent.Session, error)
	GetByID(ctx context.Context, id string) (*ent.Session, error)
	GetByTokenID(ctx context.Context, tokenID string) (*ent.Session, error)
	List(ctx context.Context, params *structs.ListSessionParams) ([]*ent.Session, error)
	Delete(ctx context.Context, id string) error
	DeactivateByUserID(ctx context.Context, userID string) error
	DeactivateByTokenID(ctx context.Context, tokenID string) error
	UpdateLastAccess(ctx context.Context, tokenID string) error
	CleanupExpiredSessions(ctx context.Context) error
	CountX(ctx context.Context, params *structs.ListSessionParams) int
}

// sessionRepository implements SessionRepositoryInterface
type sessionRepository struct {
	data              *data.Data
	sessionCache      cache.ICache[ent.Session]
	tokenSessionCache cache.ICache[ent.Session]
	userSessionsCache cache.ICache[[]string] // Store session IDs for user
	sessionTTL        time.Duration
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(d *data.Data) SessionRepositoryInterface {
	redisClient := d.GetRedis()

	return &sessionRepository{
		data:              d,
		sessionCache:      cache.NewCache[ent.Session](redisClient, "ncse_auth:sessions"),
		tokenSessionCache: cache.NewCache[ent.Session](redisClient, "ncse_auth:token_sessions"),
		userSessionsCache: cache.NewCache[[]string](redisClient, "ncse_auth:user_sessions"),
		sessionTTL:        time.Hour * 2, // 2 hours cache TTL
	}
}

// Create creates a new session
func (r *sessionRepository) Create(ctx context.Context, body *structs.SessionBody, tokenID string) (*ent.Session, error) {
	id := nanoid.Must(64)
	now := time.Now().UnixMilli()

	client := r.data.GetMasterEntClient()
	builder := client.Session.Create()
	builder.SetID(id)
	builder.SetUserID(body.UserID)
	builder.SetTokenID(tokenID)
	builder.SetIsActive(true)
	builder.SetCreatedAt(now)
	builder.SetUpdatedAt(now)
	builder.SetLastAccessAt(now)
	builder.SetExpiresAt(now + (7 * 24 * 60 * 60 * 1000)) // 7 days

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

	session, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the session
	go r.cacheSession(context.Background(), session)

	// Invalidate user sessions cache
	go r.invalidateUserSessionsCache(context.Background(), body.UserID)

	return session, nil
}

// GetByID retrieves a session by ID
func (r *sessionRepository) GetByID(ctx context.Context, id string) (*ent.Session, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("session:%s", id)
	if cached, err := r.sessionCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Fallback to database
	client := r.data.GetSlaveEntClient()
	session, err := client.Session.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache for future use
	go r.cacheSession(context.Background(), session)

	return session, nil
}

// GetByTokenID retrieves a session by token ID
func (r *sessionRepository) GetByTokenID(ctx context.Context, tokenID string) (*ent.Session, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("token:%s", tokenID)
	if cached, err := r.tokenSessionCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Fallback to database
	client := r.data.GetSlaveEntClient()
	session, err := client.Session.Query().Where(sessionEnt.TokenIDEQ(tokenID)).Only(ctx)
	if err != nil {
		return nil, err
	}

	// Cache for future use
	go r.cacheSessionByToken(context.Background(), session)

	return session, nil
}

// List retrieves sessions for user sessions
func (r *sessionRepository) List(ctx context.Context, params *structs.ListSessionParams) ([]*ent.Session, error) {
	client := r.data.GetSlaveEntClient()
	builder := client.Session.Query()

	// Apply filters
	if params.UserID != "" {
		builder = builder.Where(sessionEnt.UserIDEQ(params.UserID))
	}
	if params.IsActive != nil {
		builder = builder.Where(sessionEnt.IsActiveEQ(*params.IsActive))
	}

	// Apply cursor pagination
	if params.Cursor != "" {
		id, timestamp, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %v", err)
		}

		if params.Direction == "backward" {
			builder = builder.Where(
				sessionEnt.Or(
					sessionEnt.CreatedAtGT(timestamp),
					sessionEnt.And(
						sessionEnt.CreatedAtEQ(timestamp),
						sessionEnt.IDGT(id),
					),
				),
			)
		} else {
			builder = builder.Where(
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
		builder = builder.Order(ent.Asc(sessionEnt.FieldCreatedAt), ent.Asc(sessionEnt.FieldID))
	} else {
		builder = builder.Order(ent.Desc(sessionEnt.FieldCreatedAt), ent.Desc(sessionEnt.FieldID))
	}

	// Apply limit
	if params.Limit > 0 {
		builder = builder.Limit(params.Limit)
	}

	sessions, err := builder.All(ctx)
	if err != nil {
		return nil, err
	}

	// Cache sessions in background
	go func() {
		for _, session := range sessions {
			r.cacheSession(context.Background(), session)
		}
	}()

	return sessions, nil
}

// Delete deletes a session
func (r *sessionRepository) Delete(ctx context.Context, id string) error {
	// Get session first to know user ID for cache invalidation
	session, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	client := r.data.GetMasterEntClient()
	err = client.Session.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateSessionCache(context.Background(), id)
		r.invalidateTokenSessionCache(context.Background(), session.TokenID)
		r.invalidateUserSessionsCache(context.Background(), session.UserID)
	}()

	return nil
}

// Update updates a session
func (r *sessionRepository) Update(ctx context.Context, id string, body *structs.UpdateSessionBody) (*ent.Session, error) {
	client := r.data.GetMasterEntClient()
	builder := client.Session.UpdateOneID(id)

	if body.LastAccessAt != nil {
		builder = builder.SetLastAccessAt(*body.LastAccessAt)
	}
	if body.Location != "" {
		builder = builder.SetLocation(body.Location)
	}
	if body.IsActive != nil {
		builder = builder.SetIsActive(*body.IsActive)
	}
	if body.DeviceInfo != nil {
		builder = builder.SetDeviceInfo(*body.DeviceInfo)
	}

	builder = builder.SetUpdatedAt(time.Now().UnixMilli())

	session, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}

	// Invalidate and re-cache
	go func() {
		r.invalidateSessionCache(context.Background(), id)
		r.invalidateTokenSessionCache(context.Background(), session.TokenID)
		r.cacheSession(context.Background(), session)
		r.cacheSessionByToken(context.Background(), session)
	}()

	return session, nil
}

// DeactivateByUserID deactivates all sessions for a user
func (r *sessionRepository) DeactivateByUserID(ctx context.Context, userID string) error {
	client := r.data.GetMasterEntClient()
	_, err := client.Session.Update().
		Where(sessionEnt.UserIDEQ(userID)).
		SetIsActive(false).
		SetUpdatedAt(time.Now().UnixMilli()).
		Save(ctx)

	if err != nil {
		return err
	}

	// Invalidate user sessions cache
	go r.invalidateUserSessionsCache(context.Background(), userID)

	return nil
}

// DeactivateByTokenID deactivates a session by token ID
func (r *sessionRepository) DeactivateByTokenID(ctx context.Context, tokenID string) error {
	client := r.data.GetMasterEntClient()
	_, err := client.Session.Update().
		Where(sessionEnt.TokenIDEQ(tokenID)).
		SetIsActive(false).
		SetUpdatedAt(time.Now().UnixMilli()).
		Save(ctx)

	if err != nil {
		return err
	}

	sessions, err := client.Session.Query().
		Where(sessionEnt.TokenIDEQ(tokenID)).
		All(ctx)

	if err != nil {
		return err
	}

	// Invalidate caches
	go func() {
		if len(sessions) > 0 {
			r.invalidateSessionCache(context.Background(), sessions[0].ID)
			r.invalidateTokenSessionCache(context.Background(), tokenID)
			r.invalidateUserSessionsCache(context.Background(), sessions[0].UserID)
		}
	}()

	return nil
}

// UpdateLastAccess updates the last access time
func (r *sessionRepository) UpdateLastAccess(ctx context.Context, tokenID string) error {
	client := r.data.GetMasterEntClient()
	now := time.Now().UnixMilli()

	_, err := client.Session.Update().
		Where(sessionEnt.TokenIDEQ(tokenID)).
		SetLastAccessAt(now).
		SetUpdatedAt(now).
		Save(ctx)

	if err != nil {
		return err
	}

	sessions, err := client.Session.Query().
		Where(sessionEnt.TokenIDEQ(tokenID)).
		All(ctx)

	if err != nil {
		return err
	}

	// Update cache
	go func() {
		if len(sessions) > 0 {
			session := sessions[0]
			r.invalidateSessionCache(context.Background(), session.ID)
			r.invalidateTokenSessionCache(context.Background(), tokenID)
			// Re-cache updated session
			r.cacheSession(context.Background(), session)
			r.cacheSessionByToken(context.Background(), session)
		}
	}()

	return nil
}

// CleanupExpiredSessions removes expired sessions
func (r *sessionRepository) CleanupExpiredSessions(ctx context.Context) error {
	client := r.data.GetMasterEntClient()
	now := time.Now().UnixMilli()

	// Get expired sessions first for cache invalidation
	expiredSessions, err := client.Session.Query().
		Where(sessionEnt.ExpiresAtLT(now)).
		All(ctx)
	if err != nil {
		return err
	}

	// Delete expired sessions
	_, err = client.Session.Delete().
		Where(sessionEnt.ExpiresAtLT(now)).
		Exec(ctx)
	if err != nil {
		return err
	}

	// Invalidate caches for expired sessions
	go func() {
		for _, session := range expiredSessions {
			r.invalidateSessionCache(context.Background(), session.ID)
			r.invalidateTokenSessionCache(context.Background(), session.TokenID)
			r.invalidateUserSessionsCache(context.Background(), session.UserID)
		}
	}()

	return nil
}

// CountX counts sessions
func (r *sessionRepository) CountX(ctx context.Context, params *structs.ListSessionParams) int {
	client := r.data.GetSlaveEntClient()
	builder := client.Session.Query()

	if params.UserID != "" {
		builder = builder.Where(sessionEnt.UserIDEQ(params.UserID))
	}
	if params.IsActive != nil {
		builder = builder.Where(sessionEnt.IsActiveEQ(*params.IsActive))
	}

	return builder.CountX(ctx)
}

// cacheSession caches a session
func (r *sessionRepository) cacheSession(ctx context.Context, session *ent.Session) {
	cacheKey := fmt.Sprintf("session:%s", session.ID)
	if err := r.sessionCache.Set(ctx, cacheKey, session, r.sessionTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache session %s: %v", session.ID, err)
	}
}

// cacheSessionByToken caches a session by token
func (r *sessionRepository) cacheSessionByToken(ctx context.Context, session *ent.Session) {
	cacheKey := fmt.Sprintf("token:%s", session.TokenID)
	if err := r.tokenSessionCache.Set(ctx, cacheKey, session, r.sessionTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache session by token %s: %v", session.TokenID, err)
	}
}

// invalidateSessionCache invalidates a session cache
func (r *sessionRepository) invalidateSessionCache(ctx context.Context, sessionID string) {
	cacheKey := fmt.Sprintf("session:%s", sessionID)
	if err := r.sessionCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate session cache %s: %v", sessionID, err)
	}
}

// invalidateTokenSessionCache invalidates a token session cache
func (r *sessionRepository) invalidateTokenSessionCache(ctx context.Context, tokenID string) {
	cacheKey := fmt.Sprintf("token:%s", tokenID)
	if err := r.tokenSessionCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate token session cache %s: %v", tokenID, err)
	}
}

// invalidateUserSessionsCache invalidates user sessions cache
func (r *sessionRepository) invalidateUserSessionsCache(ctx context.Context, userID string) {
	cacheKey := fmt.Sprintf("user:%s", userID)
	if err := r.userSessionsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user sessions cache %s: %v", userID, err)
	}
}
