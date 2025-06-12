package service

import (
	"context"
	"errors"
	"ncobase/workflow/data/ent"
	"ncobase/workflow/data/repository"
	"ncobase/workflow/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
)

type ProcessDesignServiceInterface interface {
	Create(ctx context.Context, body *structs.ProcessDesignBody) (*structs.ReadProcessDesign, error)
	Get(ctx context.Context, params *structs.FindProcessDesignParams) (*structs.ReadProcessDesign, error)
	Update(ctx context.Context, body *structs.UpdateProcessDesignBody) (*structs.ReadProcessDesign, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params *structs.ListProcessDesignParams) (paging.Result[*structs.ReadProcessDesign], error)

	// Process design specific operations

	SaveDraft(ctx context.Context, id string, design types.JSON) error
	PublishDraft(ctx context.Context, id string) error
	ValidateDesign(ctx context.Context, design types.JSON) error
	ImportDesign(ctx context.Context, templateID string, design types.JSON) error
	ExportDesign(ctx context.Context, id string) (types.JSON, error)
}

type processDesignService struct {
	processDesignRepo repository.ProcessDesignRepositoryInterface
	templateRepo      repository.TemplateRepositoryInterface
	em                ext.ManagerInterface
}

func NewProcessDesignService(repo repository.Repository, em ext.ManagerInterface) ProcessDesignServiceInterface {
	return &processDesignService{
		processDesignRepo: repo.GetProcessDesign(),
		templateRepo:      repo.GetTemplate(),
		em:                em,
	}
}

// Create creates a new process design
func (s *processDesignService) Create(ctx context.Context, body *structs.ProcessDesignBody) (*structs.ReadProcessDesign, error) {
	if body.TemplateID == "" {
		return nil, errors.New(ecode.FieldIsRequired("template_id"))
	}

	// Validate process template exists
	_, err := s.templateRepo.Get(ctx, &structs.FindTemplateParams{
		Code: body.TemplateID,
	})
	if err := handleEntError(ctx, "Template", err); err != nil {
		return nil, err
	}

	// Validate design
	if body.GraphData != nil {
		if err := s.ValidateDesign(ctx, body.GraphData); err != nil {
			return nil, err
		}
	}

	design, err := s.processDesignRepo.Create(ctx, body)
	if err := handleEntError(ctx, "ProcessDesign", err); err != nil {
		return nil, err
	}

	return s.serialize(design), nil
}

// Get retrieves a specific process design
func (s *processDesignService) Get(ctx context.Context, params *structs.FindProcessDesignParams) (*structs.ReadProcessDesign, error) {
	design, err := s.processDesignRepo.Get(ctx, params)
	if err := handleEntError(ctx, "ProcessDesign", err); err != nil {
		return nil, err
	}

	return s.serialize(design), nil
}

// Update updates an existing process design
func (s *processDesignService) Update(ctx context.Context, body *structs.UpdateProcessDesignBody) (*structs.ReadProcessDesign, error) {
	if body.ID == "" {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	// Validate design if provided
	if body.GraphData != nil {
		if err := s.ValidateDesign(ctx, body.GraphData); err != nil {
			return nil, err
		}
	}

	design, err := s.processDesignRepo.Update(ctx, body)
	if err := handleEntError(ctx, "ProcessDesign", err); err != nil {
		return nil, err
	}

	return s.serialize(design), nil
}

// Delete deletes a process design
func (s *processDesignService) Delete(ctx context.Context, id string) error {
	return s.processDesignRepo.Delete(ctx, id)
}

// List returns a list of process designs
func (s *processDesignService) List(ctx context.Context, params *structs.ListProcessDesignParams) (paging.Result[*structs.ReadProcessDesign], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadProcessDesign, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		designs, err := s.processDesignRepo.List(ctx, &lp)
		if err != nil {
			logger.Errorf(ctx, "Error listing process designs: %v", err)
			return nil, 0, err
		}

		return s.serializes(designs), len(designs), nil
	})
}

// SaveDraft saves process design as draft
func (s *processDesignService) SaveDraft(ctx context.Context, id string, design types.JSON) error {
	// Validate design
	if err := s.ValidateDesign(ctx, design); err != nil {
		return err
	}

	return s.processDesignRepo.SaveDraft(ctx, id, design)
}

// PublishDraft publishes a draft process design
func (s *processDesignService) PublishDraft(ctx context.Context, id string) error {
	// Get draft design
	design, err := s.processDesignRepo.Get(ctx, &structs.FindProcessDesignParams{
		TemplateID: id,
	})
	if err != nil {
		return err
	}

	// Validate before publishing
	if err := s.ValidateDesign(ctx, design.GraphData); err != nil {
		return err
	}

	return s.processDesignRepo.PublishDraft(ctx, id)
}

// ValidateDesign validates process design
func (s *processDesignService) ValidateDesign(ctx context.Context, design types.JSON) error {
	// TODO: Implement process design validation logic
	// 1. Validate graph structure
	// 2. Validate node configurations
	// 3. Validate node connections
	// 4. Validate business rules
	return nil
}

// ImportDesign imports process design
func (s *processDesignService) ImportDesign(ctx context.Context, templateID string, design types.JSON) error {
	// Validate design
	if err := s.ValidateDesign(ctx, design); err != nil {
		return err
	}

	// Create new design
	_, err := s.Create(ctx, &structs.ProcessDesignBody{
		TemplateID: templateID,
		GraphData:  design,
		IsDraft:    true,
	})

	return err
}

// ExportDesign exports process design
func (s *processDesignService) ExportDesign(ctx context.Context, id string) (types.JSON, error) {
	design, err := s.Get(ctx, &structs.FindProcessDesignParams{
		TemplateID: id,
	})
	if err != nil {
		return nil, err
	}

	return design.GraphData, nil
}

// Serialization helpers
func (s *processDesignService) serialize(design *ent.ProcessDesign) *structs.ReadProcessDesign {
	if design == nil {
		return nil
	}

	return &structs.ReadProcessDesign{
		ID:              design.ID,
		TemplateID:      design.TemplateID,
		GraphData:       design.GraphData,
		NodeLayouts:     design.NodeLayouts,
		Properties:      design.Properties,
		ValidationRules: design.ValidationRules,
		IsDraft:         design.IsDraft,
		Version:         design.Version,
		SourceVersion:   design.SourceVersion,
		SpaceID:         design.SpaceID,
		Extras:          design.Extras,
		CreatedBy:       &design.CreatedBy,
		CreatedAt:       &design.CreatedAt,
		UpdatedBy:       &design.UpdatedBy,
		UpdatedAt:       &design.UpdatedAt,
	}
}

func (s *processDesignService) serializes(designs []*ent.ProcessDesign) []*structs.ReadProcessDesign {
	result := make([]*structs.ReadProcessDesign, len(designs))
	for i, design := range designs {
		result[i] = s.serialize(design)
	}
	return result
}
