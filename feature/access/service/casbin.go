package service

import (
	"context"
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/feature/access/data"
	"ncobase/feature/access/data/repository"
	"ncobase/feature/access/structs"
)

// CasbinServiceInterface is the interface for the service.
type CasbinServiceInterface interface {
	Create(ctx context.Context, body *structs.CasbinRuleBody) (*resp.Exception, error)
	Update(ctx context.Context, id string, updates types.JSON) (*resp.Exception, error)
	Delete(ctx context.Context, id string) (*resp.Exception, error)
	Get(ctx context.Context, id string) (*resp.Exception, error)
	List(ctx context.Context, params *structs.ListCasbinRuleParams) (*resp.Exception, error)
}

// casbinService is the struct for the service.
type casbinService struct {
	casbinRule repository.CasbinRuleRepositoryInterface
}

// NewCasbinService creates a new service.
func NewCasbinService(d *data.Data) CasbinServiceInterface {
	return &casbinService{
		casbinRule: repository.NewCasbinRule(d),
	}
}

// Create creates a new Casbin rule.
func (s *casbinService) Create(ctx context.Context, body *structs.CasbinRuleBody) (*resp.Exception, error) {
	// Create a new CasbinRule entity
	casbinRule, err := s.casbinRule.Create(ctx, body)
	if exception, err := handleEntError("CasbinRule", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: casbinRule,
	}, nil
}

// Update updates an existing Casbin rule (full and partial).
func (s *casbinService) Update(ctx context.Context, id string, updates types.JSON) (*resp.Exception, error) {
	if validator.IsEmpty(id) {
		return resp.BadRequest(ecode.FieldIsRequired("id")), nil
	}

	// Validate the updates map
	if len(updates) == 0 {
		return resp.BadRequest(ecode.FieldIsEmpty("updates fields")), nil
	}

	casbinRule, err := s.casbinRule.Update(ctx, id, updates)
	if exception, err := handleEntError("CasbinRule", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: casbinRule,
	}, nil
}

// Get retrieves a Casbin rule by ID.
func (s *casbinService) Get(ctx context.Context, id string) (*resp.Exception, error) {
	casbinRule, err := s.casbinRule.GetByID(ctx, id)
	if exception, err := handleEntError("CasbinRule", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: casbinRule,
	}, nil
}

// Delete deletes a Casbin rule by ID.
func (s *casbinService) Delete(ctx context.Context, id string) (*resp.Exception, error) {
	err := s.casbinRule.Delete(ctx, id)
	if exception, err := handleEntError("CasbinRule", err); exception != nil {
		return exception, err
	}

	return nil, nil
}

// List lists all Casbin rules based on query parameters.
func (s *casbinService) List(ctx context.Context, params *structs.ListCasbinRuleParams) (*resp.Exception, error) {
	casbinRules, err := s.casbinRule.Find(ctx, params)
	if exception, err := handleEntError("CasbinRule", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: casbinRules,
	}, nil
}
