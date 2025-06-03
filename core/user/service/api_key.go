package service

import (
	"context"
	"errors"
	"ncobase/user/data/ent"
	"ncobase/user/data/repository"
	"ncobase/user/event"
	"ncobase/user/structs"
	"time"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
)

// ApiKeyServiceInterface defines service operations for API keys
type ApiKeyServiceInterface interface {
	GenerateApiKey(ctx context.Context, userID string, request *structs.CreateApiKeyRequest) (*structs.ApiKey, error)
	GetApiKey(ctx context.Context, id string) (*structs.ApiKey, error)
	GetUserApiKeys(ctx context.Context, userID string) ([]*structs.ApiKey, error)
	DeleteApiKey(ctx context.Context, id string) error
	ValidateApiKey(ctx context.Context, key string) (*structs.ApiKey, error)
}

// apiKeyService implements ApiKeyServiceInterface
type apiKeyService struct {
	apiKey repository.ApiKeyRepositoryInterface
	ep     event.PublisherInterface
}

// NewApiKeyService creates a new API key service
func NewApiKeyService(repo *repository.Repository, ep event.PublisherInterface) ApiKeyServiceInterface {
	return &apiKeyService{
		apiKey: repo.ApiKey,
		ep:     ep,
	}
}

// GenerateApiKey creates a new API key
func (s *apiKeyService) GenerateApiKey(ctx context.Context, userID string, request *structs.CreateApiKeyRequest) (*structs.ApiKey, error) {
	if request.Name == "" {
		return nil, errors.New("API key name is required")
	}

	// currentUserID := ctxutil.GetUserID(ctx)
	// if currentUserID != userID && !ctxutil.IsAdmin(ctx) {
	// 	return nil, errors.New("unauthorized access to create API key")
	// }

	// Create API key
	row, apiKeyValue, err := s.apiKey.Create(ctx, userID, request)
	if err != nil {
		logger.Errorf(ctx, "apiKeyService.GenerateApiKey error: %v", err)
		return nil, err
	}

	// Use apiKeyValue from create operation
	apiKey := s.Serialize(row)
	apiKey.Key = apiKeyValue // Set the clear text key value (only returned once)

	// Publish API key generated event
	if s.ep != nil {
		s.ep.PublishApiKeyGenerated(ctx, userID, &types.JSON{
			"key_name": request.Name,
			"key_id":   apiKey.ID,
		})
	}

	return apiKey, nil
}

// GetApiKey retrieves an API key by ID
func (s *apiKeyService) GetApiKey(ctx context.Context, id string) (*structs.ApiKey, error) {
	row, err := s.apiKey.GetByID(ctx, id)
	if err := handleEntError(ctx, "ApiKey", err); err != nil {
		return nil, err
	}

	// Check if current user is authorized to access this API key
	currentUserID := ctxutil.GetUserID(ctx)
	isAdmin := ctxutil.GetUserIsAdmin(ctx)
	if !isAdmin && currentUserID != row.UserID {
		return nil, errors.New("unauthorized access to API key")
	}

	return s.Serialize(row), nil
}

// GetUserApiKeys retrieves all API keys for a user
func (s *apiKeyService) GetUserApiKeys(ctx context.Context, userID string) ([]*structs.ApiKey, error) {
	// Check if current user is authorized to access these API keys
	currentUserID := ctxutil.GetUserID(ctx)
	isAdmin := ctxutil.GetUserIsAdmin(ctx)
	if !isAdmin && currentUserID != userID {
		return nil, errors.New("unauthorized access to API keys")
	}

	rows, err := s.apiKey.GetByUserID(ctx, userID)
	if err != nil {
		logger.Errorf(ctx, "apiKeyService.GetUserApiKeys error: %v", err)
		return nil, err
	}

	return s.Serializes(rows), nil
}

// DeleteApiKey removes an API key
func (s *apiKeyService) DeleteApiKey(ctx context.Context, id string) error {
	// Get API key first to check ownership
	row, err := s.apiKey.GetByID(ctx, id)
	if err := handleEntError(ctx, "ApiKey", err); err != nil {
		return err
	}

	// Check if current user is authorized to delete this API key
	currentUserID := ctxutil.GetUserID(ctx)
	isAdmin := ctxutil.GetUserIsAdmin(ctx)
	if !isAdmin && currentUserID != row.UserID {
		return errors.New("unauthorized access to delete API key")
	}

	// Delete API key
	err = s.apiKey.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Publish API key deleted event
	if s.ep != nil {
		s.ep.PublishApiKeyDeleted(ctx, row.UserID, &types.JSON{
			"key_id":   id,
			"key_name": row.Name,
		})
	}

	return nil
}

// ValidateApiKey validates an API key and updates last used timestamp
func (s *apiKeyService) ValidateApiKey(ctx context.Context, key string) (*structs.ApiKey, error) {
	row, err := s.apiKey.GetByKey(ctx, key)
	if err != nil {
		return nil, errors.New("invalid API key")
	}

	// Update last used timestamp
	now := time.Now().UnixMilli()
	_ = s.apiKey.UpdateLastUsed(ctx, row.ID, now)

	return s.Serialize(row), nil
}

// Serializes converts multiple API keys to DTOs
func (s *apiKeyService) Serializes(rows []*ent.ApiKey) []*structs.ApiKey {
	rs := make([]*structs.ApiKey, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize converts an API key to DTO
func (s *apiKeyService) Serialize(apiKey *ent.ApiKey) *structs.ApiKey {
	return &structs.ApiKey{
		ID:        apiKey.ID,
		Name:      apiKey.Name,
		Key:       "", // Don't return the key hash for security
		UserID:    apiKey.UserID,
		CreatedAt: apiKey.CreatedAt,
		LastUsed:  &apiKey.LastUsed, // This is a pointer field in the model
	}
}
