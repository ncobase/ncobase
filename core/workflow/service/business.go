package service

import (
	"context"
	"errors"
	"ncobase/core/workflow/data/ent"
	"ncobase/core/workflow/data/repository"
	"ncobase/core/workflow/structs"
	"ncore/extension"
	"ncore/pkg/ecode"
	"ncore/pkg/logger"
	"ncore/pkg/paging"
	"ncore/pkg/types"
)

type BusinessServiceInterface interface {
	// Basic operations

	Create(ctx context.Context, body *structs.BusinessBody) (*structs.ReadBusiness, error)
	Get(ctx context.Context, params *structs.FindBusinessParams) (*structs.ReadBusiness, error)
	Update(ctx context.Context, body *structs.UpdateBusinessBody) (*structs.ReadBusiness, error)
	Delete(ctx context.Context, params *structs.FindBusinessParams) error
	List(ctx context.Context, params *structs.ListBusinessParams) (paging.Result[*structs.ReadBusiness], error)

	// Business specific operations

	UpdateFlowStatus(ctx context.Context, businessID string, status string) error
	UpdateBusinessData(ctx context.Context, businessID string, data types.JSON) error
	SaveDraft(ctx context.Context, businessID string, data types.JSON) error
}

type businessService struct {
	businessRepo repository.BusinessRepositoryInterface
	em           *extension.Manager
}

func NewBusinessService(repo repository.Repository, em *extension.Manager) BusinessServiceInterface {
	return &businessService{
		businessRepo: repo.GetBusiness(),
		em:           em,
	}
}

// Create creates a new business record
func (s *businessService) Create(ctx context.Context, body *structs.BusinessBody) (*structs.ReadBusiness, error) {
	if body.ModuleCode == "" {
		return nil, errors.New(ecode.FieldIsRequired("module_code"))
	}
	if body.FormCode == "" {
		return nil, errors.New(ecode.FieldIsRequired("form_code"))
	}

	// Save original data
	body.OriginData = body.CurrentData

	business, err := s.businessRepo.Create(ctx, body)
	if err := handleEntError(ctx, "Business", err); err != nil {
		return nil, err
	}

	return s.serialize(business), nil
}

// UpdateFlowStatus updates the flow status
func (s *businessService) UpdateFlowStatus(ctx context.Context, businessID string, status string) error {
	return s.businessRepo.UpdateFlowStatus(ctx, businessID, status)
}

// UpdateBusinessData updates the business data
func (s *businessService) UpdateBusinessData(ctx context.Context, businessID string, data types.JSON) error {
	return s.businessRepo.UpdateBusinessData(ctx, businessID, data)
}

// SaveDraft saves the business data as draft
func (s *businessService) SaveDraft(ctx context.Context, businessID string, data types.JSON) error {
	_, err := s.businessRepo.Update(ctx, &structs.UpdateBusinessBody{
		ID: businessID,
		BusinessBody: structs.BusinessBody{
			CurrentData: data,
			IsDraft:     true,
		},
	})
	return err
}

// Get retrieves a business record
func (s *businessService) Get(ctx context.Context, params *structs.FindBusinessParams) (*structs.ReadBusiness, error) {
	business, err := s.businessRepo.Get(ctx, params)
	if err := handleEntError(ctx, "Business", err); err != nil {
		return nil, err
	}

	return s.serialize(business), nil
}

// Update updates a business record
func (s *businessService) Update(ctx context.Context, body *structs.UpdateBusinessBody) (*structs.ReadBusiness, error) {
	if body.ID == "" {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	business, err := s.businessRepo.Update(ctx, body)
	if err := handleEntError(ctx, "Business", err); err != nil {
		return nil, err
	}

	return s.serialize(business), nil
}

// Delete deletes a business record
func (s *businessService) Delete(ctx context.Context, params *structs.FindBusinessParams) error {
	return handleEntError(ctx, "Business", s.businessRepo.Delete(ctx, params))
}

// List returns a list of business records
func (s *businessService) List(ctx context.Context, params *structs.ListBusinessParams) (paging.Result[*structs.ReadBusiness], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadBusiness, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		businesses, err := s.businessRepo.List(ctx, &lp)
		if err != nil {
			logger.Errorf(ctx, "Error listing businesses: %v", err)
			return nil, 0, err
		}

		total := s.businessRepo.CountX(ctx, params)

		return s.serializes(businesses), total, nil
	})
}

// Serialization helpers
func (s *businessService) serialize(business *ent.Business) *structs.ReadBusiness {
	if business == nil {
		return nil
	}

	return &structs.ReadBusiness{
		ID:           business.ID,
		Code:         business.Code,
		Status:       business.Status,
		ModuleCode:   business.ModuleCode,
		FormCode:     business.FormCode,
		FormVersion:  business.FormVersion,
		ProcessID:    business.ProcessID,
		TemplateID:   business.TemplateID,
		FlowStatus:   business.FlowStatus,
		OriginData:   business.OriginData,
		CurrentData:  business.CurrentData,
		Variables:    business.FlowVariables,
		IsDraft:      business.IsDraft,
		BusinessTags: business.BusinessTags,
		Viewers:      business.Viewers,
		Editors:      business.Editors,
		TenantID:     business.TenantID,
		Extras:       business.Extras,
		CreatedBy:    &business.CreatedBy,
		CreatedAt:    &business.CreatedAt,
		UpdatedBy:    &business.UpdatedBy,
		UpdatedAt:    &business.UpdatedAt,
		LastModified: &business.LastModified,
		LastModifier: &business.LastModifier,
	}
}

func (s *businessService) serializes(businesses []*ent.Business) []*structs.ReadBusiness {
	result := make([]*structs.ReadBusiness, len(businesses))
	for i, business := range businesses {
		result[i] = s.serialize(business)
	}
	return result
}
