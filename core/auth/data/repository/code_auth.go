package repository

import (
	"context"
	"ncobase/core/auth/data"
	"ncobase/core/auth/data/ent"
	codeAuthEnt "ncobase/core/auth/data/ent/codeauth"
)

// CodeAuthRepositoryInterface defines code auth repository operations
type CodeAuthRepositoryInterface interface {
	Create(ctx context.Context, email, code string) (*ent.CodeAuth, error)
	FindByCode(ctx context.Context, code string) (*ent.CodeAuth, error)
	MarkAsUsed(ctx context.Context, id string) error
}

// codeAuthRepository implements CodeAuthRepositoryInterface
type codeAuthRepository struct {
	data *data.Data
}

// NewCodeAuthRepository creates a new code auth repository
func NewCodeAuthRepository(d *data.Data) CodeAuthRepositoryInterface {
	return &codeAuthRepository{data: d}
}

// Create creates a new code auth record within a transaction
func (r *codeAuthRepository) Create(ctx context.Context, email, code string) (*ent.CodeAuth, error) {
	tx, err := r.data.GetMasterEntClient().Tx(ctx)
	if err != nil {
		return nil, err
	}

	record, err := tx.CodeAuth.Create().
		SetEmail(email).
		SetCode(code).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return record, nil
}

// FindByCode finds a code auth record by code
func (r *codeAuthRepository) FindByCode(ctx context.Context, code string) (*ent.CodeAuth, error) {
	return r.data.GetMasterEntClient().CodeAuth.Query().
		Where(codeAuthEnt.CodeEQ(code)).
		Only(ctx)
}

// MarkAsUsed marks a code auth record as used
func (r *codeAuthRepository) MarkAsUsed(ctx context.Context, id string) error {
	_, err := r.data.GetMasterEntClient().CodeAuth.UpdateOneID(id).
		SetLogged(true).
		Save(ctx)
	return err
}
