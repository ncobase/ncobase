package service

import (
	"context"
	"errors"
	"ncobase/common/ecode"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/feature/access/data"
	"ncobase/feature/access/data/ent"
	"ncobase/feature/access/data/repository"
	"ncobase/feature/access/structs"
)

// CasbinServiceInterface is the interface for the service.
type CasbinServiceInterface interface {
	Create(ctx context.Context, body *structs.CasbinRuleBody) (*structs.ReadCasbinRule, error)
	Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadCasbinRule, error)
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*structs.ReadCasbinRule, error)
	List(ctx context.Context, params *structs.ListCasbinRuleParams) ([]*structs.ReadCasbinRule, error)
	CountX(ctx context.Context, params *structs.ListCasbinRuleParams) int
}

// casbinService is the struct for the service.
type casbinService struct {
	casbin repository.CasbinRuleRepositoryInterface
}

// NewCasbinService creates a new service.
func NewCasbinService(d *data.Data) CasbinServiceInterface {
	return &casbinService{
		casbin: repository.NewCasbinRule(d),
	}
}

// Create creates a new Casbin rule.
func (s *casbinService) Create(ctx context.Context, body *structs.CasbinRuleBody) (*structs.ReadCasbinRule, error) {
	row, err := s.casbin.Create(ctx, body)
	if err := handleEntError("Casbin", err); err != nil {
		return nil, err
	}

	return s.SerializeCasbin(row), nil
}

// Update updates an existing Casbin rule (full and partial).
func (s *casbinService) Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadCasbinRule, error) {
	if validator.IsEmpty(id) {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	// Validate the updates map
	if len(updates) == 0 {
		return nil, errors.New(ecode.FieldIsEmpty("updates fields"))
	}

	row, err := s.casbin.Update(ctx, id, updates)
	if err := handleEntError("Casbin", err); err != nil {
		return nil, err
	}

	return s.SerializeCasbin(row), nil
}

// Get retrieves a Casbin rule by ID.
func (s *casbinService) Get(ctx context.Context, id string) (*structs.ReadCasbinRule, error) {
	row, err := s.casbin.GetByID(ctx, id)
	if err := handleEntError("Casbin", err); err != nil {
		return nil, err
	}

	return s.SerializeCasbin(row), nil
}

// Delete deletes a Casbin rule by ID.
func (s *casbinService) Delete(ctx context.Context, id string) error {
	err := s.casbin.Delete(ctx, id)
	if err := handleEntError("Casbin", err); err != nil {
		return err
	}

	return nil
}

// List lists all Casbin rules based on query parameters.
func (s *casbinService) List(ctx context.Context, params *structs.ListCasbinRuleParams) ([]*structs.ReadCasbinRule, error) {
	rows, err := s.casbin.Find(ctx, params)
	if err := handleEntError("Casbin", err); err != nil {
		return nil, err
	}

	return s.SerializeCasbins(rows), nil
}

// CountX gets a count of Casbin rules.
func (s *casbinService) CountX(ctx context.Context, params *structs.ListCasbinRuleParams) int {
	return s.casbin.CountX(ctx, params)
}

// SerializeCasbins serializes a list of Casbin rule entities to a response format.
func (s *casbinService) SerializeCasbins(rows []*ent.CasbinRule) []*structs.ReadCasbinRule {
	var rs []*structs.ReadCasbinRule
	for _, row := range rows {
		rs = append(rs, s.SerializeCasbin(row))
	}
	return rs
}

// SerializeCasbin serializes a Casbin rule entity to a response format.
func (s *casbinService) SerializeCasbin(row *ent.CasbinRule) *structs.ReadCasbinRule {
	return &structs.ReadCasbinRule{
		PType:     row.PType,
		V0:        row.V0,
		V1:        row.V1,
		V2:        row.V2,
		V3:        &row.V3,
		V4:        &row.V4,
		V5:        &row.V5,
		CreatedBy: &row.CreatedBy,
		CreatedAt: &row.CreatedAt,
		UpdatedBy: &row.UpdatedBy,
		UpdatedAt: &row.UpdatedAt,
	}
}
