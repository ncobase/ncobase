package repository

import (
	"context"
	"ncobase/user/data"
	"ncobase/user/data/ent"
	apiKeyEnt "ncobase/user/data/ent/apikey"
	"ncobase/user/structs"
	"time"

	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"
)

// ApiKeyRepositoryInterface defines repository operations for API keys
type ApiKeyRepositoryInterface interface {
	Create(ctx context.Context, userID string, request *structs.CreateApiKeyRequest) (*ent.ApiKey, string, error)
	GetByID(ctx context.Context, id string) (*ent.ApiKey, error)
	GetByKey(ctx context.Context, key string) (*ent.ApiKey, error)
	GetByUserID(ctx context.Context, userID string) ([]*ent.ApiKey, error)
	Delete(ctx context.Context, id string) error
	UpdateLastUsed(ctx context.Context, id string, timestamp int64) error
}

// apiKeyRepository implements ApiKeyRepositoryInterface
type apiKeyRepository struct {
	ec *ent.Client
}

// NewApiKeyRepository creates a new API key repository
func NewApiKeyRepository(d *data.Data) ApiKeyRepositoryInterface {
	ec := d.GetMasterEntClient()
	return &apiKeyRepository{ec}
}

// Create adds a new API key
func (r *apiKeyRepository) Create(ctx context.Context, userID string, request *structs.CreateApiKeyRequest) (*ent.ApiKey, string, error) {
	// Generate a unique API key
	apiKeyValue := nanoid.String(64)

	// Create builder
	builder := r.ec.ApiKey.Create()

	// Set values
	builder.SetUserID(userID)
	builder.SetName(request.Name)
	builder.SetKey(apiKeyValue)

	// Execute builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "apiKeyRepo.Create error: %v", err)
		return nil, "", err
	}

	return row, apiKeyValue, nil
}

// GetByID retrieves an API key by ID
func (r *apiKeyRepository) GetByID(ctx context.Context, id string) (*ent.ApiKey, error) {
	row, err := r.ec.ApiKey.Get(ctx, id)
	if err != nil {
		logger.Errorf(ctx, "apiKeyRepo.GetByID error: %v", err)
		return nil, err
	}

	return row, nil
}

// GetByKey retrieves an API key by its key value
func (r *apiKeyRepository) GetByKey(ctx context.Context, key string) (*ent.ApiKey, error) {
	row, err := r.ec.ApiKey.
		Query().
		Where(apiKeyEnt.KeyEQ(key)).
		Only(ctx)

	if err != nil {
		logger.Errorf(ctx, "apiKeyRepo.GetByKey error: %v", err)
		return nil, err
	}

	return row, nil
}

// GetByUserID retrieves all API keys for a user
func (r *apiKeyRepository) GetByUserID(ctx context.Context, userID string) ([]*ent.ApiKey, error) {
	if validator.IsEmpty(userID) {
		logger.Errorf(ctx, "apiKeyRepo.GetByUserID: userID is empty")
		return nil, nil
	}

	rows, err := r.ec.ApiKey.
		Query().
		Where(apiKeyEnt.UserIDEQ(userID)).
		Order(ent.Desc(apiKeyEnt.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		logger.Errorf(ctx, "apiKeyRepo.GetByUserID error: %v", err)
		return nil, err
	}

	return rows, nil
}

// Delete removes an API key
func (r *apiKeyRepository) Delete(ctx context.Context, id string) error {
	err := r.ec.ApiKey.
		DeleteOneID(id).
		Exec(ctx)

	if err != nil {
		logger.Errorf(ctx, "apiKeyRepo.Delete error: %v", err)
		return err
	}

	return nil
}

// UpdateLastUsed updates the last used timestamp
func (r *apiKeyRepository) UpdateLastUsed(ctx context.Context, id string, timestamp int64) error {
	if timestamp == 0 {
		timestamp = time.Now().UnixMilli()
	}

	_, err := r.ec.ApiKey.
		UpdateOneID(id).
		SetLastUsed(timestamp).
		Save(ctx)

	if err != nil {
		logger.Errorf(ctx, "apiKeyRepo.UpdateLastUsed error: %v", err)
		return err
	}

	return nil
}
