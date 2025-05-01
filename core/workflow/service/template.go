package service

import (
	"context"
	"errors"
	"fmt"
	"ncobase/core/workflow/data/ent"
	"ncobase/core/workflow/data/repository"
	"ncobase/core/workflow/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/jinzhu/copier"
)

type TemplateServiceInterface interface {
	Create(ctx context.Context, body *structs.TemplateBody) (*structs.ReadTemplate, error)
	Get(ctx context.Context, params *structs.FindTemplateParams) (*structs.ReadTemplate, error)
	Update(ctx context.Context, body *structs.UpdateTemplateBody) (*structs.ReadTemplate, error)
	Delete(ctx context.Context, params *structs.FindTemplateParams) error
	List(ctx context.Context, params *structs.ListTemplateParams) (paging.Result[*structs.ReadTemplate], error)

	// Template specific operations

	CreateVersion(ctx context.Context, templateID string, version string) (*structs.ReadTemplate, error)
	SetLatestVersion(ctx context.Context, templateID string) error
	Enable(ctx context.Context, templateID string) error
	Disable(ctx context.Context, templateID string) error
	ValidateTemplate(ctx context.Context, templateID string) error
}

type templateService struct {
	templateRepo repository.TemplateRepositoryInterface
	em           ext.ManagerInterface
}

func NewTemplateService(repo repository.Repository, em ext.ManagerInterface) TemplateServiceInterface {
	return &templateService{
		templateRepo: repo.GetTemplate(),
		em:           em,
	}
}

// Create creates a new template
func (s *templateService) Create(ctx context.Context, body *structs.TemplateBody) (*structs.ReadTemplate, error) {
	if body.Code == "" {
		return nil, errors.New(ecode.FieldIsRequired("code"))
	}

	// Validate template definition
	if err := s.validateTemplateDefinition(body); err != nil {
		return nil, fmt.Errorf("invalid template definition: %v", err)
	}

	// Set initial values
	if body.Version == "" {
		body.Version = "1.0.0"
	}
	body.IsLatest = true

	template, err := s.templateRepo.Create(ctx, body)
	if err := handleEntError(ctx, "Template", err); err != nil {
		return nil, err
	}

	return s.serialize(template), nil
}

