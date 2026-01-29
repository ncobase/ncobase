package repository

import (
	"context"
	"ncobase/core/auth/data"
	"ncobase/core/auth/data/ent"
)

// AuthTokenRepositoryInterface defines auth token repository operations
type AuthTokenRepositoryInterface interface {
	Create(ctx context.Context, userID string) (*ent.AuthToken, error)
}

type authTokenRepository struct {
	data *data.Data
}

// NewAuthTokenRepository creates a new auth token repository
func NewAuthTokenRepository(d *data.Data) AuthTokenRepositoryInterface {
	return &authTokenRepository{data: d}
}

// Create creates a new auth token for a user
func (r *authTokenRepository) Create(ctx context.Context, userID string) (*ent.AuthToken, error) {
	return r.data.GetMasterEntClient().AuthToken.Create().
		SetUserID(userID).
		Save(ctx)
}