// CreateVersion creates a new version of an existing template
func (s *templateService) CreateVersion(ctx context.Context, templateID string, version string) (*structs.ReadTemplate, error) {
	// Get original template
	template, err := s.templateRepo.Get(ctx, &structs.FindTemplateParams{
		Code: templateID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %v", err)
	}

	// Create new version
	newTemplate := &structs.TemplateBody{
		Name:          template.Name,
		Code:          template.Code,
		Description:   template.Description,
		Type:          template.Type,
		Version:       version,
		Status:        template.Status,
		ModuleCode:    template.ModuleCode,
		FormCode:      template.FormCode,
		NodeConfig:    template.NodeConfig,
		ProcessRules:  template.ProcessRules,
		FormConfig:    template.FormConfig,
		SourceVersion: template.Version,
	}

	result, err := s.templateRepo.Create(ctx, newTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to create version: %v", err)
	}

	// Mark old version as not latest
	if template.IsLatest {
		_, err = s.templateRepo.Update(ctx, &structs.UpdateTemplateBody{
			ID: template.ID,
			TemplateBody: structs.TemplateBody{
				IsLatest: false,
			},
		})
		if err != nil {
			logger.Errorf(ctx, "Failed to update old version latest flag: %v", err)
		}
	}

	return s.serialize(result), nil
}

// SetLatestVersion sets a template version as the latest
func (s *templateService) SetLatestVersion(ctx context.Context, templateID string) error {
	return s.templateRepo.MarkAsLatest(ctx, templateID)
}

// Enable enables a template
func (s *templateService) Enable(ctx context.Context, templateID string) error {
	template, err := s.templateRepo.Get(ctx, &structs.FindTemplateParams{
		Code: templateID,
	})
	if err != nil {
		return fmt.Errorf("failed to get template: %v", err)
	}

	_, err = s.templateRepo.Update(ctx, &structs.UpdateTemplateBody{
		ID: template.ID,
		TemplateBody: structs.TemplateBody{
			Status:   string(structs.StatusActive),
			Disabled: false,
		},
	})
	return err
}

// Disable disables a template
func (s *templateService) Disable(ctx context.Context, templateID string) error {
	return s.templateRepo.DisableTemplate(ctx, templateID)
}

// ValidateTemplate validates a template definition
func (s *templateService) ValidateTemplate(ctx context.Context, templateID string) error {
	template, err := s.templateRepo.Get(ctx, &structs.FindTemplateParams{
		Code: templateID,
	})
	if err != nil {
		return fmt.Errorf("failed to get template: %v", err)
	}

	tb := &structs.TemplateBody{}
	err = copier.CopyWithOption(&tb, &template, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	if validator.IsNotNil(err) {
		return fmt.Errorf("invalid template definition: %v", err)
	}

	return s.validateTemplateDefinition(tb)
}

// validateTemplateDefinition validates a template definition
func (s *templateService) validateTemplateDefinition(body *structs.TemplateBody) error {
	if body.Name == "" {
		return errors.New("template name is required")
	}
	if body.ModuleCode == "" {
		return errors.New("module code is required")
	}
	if body.FormCode == "" {
		return errors.New("form code is required")
	}

	// Validate node configuration
	if body.NodeConfig == nil {
		return errors.New("node configuration is required")
	}

	// Validate node relationship
	if err := s.validateNodeConnections(body.NodeConfig); err != nil {
		return fmt.Errorf("invalid node connections: %v", err)
	}

	return nil
}

// validateNodeConnections validates node connections
func (s *templateService) validateNodeConnections(nodeConfig any) error {
	// TODO: Implement node connection verification logic
	// 1. Make sure there is a start node
	// 2. Make sure there is an end node
	// 3. Verify whether the connection between nodes is valid
	// 4. Check whether there are circular dependencies
	return nil
}

// Get retrieves a template
func (s *templateService) Get(ctx context.Context, params *structs.FindTemplateParams) (*structs.ReadTemplate, error) {
	template, err := s.templateRepo.Get(ctx, params)
	if err = handleEntError(ctx, "Template", err); err != nil {
		return nil, err
	}

	return s.serialize(template), nil
}

// Update updates a template
func (s *templateService) Update(ctx context.Context, body *structs.UpdateTemplateBody) (*structs.ReadTemplate, error) {
	if body.ID == "" {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	template, err := s.templateRepo.Update(ctx, body)
	if err = handleEntError(ctx, "Template", err); err != nil {
		return nil, err
	}

	return s.serialize(template), nil
}

// Delete deletes a template
func (s *templateService) Delete(ctx context.Context, params *structs.FindTemplateParams) error {
	return handleEntError(ctx, "Template", s.templateRepo.Delete(ctx, params))
}

// List retrieves a list of templates
func (s *templateService) List(ctx context.Context, params *structs.ListTemplateParams) (paging.Result[*structs.ReadTemplate], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadTemplate, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		templates, err := s.templateRepo.List(ctx, &lp)
		if err != nil {
			logger.Errorf(ctx, "Error listing templates: %v", err)
			return nil, 0, err
		}

		total := s.templateRepo.CountX(ctx, params)

		return s.serializes(templates), total, nil
	})
}

// Serialization helpers
func (s *templateService) serialize(template *ent.Template) *structs.ReadTemplate {
	if template == nil {
		return nil
	}

	return &structs.ReadTemplate{
		ID:             template.ID,
		Name:           template.Name,
		Code:           template.Code,
		Description:    template.Description,
		Type:           template.Type,
		Version:        template.Version,
		Status:         template.Status,
		ModuleCode:     template.ModuleCode,
		FormCode:       template.FormCode,
		TemplateKey:    template.TemplateKey,
		Category:       template.Category,
		NodeConfig:     template.NodeConfig,
		NodeRules:      template.NodeRules,
		NodeEvents:     template.NodeEvents,
		ProcessRules:   template.ProcessRules,
		FormConfig:     template.FormConfig,
		FormPerms:      template.FormPermissions,
		RoleConfigs:    template.RoleConfigs,
		PermConfigs:    template.PermissionConfigs,
		VisibleRange:   template.VisibleRange,
		IsDraftEnabled: template.IsDraftEnabled,
		IsAutoStart:    template.IsAutoStart,
		StrictMode:     template.StrictMode,
		AllowCancel:    template.AllowCancel,
		AllowUrge:      template.AllowUrge,
		AllowDelegate:  template.AllowDelegate,
		AllowTransfer:  template.AllowTransfer,
		TimeoutConfig:  template.TimeoutConfig,
		ReminderConfig: template.ReminderConfig,
		SourceVersion:  template.SourceVersion,
		IsLatest:       template.IsLatest,
		Disabled:       template.Disabled,
		Extras:         template.Extras,
		TenantID:       template.TenantID,
		EffectiveTime:  &template.EffectiveTime,
		ExpireTime:     &template.ExpireTime,
		CreatedBy:      &template.CreatedBy,
		CreatedAt:      &template.CreatedAt,
		UpdatedBy:      &template.UpdatedBy,
		UpdatedAt:      &template.UpdatedAt,
	}
}

func (s *templateService) serializes(templates []*ent.Template) []*structs.ReadTemplate {
	result := make([]*structs.ReadTemplate, len(templates))
	for i, template := range templates {
		result[i] = s.serialize(template)
	}
	return result
}
